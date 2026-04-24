package dao

import (
	"context"
	"techmemo/backend/model"
	"techmemo/backend/query"
)

type KnowledgePointDao struct {
	q *query.Query
}

func (d *KnowledgePointDao) CountKnowledgePointsByNids(ctx context.Context, nids []int64) (int64, error) {
	return d.q.KnowledgePoint.
		WithContext(ctx).
		Where(d.q.KnowledgePoint.SourceNoteID.In(nids...)).
		Count()
}

func (d *KnowledgePointDao) CountKnowledgePointsByUid(ctx context.Context, userID int64) (int64, error) {
	return d.q.KnowledgePoint.
		WithContext(ctx).
		Where(d.q.KnowledgePoint.UserID.Eq(userID)).
		Count()
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

func (d *KnowledgePointDao) GetKnowledgePointsByIDs(ctx context.Context, ids []int64) ([]*model.KnowledgePoint, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return d.q.KnowledgePoint.
		WithContext(ctx).
		Where(d.q.KnowledgePoint.ID.In(ids...)).
		Find()
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

func (d *KnowledgePointDao) GetKnowledgePointsBySourceNoteID(ctx context.Context, noteID int64) ([]*model.KnowledgePoint, error) {
	return d.q.KnowledgePoint.WithContext(ctx).
		Where(d.q.KnowledgePoint.SourceNoteID.Eq(noteID)).
		Find()
}

func (d *KnowledgePointDao) DeleteKnowledgePoint(ctx context.Context, id int64) error {
	// 删除知识点关联关系（作为源或目标）
	if _, err := d.q.KnowledgeRelation.WithContext(ctx).
		Where(d.q.KnowledgeRelation.FromKnowledgeID.Eq(id)).
		Delete(); err != nil {
		return err
	}
	if _, err := d.q.KnowledgeRelation.WithContext(ctx).
		Where(d.q.KnowledgeRelation.ToKnowledgeID.Eq(id)).
		Delete(); err != nil {
		return err
	}

	// 删除知识点的 embedding
	if _, err := d.q.Embedding.WithContext(ctx).
		Where(d.q.Embedding.TargetType.Eq("knowledge")).
		Where(d.q.Embedding.TargetID.Eq(id)).
		Delete(); err != nil {
		return err
	}

	// 删除知识点本身
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

// CountKnowledgePointsByCategoryForUser 按笔记所属分类统计知识点数量（仅统计 source_note 指向未删除笔记的知识点）
func (d *KnowledgePointDao) CountKnowledgePointsByCategoryForUser(ctx context.Context, userID int64) (map[int64]int64, error) {
	var rows []struct {
		CategoryID int64 `gorm:"column:category_id"`
		Cnt        int64 `gorm:"column:cnt"`
	}
	db := d.q.KnowledgePoint.WithContext(ctx).UnderlyingDB().WithContext(ctx)
	err := db.Raw(`
		SELECT n.category_id, COUNT(kp.id)::bigint AS cnt
		FROM knowledge_point kp
		INNER JOIN note n ON n.id = kp.source_note_id
		WHERE n.user_id = ? AND n.status != 'deleted'
		GROUP BY n.category_id
	`, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int64]int64, len(rows))
	for i := range rows {
		out[rows[i].CategoryID] = rows[i].Cnt
	}
	return out, nil
}

func NewKnowledgePointDao(q *query.Query) *KnowledgePointDao {
	return &KnowledgePointDao{q: q}
}
