package service

import (
	"context"
	"fmt"
	"unicode"

	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/ai/queue"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
	"techmemo/backend/model"
)

type ChatService struct {
	chatDao   *dao.ChatDao
	searchDao *dao.SearchDao
	aiDao     *dao.AIDao
	noteDao   *dao.NoteDao
	kpDao     *dao.KnowledgePointDao
	aiClient  aiclient.AIClient
	queue     queue.Queue
}

// CreateSession 创建聊天会话
func (s *ChatService) CreateSession(ctx context.Context, userID int64) (*model.ChatSession, error) {
	return s.chatDao.CreateSession(ctx, dao.CreateSessionParams{
		UserID: userID,
		Title:  "新对话",
	})
}

// GetSessions 获取用户的会话列表
func (s *ChatService) GetSessions(ctx context.Context, userID int64, page, pageSize int) ([]*model.ChatSession, int64, error) {
	offset := (page - 1) * pageSize
	return s.chatDao.GetSessionsByUserID(ctx, userID, pageSize, offset)
}

// DeleteSession 删除会话
func (s *ChatService) DeleteSession(ctx context.Context, sessionID, userID int64) error {
	ok, err := s.chatDao.CheckSessionBelongsToUser(ctx, sessionID, userID)
	if err != nil || !ok {
		return errors.Forbidden
	}
	return s.chatDao.DeleteSession(ctx, sessionID)
}

// GetMessages 获取会话消息历史
func (s *ChatService) GetMessages(ctx context.Context, sessionID, userID int64, page, pageSize int) ([]*model.ChatMessage, int64, error) {
	ok, err := s.chatDao.CheckSessionBelongsToUser(ctx, sessionID, userID)
	if err != nil || !ok {
		return nil, 0, errors.Forbidden
	}
	offset := (page - 1) * pageSize
	return s.chatDao.GetMessagesBySessionID(ctx, sessionID, pageSize, offset)
}

// estimateTokens 估算消息的 token 数量
func (s *ChatService) estimateTokens(text string) int32 {
	var tokens float64
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			tokens += 0.6
		} else if r < 128 {
			tokens += 0.3
		} else {
			tokens += 0.5
		}
	}
	return int32(tokens)
}

// enqueueMessageEmbedding 将消息向量化任务入队
func (s *ChatService) enqueueMessageEmbedding(ctx context.Context, messageID int64) error {
	taskID := fmt.Sprintf("chat_msg_%d", messageID)
	err := s.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
		SourceNoteID: 0,
		TaskID:       taskID,
		TargetType:   "chat_message",
		TargetID:     messageID,
		ProcessType:  "embedding",
		ModelName:    s.aiClient.GetEmbeddingModelName(),
		Status:       "pending",
	})
	if err != nil {
		return err
	}
	return s.queue.Publish(queue.AITask{TaskID: taskID})
}

func NewChatService(
	chatDao *dao.ChatDao,
	searchDao *dao.SearchDao,
	aiDao *dao.AIDao,
	noteDao *dao.NoteDao,
	kpDao *dao.KnowledgePointDao,
	aiClient aiclient.AIClient,
	q queue.Queue,
) *ChatService {
	return &ChatService{
		chatDao:   chatDao,
		searchDao: searchDao,
		aiDao:     aiDao,
		noteDao:   noteDao,
		kpDao:     kpDao,
		aiClient:  aiClient,
		queue:     q,
	}
}

// SendMessage 发送消息并获取 AI 回复
func (s *ChatService) SendMessage(ctx context.Context, sessionID, userID int64, content string) (*model.ChatMessage, error) {
	ok, err := s.chatDao.CheckSessionBelongsToUser(ctx, sessionID, userID)
	if err != nil || !ok {
		return nil, errors.Forbidden
	}

	// 保存用户消息
	userMsg, err := s.chatDao.CreateMessage(ctx, dao.CreateMessageParams{
		SessionID:  sessionID,
		UserID:     userID,
		Role:       "user",
		Content:    content,
		TokenCount: s.estimateTokens(content),
	})
	if err != nil {
		return nil, errors.InternalErr
	}

	// 异步向量化用户消息
	go s.enqueueMessageEmbedding(ctx, userMsg.ID)

	// 构建 RAG 上下文
	ragMessages, err := s.buildRAGContext(ctx, content, sessionID, userID)
	if err != nil {
		return nil, errors.InternalErr
	}

	// 调用 LLM 生成回复
	reply, err := s.aiClient.ChatWithContext(ctx, ragMessages)
	if err != nil {
		return nil, errors.InternalErr
	}

	// 保存 AI 回复
	aiMsg, err := s.chatDao.CreateMessage(ctx, dao.CreateMessageParams{
		SessionID:  sessionID,
		UserID:     userID,
		Role:       "assistant",
		Content:    reply,
		TokenCount: s.estimateTokens(reply),
	})
	if err != nil {
		return nil, errors.InternalErr
	}

	// 异步向量化 AI 回复
	go s.enqueueMessageEmbedding(ctx, aiMsg.ID)

	return aiMsg, nil
}

