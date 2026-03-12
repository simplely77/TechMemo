package dao

import (
	"context"
	"techmemo/backend/model"
	"techmemo/backend/query"

	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type AIDao struct {
	q  *query.Query
	db *gorm.DB
}

func (d *AIDao) GetKnowledgePointByID(ctx context.Context, id int64) (*model.KnowledgePoint, error) {
	return d.q.KnowledgePoint.WithContext(ctx).Where(d.q.KnowledgePoint.ID.Eq(id)).First()
}

func (d *AIDao) SaveEmbedding(ctx context.Context, embeddingModel *model.EmbeddingCustom) error {
	return d.db.WithContext(ctx).Exec(
		`INSERT INTO embedding (target_type, target_id, vector, model_name) VALUES (?, ?, ?, ?)`,
		embeddingModel.TargetType,
		embeddingModel.TargetID,
		pgvector.NewVector(embeddingModel.Vector),
		embeddingModel.ModelName,
	).Error
}

func (d *AIDao) SaveKnowledgePoints(ctx context.Context, knowledgePoints []*model.KnowledgePoint) error {
	return d.q.KnowledgePoint.WithContext(ctx).Create(knowledgePoints...)
}

func (d *AIDao) GetKnowledgePointsByNoteID(ctx context.Context, noteID int64) ([]*model.KnowledgePoint, error) {
	return d.q.KnowledgePoint.WithContext(ctx).
		Where(d.q.KnowledgePoint.SourceNoteID.Eq(noteID)).
		Find()
}

func (d *AIDao) SaveKnowledgeRelations(ctx context.Context, relations []*model.KnowledgeRelation) error {
	if len(relations) == 0 {
		return nil
	}
	return d.q.KnowledgeRelation.WithContext(ctx).Create(relations...)
}

func (d *AIDao) GetKnowledgeRelationsByNoteID(ctx context.Context, noteID int64) ([]*model.KnowledgeRelation, error) {
	kps, err := d.q.KnowledgePoint.WithContext(ctx).
		Where(d.q.KnowledgePoint.SourceNoteID.Eq(noteID)).
		Select(d.q.KnowledgePoint.ID).
		Find()
	if err != nil || len(kps) == 0 {
		return nil, err
	}

	ids := make([]int64, len(kps))
	for i, kp := range kps {
		ids[i] = kp.ID
	}

	return d.q.KnowledgeRelation.WithContext(ctx).
		Where(d.q.KnowledgeRelation.FromKnowledgeID.In(ids...)).
		Find()
}

func (d *AIDao) UpdateStatus(ctx context.Context, id int64, status string) {
	d.q.AiProcessLog.
		WithContext(ctx).
		Where(d.q.AiProcessLog.ID.Eq(id)).
		Update(d.q.AiProcessLog.Status, status)
}

func (d *AIDao) GetLogsByTaskID(ctx context.Context, taskID string) ([]*model.AiProcessLog, error) {
	return d.q.AiProcessLog.
		WithContext(ctx).
		Where(d.q.AiProcessLog.TaskID.Eq(taskID)).
		Find()
}

func (d *AIDao) CreateAILog(
	ctx context.Context,
	params CreateAILogParams,
) error {

	log := &model.AiProcessLog{
		SourceNoteID: params.SourceNoteID,
		TaskID:       params.TaskID,
		TargetType:   params.TargetType,
		TargetID:     params.TargetID,
		ProcessType:  params.ProcessType,
		ModelName:    params.ModelName,
		Status:       params.Status,
	}

	return d.q.AiProcessLog.WithContext(ctx).Create(log)
}

func (d *AIDao) GetLogsByNoteID(ctx context.Context, noteID int64) ([]*model.AiProcessLog, error) {
	return d.q.AiProcessLog.
		WithContext(ctx).
		Where(d.q.AiProcessLog.SourceNoteID.Eq(noteID)).
		Find()
}

func (d *AIDao) SaveNoteRootNode(ctx context.Context, node *model.NoteRootNode) error {
	return d.db.WithContext(ctx).Create(node).Error
}

func (d *AIDao) GetRootNodesByUserID(ctx context.Context, userID int64) ([]*model.NoteRootNode, error) {
	var nodes []*model.NoteRootNode
	err := d.db.WithContext(ctx).
		Joins("JOIN knowledge_point ON knowledge_point.id = note_root_node.root_knowledge_id").
		Where("knowledge_point.user_id = ?", userID).
		Find(&nodes).Error
	return nodes, err
}

func (d *AIDao) GetGlobalRelationsByUserID(ctx context.Context, userID int64) ([]*model.KnowledgeRelation, error) {
	// 查 relation_type = 'global' 且 from 节点属于该用户
	var relations []*model.KnowledgeRelation
	err := d.db.WithContext(ctx).
		Joins("JOIN knowledge_point ON knowledge_point.id = knowledge_relation.from_knowledge_id").
		Where("knowledge_relation.relation_type = ? AND knowledge_point.user_id = ?", "global", userID).
		Find(&relations).Error
	return relations, err
}

func (d *AIDao) DeleteGlobalRelationsByUserID(ctx context.Context, userID int64) error {
	return d.db.WithContext(ctx).Exec(`
		DELETE FROM knowledge_relation
		WHERE relation_type = 'global'
		AND from_knowledge_id IN (
			SELECT id FROM knowledge_point WHERE user_id = ?
		)`, userID).Error
}

func NewAIDao(q *query.Query, db *gorm.DB) *AIDao {
	return &AIDao{q: q, db: db}
}