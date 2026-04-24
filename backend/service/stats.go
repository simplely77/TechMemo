package service

import (
	"context"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
)

type StatsService struct {
	noteDao           *dao.NoteDao
	categoryDao       *dao.CategoryDao
	knowledgePointDao *dao.KnowledgePointDao
	tagDao            *dao.TagDao
	aiDao             *dao.AIDao
}

func (s *StatsService) GetCategories(ctx context.Context, userID int64) (*dto.GetCategoriesStatsResp, error) {
	categories, err := s.categoryDao.GetCategoriesByUserID(ctx, userID)
	if err != nil {
		return nil, errors.InternalErr
	}
	noteCountByCat, err := s.noteDao.CountNotesByCategoryForUser(ctx, userID)
	if err != nil {
		return nil, errors.InternalErr
	}
	kpCountByCat, err := s.knowledgePointDao.CountKnowledgePointsByCategoryForUser(ctx, userID)
	if err != nil {
		return nil, errors.InternalErr
	}
	categoriesDto := make([]dto.CategoryStats, 0, len(categories))
	for _, category := range categories {
		nc := noteCountByCat[category.ID]
		kc := kpCountByCat[category.ID]
		categoriesDto = append(categoriesDto, dto.CategoryStats{
			CategoryID:     category.ID,
			CategoryName:   category.Name,
			NoteCount:      nc,
			KnowledgeCount: kc,
		})
	}
	return &dto.GetCategoriesStatsResp{
		Categories: categoriesDto,
	}, nil
}

func (s *StatsService) GetOverview(ctx context.Context, userID int64) (*dto.GetOverviewStatsResp, error) {
	notes, err := s.noteDao.CountNotesByUid(ctx, userID)
	if err != nil {
		return nil, errors.InternalErr
	}
	knowledgePoints, err := s.knowledgePointDao.CountKnowledgePointsByUid(ctx, userID)
	if err != nil {
		return nil, errors.InternalErr
	}
	categories, err := s.categoryDao.CountCategoriesByUid(ctx, userID)
	if err != nil {
		return nil, errors.InternalErr
	}
	tags, err := s.tagDao.CountTags(ctx, userID, "")
	if err != nil {
		return nil, errors.InternalErr
	}
	return &dto.GetOverviewStatsResp{
		TotalNotes:           notes,
		TotalKnowledgePoints: knowledgePoints,
		TotalCategories:      categories,
		TotalTags:            tags,
	}, nil
}

func NewStatsServcie(
	noteDao *dao.NoteDao,
	categoryDao *dao.CategoryDao,
	knowledgePointDao *dao.KnowledgePointDao,
	tagDao *dao.TagDao,
	aiDao *dao.AIDao,
) *StatsService {
	return &StatsService{
		noteDao:           noteDao,
		categoryDao:       categoryDao,
		knowledgePointDao: knowledgePointDao,
		tagDao:            tagDao,
		aiDao:             aiDao,
	}
}
