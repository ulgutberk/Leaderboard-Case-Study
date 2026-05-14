package repositories

import (
"context"
"fmt"
"sort"
"strconv"

"github.com/go-redis/redis/v8"
"github.com/jackc/pgx/v5/pgxpool"

"leaderboard-case-study/internal/models"
)

// ScoreRepository handles score operations.
//
// Write path: Postgres (source of truth) + Redis ZSET/HASH (cache for fast reads).
// Read path:  Redis only.
// Crash recovery: WarmCache() rebuilds Redis from Postgres at startup.
//
// Data structures per board in Redis:
//   - ZSET  leaderboard:board:{id}     → member=userID, score=actual_score
//   - HASH  leaderboard:board:{id}:ts  → field=userID, value=unix_nano (first submission time)
//
// Tie-breaking: equal scores → earlier Postgres created_at ranks higher.
type ScoreRepository interface {
SetScore(ctx context.Context, boardID int, userID string, score float64) error
GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error)
GetSurroundings(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error)
ResetScores(ctx context.Context, boardID int) error
WarmCache(ctx context.Context) error
}

type scoreRepository struct {
db    *pgxpool.Pool
redis *redis.Client
}

func NewScoreRepository(db *pgxpool.Pool, redis *redis.Client) ScoreRepository {
return &scoreRepository{db: db, redis: redis}
}

func (r *scoreRepository) zsetKey(boardID int) string {
return fmt.Sprintf("leaderboard:board:%d", boardID)
}

func (r *scoreRepository) tsKey(boardID int) string {
return fmt.Sprintf("leaderboard:board:%d:ts", boardID)
}

func (r *scoreRepository) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
var tsNano int64
err := r.db.QueryRow(ctx, `
INSERT INTO scores (board_id, user_id, score, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (board_id, user_id) DO UPDATE
SET score = EXCLUDED.score, updated_at = NOW()
RETURNING EXTRACT(EPOCH FROM created_at)::BIGINT * 1000000000`,
boardID, userID, score,
).Scan(&tsNano)
if err != nil {
return fmt.Errorf("postgres upsert score: %w", err)
}

pipe := r.redis.Pipeline()
pipe.ZAdd(ctx, r.zsetKey(boardID), &redis.Z{Score: score, Member: userID})
pipe.HSet(ctx, r.tsKey(boardID), userID, tsNano)
if _, err := pipe.Exec(ctx); err != nil {
return fmt.Errorf("redis mirror score: %w", err)
}
return nil
}

type scoredEntry struct {
UserID string
Score  float64
Ts     int64 
}

func (r *scoreRepository) fetchTimestamps(ctx context.Context, boardID int, userIDs []string) (map[string]int64, error) {
if len(userIDs) == 0 {
return map[string]int64{}, nil
}
vals, err := r.redis.HMGet(ctx, r.tsKey(boardID), userIDs...).Result()
if err != nil {
return nil, err
}
ts := make(map[string]int64, len(userIDs))
for i, v := range vals {
if v != nil {
n, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
ts[userIDs[i]] = n
}
}
return ts, nil
}

func sortEntries(entries []scoredEntry) {
sort.SliceStable(entries, func(i, j int) bool {
if entries[i].Score != entries[j].Score {
return entries[i].Score > entries[j].Score
}
return entries[i].Ts < entries[j].Ts
})
}

func (r *scoreRepository) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error) {
results, err := r.redis.ZRevRangeWithScores(ctx, r.zsetKey(boardID), 0, limit-1).Result()
if err != nil {
return nil, err
}
if len(results) == 0 {
return []models.ScoreEntry{}, nil
}

userIDs := make([]string, len(results))
entries := make([]scoredEntry, len(results))
for i, z := range results {
uid := z.Member.(string)
userIDs[i] = uid
entries[i] = scoredEntry{UserID: uid, Score: z.Score}
}

tsMap, err := r.fetchTimestamps(ctx, boardID, userIDs)
if err != nil {
return nil, err
}
for i := range entries {
entries[i].Ts = tsMap[entries[i].UserID]
}
sortEntries(entries)

out := make([]models.ScoreEntry, len(entries))
for i, e := range entries {
out[i] = models.ScoreEntry{UserID: e.UserID, Score: e.Score}
}
return out, nil
}

