package handler

import (
	"strconv"
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 获取笔记AI处理状态
// @Description 获取笔记的AI处理状态，包括笔记本身和衍生知识点的处理状态
// @Tags AI处理
// @Security BearerAuth
// @Param id path int64 true "笔记ID"
// @Success 200 {object} response.Response{data=dto.GetNoteAIStatusResp} "获取AI处理状态成功"
// @Router /api/v1/ai/note/{id}/status [get]
func HandlerGetNoteAIStatus(aiService *service.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 note_id
		noteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		// 查询 AI 状态
		resp, err := aiService.GetNoteAIStatus(c.Request.Context(), noteID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		// 返回状态
		response.Success(c, resp)
	}
}

// @Summary 提交笔记AI处理任务
// @Description 提交笔记的AI处理任务，支持多种处理类型（extract/embedding等）
// @Tags AI处理
// @Security BearerAuth
// @Produce json
// @Param id path int64 true "笔记ID"
// @Success 200 {object} response.Response{data=dto.AIProcessResp} "提交AI处理任务成功"
// @Router /api/v1/ai/note/{id} [post]
func HandlerProcessNoteAI(noteService *service.NoteService, aiService *service.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户 ID
		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}
		userID := userIDAny.(int64)

		// 获取 note_id
		noteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		// 校验笔记归属
		if !noteService.CheckNoteOwnership(c.Request.Context(), userID, noteID) {
			response.Fail(c, errors.Forbidden)
			return
		}

		// 提交 AI 任务
		taskID, err := aiService.SubmitTask(c.Request.Context(), noteID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		// 返回 task_id
		response.Success(c, dto.AIProcessResp{
			TaskID: taskID,
			Status: "processing",
		})
	}
}
