package service

import (
	"context"
	"math"
	"sort"
	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
	"techmemo/backend/model"
	"time"
)

const (
	rrfK           = 60.0 // Reciprocal Rank Fusion 平滑常数（常用 60）
	hybridMinFetch = 5
	hybridMaxFetch = 100
	// rerank 送入 CrossEncoder 的正文长度上限（与 embedding 用全文索引对齐思路）
	rerankDocMaxRunes = 200
)

func hybridFetchSize(topK int) int {
	n := topK * 3
	if n < hybridMinFetch {
		n = hybridMinFetch
	}
	if n > hybridMaxFetch {
		n = hybridMaxFetch
	}
	return n
}

// mergeRRF 合并向量列表与关键词 ID 列表（均为 1-based 名次），返回按融合分降序的 ID 及归一化相似度 [0,1]。
func mergeRRF(vec []dao.SearchResult, kwIDs []int64, topK int) (ordered []int64, similarityByID map[int64]float64) {
	type acc struct{ score float64 }
	m := make(map[int64]*acc)
	for i, sr := range vec {
		r := i + 1
		a := m[sr.TargetID]
		if a == nil {
			a = &acc{}
			m[sr.TargetID] = a
		}
		a.score += 1.0 / (rrfK + float64(r))
	}
	for i, id := range kwIDs {
		r := i + 1
		a := m[id]
		if a == nil {
			a = &acc{}
			m[id] = a
		}
		a.score += 1.0 / (rrfK + float64(r))
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
	if len(pairs) > topK {
		pairs = pairs[:topK]
	}
	ordered = make([]int64, len(pairs))
	similarityByID = make(map[int64]float64, len(pairs))
	for i, p := range pairs {
		ordered[i] = p.id
		// 单路第一名约 1/61，乘 61 映射到 ~1；双路更高，截断到 1
		similarityByID[p.id] = math.Min(1.0, p.score*61.0)
	}
	return ordered, similarityByID
}

func sliceTopK(orderedIDs []int64, simByID map[int64]float64, k int) ([]int64, map[int64]float64) {
	if k <= 0 {
		return nil, nil
	}
	n := len(orderedIDs)
	if n > k {
		n = k
	}
	out := make([]int64, 0, n)
	outSim := make(map[int64]float64, n)
	for i := 0; i < n; i++ {
		id := orderedIDs[i]
		out = append(out, id)
		outSim[id] = simByID[id]
	}
	return out, outSim
}

func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes])
}

// sigmoid 将 CrossEncoder 原始分映射到 (0,1)，单调与排序一致，便于沿用 similarity 字段
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// rerankNoteResults 按 CrossEncoder 原始分降序。
// 返回：最终 id 列表、RRF 融合分（仅退回 RRF 时非 nil）、rerank 原始分（仅成功时非 nil）。
func (s *SearchService) rerankNoteResults(
	ctx context.Context,
	query string,
	orderedIDs []int64,
	rrfSim map[int64]float64,
	topK int,
) ([]int64, map[int64]float64, map[int64]float64) {
	notes, err := s.noteDao.GetNotesByIDs(ctx, orderedIDs)
	if err != nil {
		ids, sim := sliceTopK(orderedIDs, rrfSim, topK)
		return ids, sim, nil
	}
	noteByID := make(map[int64]*model.Note, len(notes))
	for _, n := range notes {
		noteByID[n.ID] = n
	}
	type cand struct {
		id  int64
		doc string
	}
	cands := make([]cand, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		n := noteByID[id]
		if n == nil {
			continue
		}
		cands = append(cands, cand{
			id:  id,
			doc: n.Title + "\n" + truncateRunes(n.ContentMd, rerankDocMaxRunes),
		})
	}
	if len(cands) == 0 {
		ids, sim := sliceTopK(orderedIDs, rrfSim, topK)
		return ids, sim, nil
	}
	docs := make([]string, len(cands))
	for i := range cands {
		docs[i] = cands[i].doc
	}
	scores, err := s.aiClient.Rerank(ctx, query, docs)
	if err != nil || len(scores) != len(docs) {
		ids, sim := sliceTopK(orderedIDs, rrfSim, topK)
		return ids, sim, nil
	}
	type scored struct {
		id  int64
		raw float64
	}
	pairs := make([]scored, len(cands))
	for i := range cands {
		pairs[i] = scored{id: cands[i].id, raw: scores[i]}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].raw != pairs[j].raw {
			return pairs[i].raw > pairs[j].raw
		}
		return pairs[i].id < pairs[j].id
	})
	if len(pairs) > topK {
		pairs = pairs[:topK]
	}
	outIDs := make([]int64, len(pairs))
	rawByID := make(map[int64]float64, len(pairs))
	for i, p := range pairs {
		outIDs[i] = p.id
		rawByID[p.id] = p.raw
	}
	return outIDs, nil, rawByID
}

