package service

import (
	"context"
	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
)

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

	// 1. 将查询文本转换为向量
	queryVector, err := s.aiClient.GetEmbedding(ctx, req.Query)
	if err != nil {
		return nil, errors.InternalErr
	}

	// 2. 在数据库中搜索相似向量
	searchResults, err := s.searchDao.SearchEmbeddingsByVector(
		ctx,
		queryVector,
		req.SearchType,
		userID,
		req.TopK,
	)
	if err != nil {
		return nil, errors.InternalErr
	}

	if len(searchResults) == 0 {
		return &dto.SemanticSearchResp{
			Results: []dto.SearchResultItem{},
			Query:   req.Query,
			Total:   0,
		}, nil
	}

	// 3. 根据 search_type 获取详细信息
	var resultItems []dto.SearchResultItem

	if req.SearchType == "note" {
		resultItems, err = s.buildNoteResults(ctx, searchResults)
	} else {
		resultItems, err = s.buildKnowledgeResults(ctx, searchResults)
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

func (s *SearchService) buildNoteResults(
	ctx context.Context,
	searchResults []dao.SearchResult,
) ([]dto.SearchResultItem, error) {

	// 提取 note IDs
	noteIDs := make([]int64, len(searchResults))
	distanceMap := make(map[int64]float64)
	for i, sr := range searchResults {
		noteIDs[i] = sr.TargetID
		distanceMap[sr.TargetID] = sr.Distance
	}

	// 批量查询笔记
	notes, err := s.noteDao.GetNotesByIDs(ctx, noteIDs)
	if err != nil {
		return nil, err
	}

	// 构建结果
	results := make([]dto.SearchResultItem, 0, len(notes))
	for _, note := range notes {
		category, _ := s.categoryDao.GetCategoryByID(ctx, note.CategoryID)

		// 将余弦距离转换为相似度分数 (0-1)
		// 余弦距离范围 [0, 2]，0 表示完全相同
		similarity := 1.0 - (distanceMap[note.ID] / 2.0)

		item := dto.SearchResultItem{
			ID:         note.ID,
			Type:       "note",
			Title:      note.Title,
			Content:    truncateContent(note.ContentMd, 200),
			Similarity: similarity,
			NoteType:   note.NoteType,
			CreatedAt:  note.CreatedAt.Format("2006-01-02 15:04:05"),
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
	searchResults []dao.SearchResult,
) ([]dto.SearchResultItem, error) {

	// 提取 knowledge point IDs
	kpIDs := make([]int64, len(searchResults))
	distanceMap := make(map[int64]float64)
	for i, sr := range searchResults {
		kpIDs[i] = sr.TargetID
		distanceMap[sr.TargetID] = sr.Distance
	}

	// 批量查询知识点
	knowledgePoints, err := s.kpDao.GetKnowledgePointsByIDs(ctx, kpIDs)
	if err != nil {
		return nil, err
	}

	// 构建结果
	results := make([]dto.SearchResultItem, 0, len(knowledgePoints))
	for _, kp := range knowledgePoints {
		// 获取来源笔记标题
		var sourceNoteTitle string
		if kp.SourceNoteID > 0 {
			note, _ := s.noteDao.GetNoteByID(ctx, kp.SourceNoteID)
			if note != nil {
				sourceNoteTitle = note.Title
			}
		}

		similarity := 1.0 - (distanceMap[kp.ID] / 2.0)

		item := dto.SearchResultItem{
			ID:              kp.ID,
			Type:            "knowledge",
			Title:           kp.Name,
			Content:         truncateContent(kp.Description, 200),
			Similarity:      similarity,
			SourceNoteID:    kp.SourceNoteID,
			SourceNoteTitle: sourceNoteTitle,
			ImportanceScore: kp.ImportanceScore,
			CreatedAt:       kp.CreatedAt.Format("2006-01-02 15:04:05"),
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

