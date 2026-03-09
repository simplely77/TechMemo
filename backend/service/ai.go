package service

import (
	"context"
	"fmt"
	"techmemo/backend/ai/queue"
	"techmemo/backend/config"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"

	"github.com/google/uuid"
)

type AIService struct {
	aiDao *dao.AIDao
	queue queue.Queue
}

func (a *AIService) GetNoteAIStatus(ctx context.Context, noteID int64) (dto.GetNoteAIStatusResp, error) {
	logs, err := a.aiDao.GetLogsByNoteID(ctx, noteID)
	if err != nil {
		return dto.GetNoteAIStatusResp{}, err
	}

	if len(logs) == 0 {
		return dto.GetNoteAIStatusResp{
			NoteID: noteID,
			Status: "not_started",
		}, nil
	}

	progress := make(map[string]string)
	overallStatus := "completed" // 默认最优状态

	for _, log := range logs {
		progress[log.ProcessType] = log.Status

		switch log.Status {
		case "failed":
			overallStatus = "failed" // 最差状态覆盖
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

func (a *AIService) SetQueue(queue queue.Queue) {
	a.queue = queue
}

func (a *AIService) SubmitTask(
	ctx context.Context,
	noteID int64,
) (string, error) {
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
		return "", err
	}

	return taskID, a.queue.Publish(queue.AITask{TaskID: taskID})
}

func generateTaskID(id int64) string {
	return fmt.Sprintf(
		"ai:%d:%s",
		id,
		uuid.NewString(),
	)
}

func NewAIService(aiDao *dao.AIDao) *AIService {
	return &AIService{aiDao: aiDao}
}