func (s *SearchService) rerankKnowledgeResults(
	ctx context.Context,
	query string,
	orderedIDs []int64,
	rrfSim map[int64]float64,
	topK int,
) ([]int64, map[int64]float64, map[int64]float64) {
	kps, err := s.kpDao.GetKnowledgePointsByIDs(ctx, orderedIDs)
	if err != nil {
		ids, sim := sliceTopK(orderedIDs, rrfSim, topK)
		return ids, sim, nil
	}
	kpByID := make(map[int64]*model.KnowledgePoint, len(kps))
	for _, kp := range kps {
		kpByID[kp.ID] = kp
	}
	type cand struct {
		id  int64
		doc string
	}
	cands := make([]cand, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		kp := kpByID[id]
		if kp == nil {
			continue
		}
		cands = append(cands, cand{
			id:  id,
			doc: kp.Name + "\n" + truncateRunes(kp.Description, rerankDocMaxRunes),
		})
	}
	if len(cands) == 0 {
		ids, sim := sliceTopK(orderedIDs, rrfSim, topK)
		return ids, sim, nil
	}
	docs := make([]string, len(cands))
	for i := range cands {
		docs[i] = cands[i].doc
	}
	scores, err := s.aiClient.Rerank(ctx, query, docs)
	if err != nil || len(scores) != len(docs) {
		ids, sim := sliceTopK(orderedIDs, rrfSim, topK)
		return ids, sim, nil
	}
	type scored struct {
		id  int64
		raw float64
	}
	pairs := make([]scored, len(cands))
	for i := range cands {
		pairs[i] = scored{id: cands[i].id, raw: scores[i]}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].raw != pairs[j].raw {
			return pairs[i].raw > pairs[j].raw
		}
		return pairs[i].id < pairs[j].id
	})
	if len(pairs) > topK {
		pairs = pairs[:topK]
	}
	outIDs := make([]int64, len(pairs))
	rawByID := make(map[int64]float64, len(pairs))
	for i, p := range pairs {
		outIDs[i] = p.id
		rawByID[p.id] = p.raw
	}
	return outIDs, nil, rawByID
}

type SearchService struct {
	searchDao   *dao.SearchDao
	noteDao     *dao.NoteDao
	kpDao       *dao.KnowledgePointDao
	categoryDao *dao.CategoryDao
	aiClient    aiclient.AIClient
}

