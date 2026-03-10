package service

import (
	"context"
	"fmt"
	"log"
	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/ai/queue"
	"techmemo/backend/common/errors"
	"techmemo/backend/config"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
	"techmemo/backend/model"

	"github.com/google/uuid"
)

type AIService struct {
	aiDao    *dao.AIDao
	noteDao  *dao.NoteDao
	queue    queue.Queue
	aiClient aiclient.AIClient
}

func (a *AIService) GetNoteAIStatus(ctx context.Context, noteID int64) (dto.GetNoteAIStatusResp, error) {
	logs, err := a.aiDao.GetLogsByNoteID(ctx, noteID)
	if err != nil {
		return dto.GetNoteAIStatusResp{}, errors.InternalErr
	}

	if len(logs) == 0 {
		return dto.GetNoteAIStatusResp{
			NoteID: noteID,
			Status: "not_started",
		}, nil
	}

	progress := make(map[string]string)
	overallStatus := "completed"

	for _, logItem := range logs {
		progress[logItem.ProcessType] = logItem.Status

		switch logItem.Status {
		case "failed":
			overallStatus = "failed"
		case "processing":
			if overallStatus != "failed" {
				overallStatus = "processing"
			}
		case "pending":
			if overallStatus != "failed" && overallStatus != "processing" {
				overallStatus = "pending"
			}
		}
	}

	return dto.GetNoteAIStatusResp{
		NoteID:   noteID,
		Status:   overallStatus,
		Progress: progress,
	}, nil
}

func (a *AIService) GetQueue() queue.Queue {
	return a.queue
}

func (a *AIService) SetQueue(q queue.Queue) {
	a.queue = q
}

func (a *AIService) SetAIClient(client aiclient.AIClient) {
	a.aiClient = client
}

func (a *AIService) SubmitTask(ctx context.Context, noteID int64) (string, error) {
	taskID := generateTaskID(noteID)

	err := a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
		SourceNoteID: noteID,
		TaskID:       taskID,
		TargetType:   "note",
		TargetID:     noteID,
		ProcessType:  "classify",
		ModelName:    config.AppConfig.AI.Chat.Model,
		Status:       "pending",
	})
	if err != nil {
		return "", errors.InternalErr
	}

	if err := a.queue.Publish(queue.AITask{TaskID: taskID}); err != nil {
		return "", errors.InternalErr
	}
	return taskID, nil
}

// ProcessTask 处理一个 AI 任务，由 Worker 调用。
func (a *AIService) ProcessTask(ctx context.Context, task queue.AITask) {
	logs, err := a.aiDao.GetLogsByTaskID(ctx, task.TaskID)
	if err != nil {
		log.Println("获取任务日志失败:", err)
		return
	}

	for _, logItem := range logs {
		if logItem.Status != "pending" {
			continue
		}

		a.aiDao.UpdateStatus(ctx, logItem.ID, "processing")

		switch logItem.ProcessType {
		case "classify":
			a.handleClassify(ctx, logItem)
		case "extract":
			a.handleExtract(ctx, logItem)
		case "embedding":
			a.handleEmbedding(ctx, logItem)
		default:
			a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		}
	}
}

func (a *AIService) handleClassify(ctx context.Context, logItem *model.AiProcessLog) {
	note, err := a.noteDao.GetNoteByID(ctx, logItem.TargetID)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("获取笔记失败: %v", err)
		return
	}

	noteType, err := a.aiClient.ClassifyNoteType(ctx, note.ContentMd)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("分类笔记类型失败: %v", err)
		return
	}

	if err := a.noteDao.UpdateNote(ctx, logItem.TargetID, dao.UpdateNoteParams{NoteType: &noteType}); err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("更新笔记类型失败: %v", err)
		return
	}

	switch noteType {
	case "knowledge":
		_ = a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "extract",
			ModelName:    a.aiClient.GetChatModelName(),
			Status:       "pending",
		})
		_ = a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "embedding",
			ModelName:    a.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		_ = a.queue.Publish(queue.AITask{TaskID: logItem.TaskID})
	case "reference":
		_ = a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "embedding",
			ModelName:    a.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		_ = a.queue.Publish(queue.AITask{TaskID: logItem.TaskID})
	default:
		// ignore 类型，直接完成
	}

	a.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
	log.Printf("笔记分类完成，笔记ID: %d, 类型: %s", logItem.TargetID, noteType)
}

func (a *AIService) handleExtract(ctx context.Context, logItem *model.AiProcessLog) {
	note, err := a.noteDao.GetNoteByID(ctx, logItem.TargetID)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("获取笔记失败: %v", err)
		return
	}

	knowledgePoints, err := a.aiClient.ExtractKnowledgePoints(ctx, note.ContentMd)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("抽取知识点失败: %v", err)
		return
	}

	kpModels := make([]*model.KnowledgePoint, 0, len(knowledgePoints))
	for _, kp := range knowledgePoints {
		kpModels = append(kpModels, &model.KnowledgePoint{
			UserID:          note.UserID,
			Name:            kp.Name,
			Description:     kp.Description,
			ImportanceScore: kp.ImportanceScore,
			SourceNoteID:    logItem.TargetID,
		})
	}

	if err := a.aiDao.SaveKnowledgePoints(ctx, kpModels); err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("保存知识点失败: %v", err)
		return
	}

	for _, kp := range kpModels {
		_ = a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "knowledge",
			TargetID:     kp.ID,
			ProcessType:  "embedding",
			ModelName:    a.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		_ = a.queue.Publish(queue.AITask{TaskID: logItem.TaskID})
	}

	a.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
	log.Printf("知识抽取完成，笔记ID: %d, 抽取知识点数: %d", logItem.TargetID, len(knowledgePoints))
}

func (a *AIService) handleEmbedding(ctx context.Context, logItem *model.AiProcessLog) {
	var text string

	switch logItem.TargetType {
	case "note":
		note, err := a.noteDao.GetNoteByID(ctx, logItem.TargetID)
		if err != nil {
			a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
			log.Printf("获取笔记失败: %v", err)
			return
		}
		text = note.ContentMd
	case "knowledge":
		kp, err := a.aiDao.GetKnowledgePointByID(ctx, logItem.TargetID)
		if err != nil {
			a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
			log.Printf("获取知识点失败: %v", err)
			return
		}
		text = kp.Name + "\n" + kp.Description
	default:
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		return
	}

	embedding, err := a.aiClient.GetEmbedding(ctx, text)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("获取向量失败: %v", err)
		return
	}

	if err := a.aiDao.SaveEmbedding(ctx, &model.EmbeddingCustom{
		TargetType: logItem.TargetType,
		TargetID:   logItem.TargetID,
		Vector:     embedding,
		ModelName:  a.aiClient.GetEmbeddingModelName(),
	}); err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("保存向量失败: %v", err)
		return
	}

	a.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
}

func generateTaskID(id int64) string {
	return fmt.Sprintf("ai:%d:%s", id, uuid.NewString())
}

func NewAIService(aiDao *dao.AIDao, noteDao *dao.NoteDao) *AIService {
	return &AIService{
		aiDao:   aiDao,
		noteDao: noteDao,
	}
}
