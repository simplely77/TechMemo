package dao

import (
	"context"
	"math"
	"sort"
	"strings"
	"techmemo/backend/model"
	"techmemo/backend/query"
	"time"

	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// RRFK Reciprocal Rank Fusion 平滑常数（常用 60）
	RRFK = 60.0
	hybridMinFetch = 5
	hybridMaxFetch = 100
)

// HybridFetchN 混合检索时每路候选条数：约为 topK 的 3 倍，夹在 [5,100]，与语义搜索 API 一致。
func HybridFetchN(topK int) int {
	n := topK * 3
	if n < hybridMinFetch {
		n = hybridMinFetch
	}
	if n > hybridMaxFetch {
		n = hybridMaxFetch
	}
	return n
}

// MergeRRF 合并向量检索结果与关键词命中的 ID 列表（两路均为 1-based 名次），按融合分降序返回 ID 及归一化相似度 [0,1]；outLimit 为融合后保留的最大条数。
func MergeRRF(vec []SearchResult, kwIDs []int64, outLimit int) (ordered []int64, similarityByID map[int64]float64) {
	type acc struct{ score float64 }
	m := make(map[int64]*acc)
	for i, sr := range vec {
		r := i + 1
		a := m[sr.TargetID]
		if a == nil {
			a = &acc{}
			m[sr.TargetID] = a
		}
		a.score += 1.0 / (RRFK + float64(r))
	}
	for i, id := range kwIDs {
		r := i + 1
		a := m[id]
		if a == nil {
			a = &acc{}
			m[id] = a
		}
		a.score += 1.0 / (RRFK + float64(r))
	}
	type pair struct {
		id    int64
		score float64
	}
	pairs := make([]pair, 0, len(m))
	for id, a := range m {
		pairs = append(pairs, pair{id: id, score: a.score})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].score != pairs[j].score {
			return pairs[i].score > pairs[j].score
		}
		return pairs[i].id < pairs[j].id
	})
	if outLimit > 0 && len(pairs) > outLimit {
		pairs = pairs[:outLimit]
	}
	ordered = make([]int64, len(pairs))
	similarityByID = make(map[int64]float64, len(pairs))
	for i, p := range pairs {
		ordered[i] = p.id
		// 单路第一名 score≈1/(RRFK+1)，乘 (RRFK+1) 映射到约 1
		similarityByID[p.id] = math.Min(1.0, p.score*(RRFK+1))
	}
	return ordered, similarityByID
}

// HybridSearchEmbeddings 对笔记或知识点做「向量 + 关键词子串」混合检索（RRF），返回最多 topK 条，与 /search/semantic 前两步一致（不含 CrossEncoder rerank）。
func (d *SearchDao) HybridSearchEmbeddings(
	ctx context.Context,
	vector []float32,
	userID int64,
	queryText string,
	targetType string,
	topK int,
) ([]SearchResult, error) {
	if topK <= 0 {
		return nil, nil
	}
	fetchN := HybridFetchN(topK)
	vec, err := d.SearchEmbeddingsByVector(ctx, vector, targetType, userID, fetchN)
	if err != nil {
		return nil, err
	}
	var kwIDs []int64
	switch targetType {
	case "note":
		kwIDs, err = d.SearchNoteIDsByKeyword(ctx, userID, queryText, fetchN)
	case "knowledge":
		kwIDs, err = d.SearchKnowledgeIDsByKeyword(ctx, userID, queryText, fetchN)
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	orderedIDs, simByID := MergeRRF(vec, kwIDs, fetchN)
	if len(orderedIDs) == 0 {
		return nil, nil
	}
	if len(orderedIDs) > topK {
		orderedIDs = orderedIDs[:topK]
	}
	out := make([]SearchResult, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		out = append(out, SearchResult{
			TargetID:   id,
			TargetType: targetType,
			Distance:   1.0 - simByID[id],
		})
	}
	return out, nil
}

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