func (s *SearchService) SemanticSearch(
	ctx context.Context,
	req *dto.SemanticSearchReq,
	userID int64,
) (*dto.SemanticSearchResp, error) {
	if req.Query != "" {
		searchHistory := &model.SearchHistory{
			UserID:         userID,
			Keyword:        req.Query,
			SearchType:     "hybrid",
			TargetType:     req.SearchType,
			LastSearchedAt: time.Now(),
		}
		if err := s.searchDao.SaveSearchHistory(ctx, searchHistory); err != nil {
			return nil, errors.InternalErr
		}
	}

	fetchN := hybridFetchSize(req.TopK)

	// 1. 向量检索（多取候选供 RRF 融合）
	queryVector, err := s.aiClient.GetEmbedding(ctx, req.Query)
	if err != nil {
		return nil, errors.InternalErr
	}
	vecResults, err := s.searchDao.SearchEmbeddingsByVector(
		ctx,
		queryVector,
		req.SearchType,
		userID,
		fetchN,
	)
	if err != nil {
		return nil, errors.InternalErr
	}

	// 2. 关键词子串检索
	var kwIDs []int64
	if req.SearchType == "note" {
		kwIDs, err = s.searchDao.SearchNoteIDsByKeyword(ctx, userID, req.Query, fetchN)
	} else {
		kwIDs, err = s.searchDao.SearchKnowledgeIDsByKeyword(ctx, userID, req.Query, fetchN)
	}
	if err != nil {
		return nil, errors.InternalErr
	}

	// 3. RRF 混合排序（候选池）
	orderedIDs, simByID := mergeRRF(vecResults, kwIDs, fetchN)
	if len(orderedIDs) == 0 {
		return &dto.SemanticSearchResp{
			Results: []dto.SearchResultItem{},
			Query:   req.Query,
			Total:   0,
		}, nil
	}

	// 4. 可选：CrossEncoder 重排序（按原始分排序）；未配置或失败则退回 RRF 并截断 top_k
	var finalIDs []int64
	var displaySim map[int64]float64
	var rerankRaw map[int64]float64
	if s.aiClient.RerankEnabled() {
		if req.SearchType == "note" {
			finalIDs, displaySim, rerankRaw = s.rerankNoteResults(ctx, req.Query, orderedIDs, simByID, req.TopK)
		} else {
			finalIDs, displaySim, rerankRaw = s.rerankKnowledgeResults(ctx, req.Query, orderedIDs, simByID, req.TopK)
		}
	} else {
		finalIDs, displaySim = sliceTopK(orderedIDs, simByID, req.TopK)
	}

	var resultItems []dto.SearchResultItem
	if req.SearchType == "note" {
		resultItems, err = s.buildNoteResults(ctx, finalIDs, displaySim, rerankRaw)
	} else {
		resultItems, err = s.buildKnowledgeResults(ctx, finalIDs, displaySim, rerankRaw)
	}
	if err != nil {
		return nil, errors.InternalErr
	}

	return &dto.SemanticSearchResp{
		Results: resultItems,
		Query:   req.Query,
		Total:   len(resultItems),
	}, nil
}

