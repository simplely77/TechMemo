package dao

import (
	"context"
	"techmemo/backend/model"
	"techmemo/backend/query"
	"time"

	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SearchDao struct {
	q  *query.Query
	db *gorm.DB
}

// SearchResult 向量搜索结果
type SearchResult struct {
	TargetID   int64
	TargetType string
	Distance   float64
}

// SearchEmbeddingsByVector 使用向量搜索 embedding 表
// 使用余弦距离 <=> 操作符，返回最相似的 topK 个结果
func (d *SearchDao) SearchEmbeddingsByVector(
	ctx context.Context,
	vector []float32,
	targetType string,
	userID int64,
	topK int,
	threshold float32,
) ([]SearchResult, error) {
	var results []SearchResult

	// 构建 SQL 查询
	// 使用 pgvector 的余弦距离操作符 <=>
	// 需要 JOIN 对应的表来过滤 user_id
	var querySQL string
	switch targetType {
	case "note":
		querySQL = `
			SELECT e.target_id, e.target_type, (e.vector <=> $1) as distance
			FROM embedding e
			INNER JOIN note n ON e.target_id = n.id
			WHERE e.target_type = 'note'
			  AND n.user_id = $2
			  AND n.status != 'deleted'
			  AND (e.vector <=> $1) < $4
			ORDER BY distance
			LIMIT $3
		`
	case "knowledge":
		querySQL = `
			SELECT e.target_id, e.target_type, (e.vector <=> $1) as distance
			FROM embedding e
			INNER JOIN knowledge_point kp ON e.target_id = kp.id
			WHERE e.target_type = 'knowledge'
			  AND kp.user_id = $2
			  AND (e.vector <=> $1) < $4
			ORDER BY distance
			LIMIT $3
		`
	default:
		return nil, nil
	}

	err := d.db.WithContext(ctx).Raw(
		querySQL,
		pgvector.NewVector(vector),
		userID,
		topK,
		threshold,
	).Scan(&results).Error

	return results, err
}

func NewSearchDao(q *query.Query, db *gorm.DB) *SearchDao {
	return &SearchDao{q: q, db: db}
}

func (d *SearchDao) SaveSearchHistory(ctx context.Context, searchHistory *model.SearchHistory) error {
	return d.q.SearchHistory.
		WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "keyword"},
				{Name: "search_type"},
				{Name: "target_type"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"last_searched_at": time.Now(),
			}),
		}).Create(searchHistory)
}

// ListSearchHistory 按最近搜索时间倒序分页返回当前用户的搜索历史
func (d *SearchDao) ListSearchHistory(ctx context.Context, userID int64, offset, limit int, searchType, targetType string) ([]*model.SearchHistory, int64, error) {
	return d.q.SearchHistory.
		WithContext(ctx).
		Where(d.q.SearchHistory.UserID.Eq(userID)).
		Where(d.q.SearchHistory.SearchType.Eq(searchType)).
		Where(d.q.SearchHistory.TargetType.Eq(targetType)).
		Order(d.q.SearchHistory.LastSearchedAt.Desc()).
		FindByPage(offset, limit)
}
