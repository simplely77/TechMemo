package worker

import (
	"context"
	"techmemo/backend/ai/queue"
	"techmemo/backend/service"
)

type Handler struct {
	aiService *service.AIService
}

func NewHandler(aiService *service.AIService) *Handler {
	return &Handler{aiService: aiService}
}

func (h *Handler) Handler(ctx context.Context, task queue.AITask) {
	h.aiService.ProcessTask(ctx, task)
}