func (r *scoreRepository) GetSurroundings(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error) {
rank, err := r.redis.ZRevRank(ctx, r.zsetKey(boardID), userID).Result()
if err == redis.Nil {
return nil, nil
}
if err != nil {
return nil, err
}

buffer := n * 3
start := rank - buffer
if start < 0 {
start = 0
}
end := rank + buffer

results, err := r.redis.ZRevRangeWithScores(ctx, r.zsetKey(boardID), start, end).Result()
if err != nil {
return nil, err
}

userIDs := make([]string, len(results))
entries := make([]scoredEntry, len(results))
for i, z := range results {
uid := z.Member.(string)
userIDs[i] = uid
entries[i] = scoredEntry{UserID: uid, Score: z.Score}
}

tsMap, err := r.fetchTimestamps(ctx, boardID, userIDs)
if err != nil {
return nil, err
}
for i := range entries {
entries[i].Ts = tsMap[entries[i].UserID]
}
sortEntries(entries)

userIdx := -1
for i, e := range entries {
if e.UserID == userID {
userIdx = i
break
}
}
if userIdx == -1 {
return nil, nil
}

aboveStart := userIdx - int(n)
if aboveStart < 0 {
aboveStart = 0
}
above := make([]models.ScoreEntry, 0, n)
for i := aboveStart; i < userIdx; i++ {
above = append(above, models.ScoreEntry{UserID: entries[i].UserID, Score: entries[i].Score})
}

belowEnd := userIdx + int(n) + 1
if belowEnd > len(entries) {
belowEnd = len(entries)
}
below := make([]models.ScoreEntry, 0, n)
for i := userIdx + 1; i < belowEnd; i++ {
below = append(below, models.ScoreEntry{UserID: entries[i].UserID, Score: entries[i].Score})
}

return &models.SurroundingsResponse{
User:  models.ScoreEntry{UserID: userID, Score: entries[userIdx].Score},
Above: above,
Below: below,
}, nil
}

func (r *scoreRepository) ResetScores(ctx context.Context, boardID int) error {
if _, err := r.db.Exec(ctx, `DELETE FROM scores WHERE board_id = $1`, boardID); err != nil {
return fmt.Errorf("postgres delete scores: %w", err)
}
pipe := r.redis.Pipeline()
pipe.Del(ctx, r.zsetKey(boardID))
pipe.Del(ctx, r.tsKey(boardID))
if _, err := pipe.Exec(ctx); err != nil {
return fmt.Errorf("redis delete scores: %w", err)
}
return nil
}

func (r *scoreRepository) WarmCache(ctx context.Context) error {
rows, err := r.db.Query(ctx, `
SELECT board_id, user_id, score,
       EXTRACT(EPOCH FROM created_at)::BIGINT * 1000000000 AS ts_nano
FROM scores`)
if err != nil {
return fmt.Errorf("warm cache query: %w", err)
}
defer rows.Close()

type scoreRow struct {
boardID int
userID  string
score   float64
tsNano  int64
}

byBoard := map[int][]scoreRow{}
for rows.Next() {
var rw scoreRow
if err := rows.Scan(&rw.boardID, &rw.userID, &rw.score, &rw.tsNano); err != nil {
return err
}
byBoard[rw.boardID] = append(byBoard[rw.boardID], rw)
}
if err := rows.Err(); err != nil {
return err
}

for boardID, scores := range byBoard {
pipe := r.redis.Pipeline()
pipe.Del(ctx, r.zsetKey(boardID))
pipe.Del(ctx, r.tsKey(boardID))
for _, s := range scores {
pipe.ZAdd(ctx, r.zsetKey(boardID), &redis.Z{Score: s.score, Member: s.userID})
pipe.HSet(ctx, r.tsKey(boardID), s.userID, s.tsNano)
}
if _, err := pipe.Exec(ctx); err != nil {
return fmt.Errorf("warm cache board %d: %w", boardID, err)
}
}
return nil
}