// buildRAGContext 构建 RAG 上下文
func (s *ChatService) buildRAGContext(ctx context.Context, userQuery string, sessionID, userID int64) ([]aiclient.ChatMessage, error) {
	const (
		MaxTokens          = 4000
		RecentMessageCount = 6
		InitialTopK        = 5
		MinTopK            = 2
		ChatThreshold      = 0.3
		NoteThreshold      = 0.4
		KnowledgeThreshold = 0.4
	)

	// 向量化用户查询
	queryVector, err := s.aiClient.GetEmbedding(ctx, userQuery)
	if err != nil {
		return nil, err
	}

	// 获取本会话最近消息
	recentMsgs, err := s.chatDao.GetRecentMessages(ctx, sessionID, RecentMessageCount)
	if err != nil {
		return nil, err
	}

	// 语义搜索
	topK := InitialTopK
	chatResults, _ := s.searchDao.SearchEmbeddingsByVector(ctx, queryVector, "chat_message", userID, topK, ChatThreshold)
	noteResults, _ := s.searchDao.SearchEmbeddingsByVector(ctx, queryVector, "note", userID, topK, NoteThreshold)
	kpResults, _ := s.searchDao.SearchEmbeddingsByVector(ctx, queryVector, "knowledge", userID, topK, KnowledgeThreshold)

	// 合并上下文
	contextMsgs := s.mergeContext(ctx, recentMsgs, chatResults, noteResults, kpResults, userQuery)

	// 动态裁剪
	totalTokens := s.calculateTotalTokens(contextMsgs)
	for totalTokens > MaxTokens && topK > MinTopK {
		topK--
		chatResults, _ = s.searchDao.SearchEmbeddingsByVector(ctx, queryVector, "chat_message", userID, topK, ChatThreshold)
		contextMsgs = s.mergeContext(ctx, recentMsgs, chatResults, noteResults, kpResults, userQuery)
		totalTokens = s.calculateTotalTokens(contextMsgs)
	}

	// 如果仍超限，截断知识库
	if totalTokens > MaxTokens {
		contextMsgs = s.truncateKnowledgeContext(contextMsgs, MaxTokens)
	}

	return contextMsgs, nil
}

// mergeContext 合并上下文
func (s *ChatService) mergeContext(ctx context.Context, recentMsgs []*model.ChatMessage, chatResults, noteResults, kpResults []dao.SearchResult, userQuery string) []aiclient.ChatMessage {
	var result []aiclient.ChatMessage

	// 知识库上下文
	for _, nr := range noteResults {
		note, _ := s.noteDao.GetNoteByID(ctx, nr.TargetID)
		if note != nil {
			result = append(result, aiclient.ChatMessage{
				Role:    "system",
				Content: fmt.Sprintf("[笔记] %s\n%s", note.Title, note.ContentMd),
			})
		}
	}

	for _, kr := range kpResults {
		kp, _ := s.kpDao.GetKnowledgePointByID(ctx, kr.TargetID)
		if kp != nil {
			result = append(result, aiclient.ChatMessage{
				Role:    "system",
				Content: fmt.Sprintf("[知识点] %s\n%s", kp.Name, kp.Description),
			})
		}
	}

	// 语义相关的历史消息（去重最近消息）
	recentIDs := make(map[int64]bool)
	for _, msg := range recentMsgs {
		recentIDs[msg.ID] = true
	}

	for _, cr := range chatResults {
		if !recentIDs[cr.TargetID] {
			msg, _ := s.chatDao.GetMessageByID(ctx, cr.TargetID)
			if msg != nil {
				result = append(result, aiclient.ChatMessage{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}
		}
	}

	// 最近消息
	for _, msg := range recentMsgs {
		result = append(result, aiclient.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 当前查询
	result = append(result, aiclient.ChatMessage{
		Role:    "user",
		Content: userQuery,
	})

	return result
}

// calculateTotalTokens 计算总 token 数
func (s *ChatService) calculateTotalTokens(msgs []aiclient.ChatMessage) int32 {
	var total int32
	for _, msg := range msgs {
		total += s.estimateTokens(msg.Content)
	}
	return total
}

// truncateKnowledgeContext 截断知识库上下文
func (s *ChatService) truncateKnowledgeContext(msgs []aiclient.ChatMessage, maxTokens int32) []aiclient.ChatMessage {
	var result []aiclient.ChatMessage
	var tokens int32

	// 保留最后的用户消息和最近的对话消息
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		msgTokens := int32(len([]rune(msg.Content)) / 2)

		if msg.Role == "system" {
			continue
		}

		if tokens+msgTokens <= maxTokens {
			result = append([]aiclient.ChatMessage{msg}, result...)
			tokens += msgTokens
		}
	}

	return result
}
