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

// @Summary 触发全局思维导图生成
// @Description 异步生成全局思维导图，分析所有笔记顶节点间的关联关系，完成后可通过 GET /mindmap/global 获取结果
// @Tags AI处理
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.AIProcessResp} "任务提交成功"
// @Router /api/v1/ai/mindmap/global [post]
func HandlerGenerateGlobalMindMap(aiService *service.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}
		userID := userIDAny.(int64)

		taskID, err := aiService.SubmitGlobalMindMapTask(c.Request.Context(), userID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, dto.AIProcessResp{
			TaskID: taskID,
			Status: "processing",
		})
	}
}

// @Summary 查询 AI 任务状态
// @Description 通过 taskID 查询任务处理状态，适用于全局思维导图等异步任务
// @Tags AI处理
// @Security BearerAuth
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response{data=dto.GetTaskStatusResp} "查询成功"
// @Router /api/v1/ai/task/{task_id}/status [get]
func HandlerGetTaskStatus(aiService *service.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("task_id")
		if taskID == "" {
			response.Fail(c, errors.InvalidParam)
			return
		}

		resp, err := aiService.GetTaskStatus(c.Request.Context(), taskID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, resp)
	}
}
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
