package dao

import (
	"context"
	"strings"
	"techmemo/backend/model"
	"techmemo/backend/query"
	"time"

	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// likeSubstringPattern 构造 ILIKE 子串匹配的 pattern，并转义 % _ \
func likeSubstringPattern(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)
	return "%" + q + "%"
}

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

// SearchEmbeddingsByVector 使用向量搜索 embedding 表。
// 使用 pgvector 余弦距离 <=>，按距离升序取 topK（不做额外距离阈值，避免误杀相关结果）。
func (d *SearchDao) SearchEmbeddingsByVector(
	ctx context.Context,
	vector []float32,
	targetType string,
	userID int64,
	topK int,
) ([]SearchResult, error) {
	var results []SearchResult

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
	).Scan(&results).Error

	return results, err
}

// SearchNoteIDsByKeyword 标题/正文子串匹配（ILIKE），按标题优先、更新时间倒序，返回有序 ID。
func (d *SearchDao) SearchNoteIDsByKeyword(
	ctx context.Context,
	userID int64,
	query string,
	limit int,
) ([]int64, error) {
	q := strings.TrimSpace(query)
	if q == "" || limit <= 0 {
		return nil, nil
	}
	pat := likeSubstringPattern(q)
	var ids []int64
	err := d.db.WithContext(ctx).Raw(`
		SELECT id FROM note
		WHERE user_id = ?
		  AND status != 'deleted'
		  AND (title ILIKE ? ESCAPE '\'
		   OR content_md ILIKE ? ESCAPE '\')
		ORDER BY
		  CASE WHEN title ILIKE ? ESCAPE '\' THEN 0 ELSE 1 END,
		  updated_at DESC
		LIMIT ?
	`, userID, pat, pat, pat, limit).Scan(&ids).Error
	return ids, err
}

// SearchKnowledgeIDsByKeyword 名称/描述子串匹配，名称优先。
func (d *SearchDao) SearchKnowledgeIDsByKeyword(
	ctx context.Context,
	userID int64,
	query string,
	limit int,
) ([]int64, error) {
	q := strings.TrimSpace(query)
	if q == "" || limit <= 0 {
		return nil, nil
	}
	pat := likeSubstringPattern(q)
	var ids []int64
	err := d.db.WithContext(ctx).Raw(`
		SELECT id FROM knowledge_point
		WHERE user_id = ?
		  AND (name ILIKE ? ESCAPE '\'
		   OR COALESCE(description, '') ILIKE ? ESCAPE '\')
		ORDER BY
		  CASE WHEN name ILIKE ? ESCAPE '\' THEN 0 ELSE 1 END,
		  id DESC
		LIMIT ?
	`, userID, pat, pat, pat, limit).Scan(&ids).Error
	return ids, err
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
