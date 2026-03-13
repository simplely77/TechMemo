package dao

import (
	"context"
	"techmemo/backend/query"

	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
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
) ([]SearchResult, error) {
	var results []SearchResult

	// 构建 SQL 查询
	// 使用 pgvector 的余弦距离操作符 <=>
	// 需要 JOIN 对应的表来过滤 user_id
	var querySQL string
	if targetType == "note" {
		querySQL = `
			SELECT e.target_id, e.target_type, (e.vector <=> $1) as distance
			FROM embedding e
			INNER JOIN note n ON e.target_id = n.id
			WHERE e.target_type = 'note'
			  AND n.user_id = $2
			  AND n.status != 'deleted'
			ORDER BY distance
			LIMIT $3
		`
	} else { // knowledge
		querySQL = `
			SELECT e.target_id, e.target_type, (e.vector <=> $1) as distance
			FROM embedding e
			INNER JOIN knowledge_point kp ON e.target_id = kp.id
			WHERE e.target_type = 'knowledge'
			  AND kp.user_id = $2
			ORDER BY distance
			LIMIT $3
		`
	}

	err := d.db.WithContext(ctx).Raw(
		querySQL,
		pgvector.NewVector(vector),
		userID,
		topK,
	).Scan(&results).Error

	return results, err
}

func NewSearchDao(q *query.Query, db *gorm.DB) *SearchDao {
	return &SearchDao{q: q, db: db}
}
