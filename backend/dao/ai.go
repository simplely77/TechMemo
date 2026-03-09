package dao

import (
	"context"
	"techmemo/backend/model"
	"techmemo/backend/query"

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
	// 将自定义类型转换为数据库模型
	dbEmbedding, err := embeddingModel.ToDBFormat()
	if err != nil {
		return err
	}

	// 使用生成的模型保存到数据库
	return d.q.Embedding.WithContext(ctx).Create(dbEmbedding)
}

func (d *AIDao) SaveKnowledgePoints(ctx context.Context, knowledgePoints []*model.KnowledgePoint) error {
	return d.q.KnowledgePoint.WithContext(ctx).Create(knowledgePoints...)
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

func NewAIDao(q *query.Query, db *gorm.DB) *AIDao {
	return &AIDao{q: q, db: db}
}
