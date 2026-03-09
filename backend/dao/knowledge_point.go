package dao

import (
	"context"
	"techmemo/backend/model"
	"techmemo/backend/query"
)

type KnowledgePointDao struct {
	q *query.Query
}

func (d *KnowledgePointDao) GetKnowledgePoints(ctx context.Context, params GetKnowledgePointsParams) ([]*model.KnowledgePoint, int64, error) {
	q := d.q.KnowledgePoint.WithContext(ctx).
		Where(d.q.KnowledgePoint.UserID.Eq(params.UserID))

	if params.SourceNoteID > 0 {
		q = q.Where(d.q.KnowledgePoint.SourceNoteID.Eq(params.SourceNoteID))
	}

	if params.Keyword != "" {
		q = q.Where(d.q.KnowledgePoint.Name.Like("%" + params.Keyword + "%"))
	}

	if params.MinImportance > 0 {
		q = q.Where(d.q.KnowledgePoint.ImportanceScore.Gte(params.MinImportance))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	knowledgePoints, err := q.
		Offset(int(offset)).
		Limit(int(params.PageSize)).
		Order(d.q.KnowledgePoint.CreatedAt.Desc()).
		Find()

	return knowledgePoints, total, err
}

func (d *KnowledgePointDao) GetKnowledgePointByID(ctx context.Context, id int64) (*model.KnowledgePoint, error) {
	return d.q.KnowledgePoint.WithContext(ctx).
		Where(d.q.KnowledgePoint.ID.Eq(id)).
		First()
}

func (d *KnowledgePointDao) UpdateKnowledgePoint(ctx context.Context, id int64, params UpdateKnowledgePointParams) error {
	updateFields := make(map[string]interface{})

	if params.Name != "" {
		updateFields[d.q.KnowledgePoint.Name.ColumnName().String()] = params.Name
	}
	if params.Description != "" {
		updateFields[d.q.KnowledgePoint.Description.ColumnName().String()] = params.Description
	}
	if params.ImportanceScore > 0 {
		updateFields[d.q.KnowledgePoint.ImportanceScore.ColumnName().String()] = params.ImportanceScore
	}

	_, err := d.q.KnowledgePoint.WithContext(ctx).
		Where(d.q.KnowledgePoint.ID.Eq(id)).
		Updates(updateFields)

	return err
}

func (d *KnowledgePointDao) DeleteKnowledgePoint(ctx context.Context, id int64) error {
	_, err := d.q.KnowledgePoint.WithContext(ctx).
		Where(d.q.KnowledgePoint.ID.Eq(id)).
		Delete()
	return err
}

func (d *KnowledgePointDao) GetRelatedKnowledge(ctx context.Context, knowledgeID int64) ([]*model.KnowledgeRelation, error) {
	return d.q.KnowledgeRelation.WithContext(ctx).
		Where(d.q.KnowledgeRelation.FromKnowledgeID.Eq(knowledgeID)).
		Or(d.q.KnowledgeRelation.ToKnowledgeID.Eq(knowledgeID)).
		Find()
}

func NewKnowledgePointDao(q *query.Query) *KnowledgePointDao {
	return &KnowledgePointDao{q: q}
}
