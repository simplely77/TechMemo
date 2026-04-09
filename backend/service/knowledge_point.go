package service

import (
	"context"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
	"techmemo/backend/model"
	"time"
)

type KnowledgePointService struct {
	knowledgePointDao *dao.KnowledgePointDao
	noteDao           *dao.NoteDao
	searchDao         *dao.SearchDao
}

func (s *KnowledgePointService) GetKnowledgePoints(ctx context.Context, req *dto.GetKnowledgePointsReq, userID int64) (*dto.GetKnowledgePointsResp, error) {
	params := dao.GetKnowledgePointsParams{
		UserID:        userID,
		SourceNoteID:  req.SourceNoteID,
		Keyword:       req.Keyword,
		MinImportance: req.MinImportance,
		Page:          req.Page,
		PageSize:      req.PageSize,
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	knowledgePoints, total, err := s.knowledgePointDao.GetKnowledgePoints(ctx, params)
	if err != nil {
		return nil, errors.InternalErr
	}

	items := make([]dto.KnowledgePointItem, 0, len(knowledgePoints))
	for _, kp := range knowledgePoints {
		item := dto.KnowledgePointItem{
			ID:              kp.ID,
			Name:            kp.Name,
			Description:     kp.Description,
			ImportanceScore: kp.ImportanceScore,
			SourceNoteID:    kp.SourceNoteID,
			CreatedAt:       kp.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if kp.SourceNoteID > 0 {
			note, err := s.noteDao.GetNoteByID(ctx, kp.SourceNoteID)
			if err == nil && note != nil {
				item.SourceNoteTitle = note.Title
			}
		}

		items = append(items, item)
	}
	if req.Keyword != "" {
		searchHistory := &model.SearchHistory{
			UserID:         userID,
			Keyword:        req.Keyword,
			SearchType:     "keyword",
			TargetType:     "knowledge",
			LastSearchedAt: time.Now(),
		}
		if err := s.searchDao.SaveSearchHistory(ctx, searchHistory); err != nil {
			return nil, errors.InternalErr
		}
	}

	return &dto.GetKnowledgePointsResp{
		KnowledgePoints: items,
		Total:           total,
		Page:            params.Page,
		PageSize:        params.PageSize,
	}, nil
}

func (s *KnowledgePointService) GetKnowledgePoint(ctx context.Context, id int64, userID int64) (*dto.GetKnowledgePointResp, error) {
	kp, err := s.knowledgePointDao.GetKnowledgePointByID(ctx, id)
	if err != nil {
		return nil, errors.InternalErr
	}

	if kp == nil {
		return nil, errors.NotFound
	}

	if kp.UserID != userID {
		return nil, errors.Forbidden
	}

	resp := &dto.GetKnowledgePointResp{
		ID:              kp.ID,
		Name:            kp.Name,
		Description:     kp.Description,
		ImportanceScore: kp.ImportanceScore,
		SourceNoteID:    kp.SourceNoteID,
		CreatedAt:       kp.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if kp.SourceNoteID > 0 {
		note, err := s.noteDao.GetNoteByID(ctx, kp.SourceNoteID)
		if err == nil && note != nil {
			resp.SourceNoteTitle = note.Title
		}
	}

	relations, err := s.knowledgePointDao.GetRelatedKnowledge(ctx, id)
	if err == nil && len(relations) > 0 {
		relatedKnowledge := make([]dto.RelatedKnowledge, 0, len(relations))
		for _, rel := range relations {
			var relatedID int64
			var relationType string

			if rel.FromKnowledgeID == id {
				relatedID = rel.ToKnowledgeID
				relationType = rel.RelationType
			} else {
				relatedID = rel.FromKnowledgeID
				relationType = rel.RelationType
			}

			relatedKP, err := s.knowledgePointDao.GetKnowledgePointByID(ctx, relatedID)
			if err == nil && relatedKP != nil {
				relatedKnowledge = append(relatedKnowledge, dto.RelatedKnowledge{
					ID:           relatedID,
					Name:         relatedKP.Name,
					RelationType: relationType,
				})
			}
		}
		resp.RelatedKnowledge = relatedKnowledge
	}

	return resp, nil
}

func (s *KnowledgePointService) UpdateKnowledgePoint(ctx context.Context, id int64, req *dto.UpdateKnowledgePointReq, userID int64) error {
	kp, err := s.knowledgePointDao.GetKnowledgePointByID(ctx, id)
	if err != nil {
		return errors.InternalErr
	}

	if kp == nil {
		return errors.NotFound
	}

	if kp.UserID != userID {
		return errors.Forbidden
	}

	params := dao.UpdateKnowledgePointParams{
		Name:            req.Name,
		Description:     req.Description,
		ImportanceScore: req.ImportanceScore,
	}

	return s.knowledgePointDao.UpdateKnowledgePoint(ctx, id, params)
}

func (s *KnowledgePointService) DeleteKnowledgePoint(ctx context.Context, id int64, userID int64) error {
	kp, err := s.knowledgePointDao.GetKnowledgePointByID(ctx, id)
	if err != nil {
		return errors.InternalErr
	}

	if kp == nil {
		return errors.NotFound
	}

	if kp.UserID != userID {
		return errors.Forbidden
	}

	return s.knowledgePointDao.DeleteKnowledgePoint(ctx, id)
}

func NewKnowledgePointService(knowledgePointDao *dao.KnowledgePointDao, noteDao *dao.NoteDao, searchDao *dao.SearchDao) *KnowledgePointService {
	return &KnowledgePointService{
		knowledgePointDao: knowledgePointDao,
		noteDao:           noteDao,
		searchDao:         searchDao,
	}
}