// GetSearchHistory 分页获取当前用户的搜索历史（按最近搜索时间倒序）
func (s *SearchService) GetSearchHistory(
	ctx context.Context,
	req *dto.GetSearchHistoryReq,
	userID int64,
) (*dto.GetSearchHistoryResp, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	offset := int((req.Page - 1) * req.PageSize)
	limit := int(req.PageSize)

	rows, total, err := s.searchDao.ListSearchHistory(ctx, userID, offset, limit, req.SearchType, req.TargetType)
	if err != nil {
		return nil, errors.InternalErr
	}

	items := make([]dto.SearchHistoryItem, 0, len(rows))
	for _, h := range rows {
		items = append(items, dto.SearchHistoryItem{
			ID:             h.ID,
			Keyword:        h.Keyword,
			SearchType:     h.SearchType,
			TargetType:     h.TargetType,
			LastSearchedAt: h.LastSearchedAt,
			CreatedAt:      h.CreatedAt,
		})
	}

	return &dto.GetSearchHistoryResp{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *SearchService) buildNoteResults(
	ctx context.Context,
	orderedIDs []int64,
	rrfSimByID map[int64]float64,
	rerankRawByID map[int64]float64,
) ([]dto.SearchResultItem, error) {

	notes, err := s.noteDao.GetNotesByIDs(ctx, orderedIDs)
	if err != nil {
		return nil, err
	}

	noteByID := make(map[int64]*model.Note, len(notes))
	for _, n := range notes {
		noteByID[n.ID] = n
	}

	results := make([]dto.SearchResultItem, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		note := noteByID[id]
		if note == nil {
			continue
		}
		category, _ := s.categoryDao.GetCategoryByID(ctx, note.CategoryID)

		var item dto.SearchResultItem
		if rerankRawByID != nil {
			if raw, ok := rerankRawByID[id]; ok {
				rs := raw
				item = dto.SearchResultItem{
					ID:          note.ID,
					Type:        "note",
					Title:       note.Title,
					Content:     truncateContent(note.ContentMd, 200),
					Similarity:  sigmoid(raw),
					RerankScore: &rs,
					NoteType:    note.NoteType,
					CreatedAt:   note.CreatedAt.Format("2006-01-02 15:04:05"),
				}
			}
		}
		if item.ID == 0 {
			item = dto.SearchResultItem{
				ID:         note.ID,
				Type:       "note",
				Title:      note.Title,
				Content:    truncateContent(note.ContentMd, 200),
				Similarity: rrfSimByID[id],
				NoteType:   note.NoteType,
				CreatedAt:  note.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		if category != nil {
			item.Category = &dto.NoteCategory{
				ID:   category.ID,
				Name: category.Name,
			}
		}

		results = append(results, item)
	}

	return results, nil
}

func (s *SearchService) buildKnowledgeResults(
	ctx context.Context,
	orderedIDs []int64,
	rrfSimByID map[int64]float64,
	rerankRawByID map[int64]float64,
) ([]dto.SearchResultItem, error) {

	knowledgePoints, err := s.kpDao.GetKnowledgePointsByIDs(ctx, orderedIDs)
	if err != nil {
		return nil, err
	}

	kpByID := make(map[int64]*model.KnowledgePoint, len(knowledgePoints))
	for _, kp := range knowledgePoints {
		kpByID[kp.ID] = kp
	}

	results := make([]dto.SearchResultItem, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		kp := kpByID[id]
		if kp == nil {
			continue
		}
		var sourceNoteTitle string
		if kp.SourceNoteID > 0 {
			note, _ := s.noteDao.GetNoteByID(ctx, kp.SourceNoteID)
			if note != nil {
				sourceNoteTitle = note.Title
			}
		}

		var item dto.SearchResultItem
		if rerankRawByID != nil {
			if raw, ok := rerankRawByID[id]; ok {
				rs := raw
				item = dto.SearchResultItem{
					ID:              kp.ID,
					Type:            "knowledge",
					Title:           kp.Name,
					Content:         truncateContent(kp.Description, 200),
					Similarity:      sigmoid(raw),
					RerankScore:     &rs,
					SourceNoteID:    kp.SourceNoteID,
					SourceNoteTitle: sourceNoteTitle,
					ImportanceScore: kp.ImportanceScore,
					CreatedAt:       kp.CreatedAt.Format("2006-01-02 15:04:05"),
				}
			}
		}
		if item.ID == 0 {
			item = dto.SearchResultItem{
				ID:              kp.ID,
				Type:            "knowledge",
				Title:           kp.Name,
				Content:         truncateContent(kp.Description, 200),
				Similarity:      rrfSimByID[id],
				SourceNoteID:    kp.SourceNoteID,
				SourceNoteTitle: sourceNoteTitle,
				ImportanceScore: kp.ImportanceScore,
				CreatedAt:       kp.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		results = append(results, item)
	}

	return results, nil
}

// truncateContent 截断内容到指定长度
func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "..."
}

func NewSearchService(
	searchDao *dao.SearchDao,
	noteDao *dao.NoteDao,
	kpDao *dao.KnowledgePointDao,
	categoryDao *dao.CategoryDao,
	aiClient aiclient.AIClient,
) *SearchService {
	return &SearchService{
		searchDao:   searchDao,
		noteDao:     noteDao,
		kpDao:       kpDao,
		categoryDao: categoryDao,
		aiClient:    aiClient,
	}
}
