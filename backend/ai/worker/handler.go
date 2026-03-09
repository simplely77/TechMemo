package worker

import (
	"context"
	"log"
	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/ai/queue"
	"techmemo/backend/dao"
	"techmemo/backend/model"
	"techmemo/backend/service"
)

type Handler struct {
	aiService *service.AIService
	aiDao     *dao.AIDao
	noteDao   *dao.NoteDao
	aiClient  aiclient.AIClient
}

func NewHandler(aiService *service.AIService, aiDao *dao.AIDao, noteDao *dao.NoteDao, aiClient aiclient.AIClient) *Handler {
	return &Handler{
		aiService: aiService,
		aiDao:     aiDao,
		noteDao:   noteDao,
		aiClient:  aiClient,
	}
}

func (h *Handler) Handler(ctx context.Context, task queue.AITask) {
	logs, err := h.aiDao.GetLogsByTaskID(ctx, task.TaskID)
	if err != nil {
		log.Println("get logs failed:", err)
		return
	}

	for _, logItem := range logs {
		if logItem.Status != "pending" {
			continue
		}

		h.aiDao.UpdateStatus(ctx, logItem.ID, "processing")

		switch logItem.ProcessType {
		case "classify":
			h.handleClassify(ctx, logItem)
		case "extract":
			h.handleExtract(ctx, logItem)
		case "embedding":
			h.handleEmbedding(ctx, logItem)
		default:
			h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		}
	}
}

func (h *Handler) handleClassify(ctx context.Context, logItem *model.AiProcessLog) {
	// 获取笔记内容
	note, err := h.noteDao.GetNoteByID(ctx, logItem.TargetID)
	if err != nil {
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("获取笔记失败: %v", err)
		return
	}

	// 分类笔记类型
	noteType, err := h.aiClient.ClassifyNoteType(ctx, note.ContentMd)
	if err != nil {
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("分类笔记类型失败: %v", err)
		return
	}

	err = h.noteDao.UpdateNote(ctx, logItem.TargetID, dao.UpdateNoteParams{
		NoteType: &noteType,
	})
	if err != nil {
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("更新笔记类型失败: %v", err)
		return
	}

	switch noteType {
	case "knowledge":
		_ = h.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "extract",
			ModelName:    h.aiClient.GetChatModelName(),
			Status:       "pending",
		})
		_ = h.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "embedding",
			ModelName:    h.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		h.aiService.GetQueue().Publish(queue.AITask{
			TaskID: logItem.TaskID,
		})
	case "reference":
		// 普通笔记，不需要抽取知识点
		_ = h.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "embedding",
			ModelName:    h.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		h.aiService.GetQueue().Publish(queue.AITask{
			TaskID: logItem.TaskID,
		})
	default:
		h.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
		return
	}

	h.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
	log.Printf("笔记分类完成，笔记ID: %d, 类型: %s", logItem.TargetID, noteType)
}

// handleExtract 处理知识抽取任务
func (h *Handler) handleExtract(ctx context.Context, logItem *model.AiProcessLog) {
	// 获取笔记内容
	note, err := h.noteDao.GetNoteByID(ctx, logItem.TargetID)
	if err != nil {
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("获取笔记失败: %v", err)
		return
	}

	// 抽取知识点
	knowledgePoints, err := h.aiClient.ExtractKnowledgePoints(ctx, note.ContentMd)
	if err != nil {
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("抽取知识点失败: %v", err)
		return
	}

	// 保存知识点到数据库（需要实现对应的DAO方法）
	knowledgePointsModel := make([]*model.KnowledgePoint, 0, len(knowledgePoints))
	for _, kp := range knowledgePoints {
		knowledgePointsModel = append(knowledgePointsModel, &model.KnowledgePoint{
			UserID:          note.UserID,
			Name:            kp.Name,
			Description:     kp.Description,
			ImportanceScore: kp.ImportanceScore,
			SourceNoteID:    logItem.TargetID,
		})
	}
	err = h.aiDao.SaveKnowledgePoints(ctx, knowledgePointsModel)
	if err != nil {
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("保存知识点失败: %v", err)
		return
	}
	for _, kp := range knowledgePointsModel {
		_ = h.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "knowledge",
			TargetID:     kp.ID,
			ProcessType:  "embedding",
			ModelName:    h.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		h.aiService.GetQueue().Publish(queue.AITask{
			TaskID: logItem.TaskID,
		})
	}

	h.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
	log.Printf("知识抽取完成，笔记ID: %d, 抽取知识点数: %d", logItem.TargetID, len(knowledgePoints))
}

// handleEmbedding 处理向量生成任务
func (h *Handler) handleEmbedding(ctx context.Context, logItem *model.AiProcessLog) {
	var text string

	switch logItem.TargetType {
	case "note":
		note, err := h.noteDao.GetNoteByID(ctx, logItem.TargetID)
		if err != nil {
			log.Printf("获取笔记失败: %v", err)
			h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
			return
		}
		text = note.ContentMd

	case "knowledge":
		kp, err := h.aiDao.GetKnowledgePointByID(ctx, logItem.TargetID)
		if err != nil {
			log.Printf("获取知识点失败: %v", err)
			h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
			return
		}
		text = kp.Name + "\n" + kp.Description

	default:
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		return
	}

	embedding, err := h.aiClient.GetEmbedding(ctx, text)
	if err != nil {
		log.Printf("获取向量失败: %v", err)
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		return
	}

	embeddingModel := &model.EmbeddingCustom{
		TargetType: logItem.TargetType,
		TargetID:   logItem.TargetID,
		Vector:     embedding,
		ModelName:  h.aiClient.GetEmbeddingModelName(),
	}

	err = h.aiDao.SaveEmbedding(ctx, embeddingModel)
	if err != nil {
		log.Printf("保存向量失败: %v", err)
		h.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		return
	}
	h.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
}
