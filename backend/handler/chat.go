package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 创建聊天会话
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.CreateSessionResp}
// @Router /api/v1/chat/sessions [post]
func HandlerCreateSession(svc *service.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		session, err := svc.CreateSession(c.Request.Context(), userID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}
		response.Success(c, dto.CreateSessionResp{
			ID:        session.ID,
			Title:     session.Title,
			CreatedAt: session.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: session.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

// @Summary 获取会话列表
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=dto.ChatSessionListResp}
// @Router /api/v1/chat/sessions [get]
func HandlerGetSessions(svc *service.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		page := 1
		pageSize := 10
		if p := c.Query("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}
		if ps := c.Query("page_size"); ps != "" {
			if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
				pageSize = v
			}
		}

		sessions, total, err := svc.GetSessions(c.Request.Context(), userID, page, pageSize)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		resp := dto.ChatSessionListResp{
			Sessions: make([]dto.CreateSessionResp, len(sessions)),
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}
		for i, s := range sessions {
			resp.Sessions[i] = dto.CreateSessionResp{
				ID:        s.ID,
				Title:     s.Title,
				CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: s.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
		}
		response.Success(c, resp)
	}
}

// @Summary 删除会话
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "会话ID"
// @Success 200 {object} response.Response
// @Router /api/v1/chat/sessions/{id} [delete]
func HandlerDeleteSession(svc *service.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		if err := svc.DeleteSession(c.Request.Context(), sessionID, userID); err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}
		response.Success(c, nil)
	}
}

// @Summary 发送消息
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "会话ID"
// @Param request body dto.SendMessageReq true "消息内容"
// @Success 200 {object} response.Response{data=dto.ChatMessageResp}
// @Router /api/v1/chat/sessions/{id}/messages [post]
func HandlerSendMessage(svc *service.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		var req dto.SendMessageReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		msg, err := svc.SendMessage(c.Request.Context(), sessionID, userID, req.Content)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, dto.ChatMessageResp{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

// @Summary 发送消息(流式返回)
// @Tags 聊天
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param id path int true "会话ID"
// @Param request body dto.SendMessageReq true "消息内容"
// @Success 200 {string} string "SSE stream"
// @Router /api/v1/chat/sessions/{id}/stream [post]
func HandlerSendMessageStream(svc *service.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")

		sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		var req dto.SendMessageReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		// 设置 SSE 头
		c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.String(http.StatusInternalServerError, "Streaming unsupported")
			return
		}

		ctx := c.Request.Context()

		go func() {
			ticker := time.NewTicker(20 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					fmt.Fprintf(c.Writer, ": ping\n\n")
					flusher.Flush()
				}
			}
		}()

		// 用于收集完整回答（后面存数据库）
		var fullContent strings.Builder

		err = svc.SendMessageStream(
			ctx,
			sessionID,
			userID,
			req.Content,
			func(delta string) bool {
				select {
				case <-ctx.Done():
					return false
				default:
				}
				// 累积完整回答
				fullContent.WriteString(delta)

				// SSE 输出
				writeSSE(c.Writer, "", delta)
				flusher.Flush()
				return true
			},
		)

		if err != nil {
			writeSSE(c.Writer, "error", err.Error())
			flusher.Flush()
			return
		}

		// 结束标记
		writeSSE(c.Writer, "done", "[DONE]")
		flusher.Flush()
	}
}

func writeSSE(w io.Writer, event, data string) {
	if event != "" {
		fmt.Fprintf(w, "event: %s\n", event)
	}
	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}

// @Summary 获取会话消息历史
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "会话ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.ChatMessageListResp}
// @Router /api/v1/chat/sessions/{id}/messages [get]
func HandlerGetMessages(svc *service.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			log.Println(err)
			response.Fail(c, errors.InvalidParam)
			return
		}

		page := 1
		pageSize := 20
		if p := c.Query("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}
		if ps := c.Query("page_size"); ps != "" {
			if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
				pageSize = v
			}
		}

		messages, total, err := svc.GetMessages(c.Request.Context(), sessionID, userID, page, pageSize)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		resp := dto.ChatMessageListResp{
			Messages: make([]dto.ChatMessageResp, len(messages)),
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}
		for i, m := range messages {
			resp.Messages[i] = dto.ChatMessageResp{
				ID:        m.ID,
				SessionID: m.SessionID,
				Role:      m.Role,
				Content:   m.Content,
				CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}
		response.Success(c, resp)
	}
}
