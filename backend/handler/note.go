package handler

import (
	"strconv"
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 获取笔记列表
// @Description 获取笔记列表
// @Tags 笔记
// @Security BearerAuth
// @Param category_id query int64 false "分类ID"
// @Param tag_ids query []int64 false "标签ID列表"
// @Param keyword query string false "搜索关键词"
// @Param note_type query string false "笔记类型"
// @Param page query int64 false "页码（默认 1）"
// @Param page_size query int64 false "每页数量（默认 20）"
// @Param sort query string false "排序字段（默认 created_at_desc）"
// @Success 200 {object} response.Response{data=dto.GetNotesResp} "获取笔记列表成功"
// @Router /api/v1/notes [get]
func HandlerGetNotes(noteService *service.NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.GetNotesReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID, ok := userIDAny.(int64)
		if !ok {
			response.Fail(c, errors.InternalErr)
			return
		}
		resp, err := noteService.GetNotes(c.Request.Context(), &req, userID)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 创建笔记
// @Description 创建笔记
// @Tags 笔记
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body dto.CreateNoteReq true "创建笔记参数"
// @Success 200 {object} response.Response{data=dto.CreateNoteResp} "创建笔记成功"
// @Router /api/v1/notes [post]
func HandlerCreateNote(noteService *service.NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateNoteReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID, ok := userIDAny.(int64)
		if !ok {
			response.Fail(c, errors.InternalErr)
			return
		}

		resp, err := noteService.CreateNoteWithTags(c.Request.Context(), &req, userID)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 获取笔记详情
// @Description 获取笔记详情
// @Tags 笔记
// @Security BearerAuth
// @Param id path int64 false "笔记ID"
// @Success 200 {object} response.Response{data=dto.GetNoteResp} "获取笔记详情成功"
// @Router /api/v1/notes/{id} [get]
func HandlerGetNote(noteService *service.NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		resp, err := noteService.GetNote(c.Request.Context(), id)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}
		response.Success(c, resp)
	}
}

// @Summary 更新笔记
// @Description 更新笔记的标题、内容和分类（不修改标签）
// @Tags 笔记
// @Security BearerAuth
// @Param id path int64 true "笔记ID"
// @Param body body dto.UpdateNoteReq true "更新笔记请求体"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/notes/{id} [put]
func HandlerUpdateNote(noteService *service.NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}
		var req dto.UpdateNoteReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		err = noteService.UpdateNote(c.Request.Context(), id, &req)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		response.Success(c, nil)
	}
}

// @Summary 更新笔记标签
// @Description 更新笔记的标签列表（覆盖原有标签）
// @Tags 笔记
// @Security BearerAuth
// @Param id path int64 true "笔记ID"
// @Param body body dto.UpdateNoteTagsReq true "更新笔记标签请求体"
// @Success 200 {object} response.Response "更新成功"
// @Router /api/v1/notes/{id}/tags [put]
func HandlerUpdateNoteTags(noteService *service.NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}
		var req dto.UpdateNoteTagsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}
		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID, ok := userIDAny.(int64)
		if !ok {
			response.Fail(c, errors.InternalErr)
			return
		}

		err = noteService.UpdateNoteTags(c.Request.Context(), userID, id, &req)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		response.Success(c, nil)
	}
}

// @Summary 删除笔记
// @Description 软删除指定笔记（将笔记状态标记为 deleted，不物理删除数据）
// @Tags 笔记
// @Security BearerAuth
// @Param id path int64 true "笔记ID"
// @Success 200 {object} response.Response "删除笔记成功"
// @Router /api/v1/notes/{id} [delete]
func HandlerDeleteNote(noteService *service.NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		err = noteService.DeleteNote(c.Request.Context(), id)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		response.Success(c, nil)
	}
}

// @Summary 获取笔记版本历史
// @Description 获取指定笔记的历史版本列表，支持按创建时间排序
// @Tags 笔记
// @Security BearerAuth
// @Param id path int64 true "笔记ID"
// @Param sort query string false "排序方式（created_at_desc / created_at_asc），默认 created_at_desc"
// @Success 200 {object} response.Response{data=dto.GetNoteVersionsResp} "获取笔记版本历史成功"
// @Router /api/v1/notes/{id}/versions [get]
func HandlerGetNoteVersions(noteService *service.NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		sort := c.Query("sort")
		if sort == "" {
			sort = "created_at_desc"
		}

		resp, err := noteService.GetNoteVersions(c.Request.Context(), id, sort)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 恢复笔记历史版本
// @Description 将笔记内容恢复到指定的历史版本，并生成一条新的版本记录
// @Tags 笔记
// @Security BearerAuth
// @Param id path int64 true "笔记ID"
// @Param version_id path int64 true "历史版本ID"
// @Success 200 {object} response.Response "恢复笔记版本成功"
// @Router /api/v1/notes/{id}/versions/{version_id}/restore [post]
func HandlerRestoreNote(noteService *service.NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}
		versionIDStr := c.Param("version_id")
		versionID, err := strconv.ParseInt(versionIDStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}
		err = noteService.RestoreNote(c.Request.Context(), id, versionID)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		response.Success(c, nil)
	}
}
