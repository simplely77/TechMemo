package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/common/errors"
	"techmemo/backend/dao"
	"techmemo/backend/model"
)

type ChatService struct {
	chatDao   *dao.ChatDao
	searchDao *dao.SearchDao
	noteDao   *dao.NoteDao
	kpDao     *dao.KnowledgePointDao
	aiClient  aiclient.AIClient
}

var titleNewSessionSeq = regexp.MustCompile(`^新会话\s*(\d+)\s*$`)

func nextDefaultSessionTitle(sessions []*model.ChatSession) string {
	maxN := 0
	for _, sess := range sessions {
		sub := titleNewSessionSeq.FindStringSubmatch(strings.TrimSpace(sess.Title))
		if len(sub) == 2 {
			n, err := strconv.Atoi(sub[1])
			if err == nil && n > maxN {
				maxN = n
			}
		}
	}
	return fmt.Sprintf("新会话 %d", maxN+1)
}

// CreateSession 创建聊天会话；explicitTitle 非空时作为标题，否则生成「新会话 N」递增名
func (s *ChatService) CreateSession(ctx context.Context, userID int64, explicitTitle string) (*model.ChatSession, error) {
	title := strings.TrimSpace(explicitTitle)
	if title != "" {
		if utf8.RuneCountInString(title) > 200 {
			runes := []rune(title)
			title = string(runes[:200])
		}
		return s.chatDao.CreateSession(ctx, dao.CreateSessionParams{
			UserID: userID,
			Title:  title,
		})
	}
	sessions, _, err := s.chatDao.GetSessionsByUserID(ctx, userID, 2000, 0)
	if err != nil {
		return nil, err
	}
	title = nextDefaultSessionTitle(sessions)
	return s.chatDao.CreateSession(ctx, dao.CreateSessionParams{
		UserID: userID,
		Title:  title,
	})
}

// UpdateSessionTitle 更新会话标题（校验归属）
func (s *ChatService) UpdateSessionTitle(ctx context.Context, sessionID, userID int64, title string) (*model.ChatSession, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.InvalidParam
	}
	if utf8.RuneCountInString(title) > 200 {
		return nil, errors.InvalidParam
	}
	ok, err := s.chatDao.CheckSessionBelongsToUser(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Forbidden
	}
	if err := s.chatDao.UpdateSessionTitle(ctx, sessionID, title); err != nil {
		return nil, err
	}
	return s.chatDao.GetSessionByID(ctx, sessionID)
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

func NewChatService(
	chatDao *dao.ChatDao,
	searchDao *dao.SearchDao,
	noteDao *dao.NoteDao,
	kpDao *dao.KnowledgePointDao,
	aiClient aiclient.AIClient,
) *ChatService {
	return &ChatService{
		chatDao:   chatDao,
		searchDao: searchDao,
		noteDao:   noteDao,
		kpDao:     kpDao,
		aiClient:  aiClient,
	}
}

// SendMessage 发送消息并获取 AI 回复
func (s *ChatService) SendMessage(ctx context.Context, sessionID, userID int64, content string) (*model.ChatMessage, error) {
	ok, err := s.chatDao.CheckSessionBelongsToUser(ctx, sessionID, userID)
	if err != nil || !ok {
		return nil, errors.Forbidden
	}

	if _, err := s.chatDao.CreateMessage(ctx, dao.CreateMessageParams{
		SessionID:  sessionID,
		UserID:     userID,
		Role:       "user",
		Content:    content,
		TokenCount: s.estimateTokens(content),
	}); err != nil {
		return nil, errors.InternalErr
	}

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

	return aiMsg, nil
}

func (s *ChatService) SendMessageStream(ctx context.Context, sessionID int64, userID int64, content string, onDelta func(delta string) bool) error {
	ok, err := s.chatDao.CheckSessionBelongsToUser(ctx, sessionID, userID)
	if err != nil || !ok {
		return errors.Forbidden
	}

	if _, err := s.chatDao.CreateMessage(ctx, dao.CreateMessageParams{
		SessionID:  sessionID,
		UserID:     userID,
		Role:       "user",
		Content:    content,
		TokenCount: s.estimateTokens(content),
	}); err != nil {
		log.Println(err)
		return errors.InternalErr
	}

	ragMessages, err := s.buildRAGContext(ctx, content, sessionID, userID)
	if err != nil {
		log.Println(err)
		return errors.InternalErr
	}

	var fullReply strings.Builder

	err = s.aiClient.ChatStream(ctx, ragMessages, func(delta string) bool {

		fullReply.WriteString(delta)

		return onDelta(delta) // 由上层决定是否继续
	})
	if err != nil {
		log.Println(err)
		return errors.InternalErr
	}

	reply := fullReply.String()
	log.Println(reply)
	if reply != "" {
		if _, err := s.chatDao.CreateMessage(ctx, dao.CreateMessageParams{
			SessionID:  sessionID,
			UserID:     userID,
			Role:       "assistant",
			Content:    reply,
			TokenCount: s.estimateTokens(reply),
		}); err != nil {
			log.Println(err)
			return errors.InternalErr
		}
	}
	return nil
}

// ragChatSystemPrompt 置于 RAG 消息最前；须在 truncate 之后 prepend，否则会被 truncateKnowledgeContext 丢掉。
const ragChatSystemPrompt = `你是 TechMemo 的智能问答助手，帮助用户理解和运用其个人技术笔记库中的内容。

【上下文里各类内容的含义】
- 以「[笔记]」开头：用户原始笔记。第一行是标题，后面是正文（Markdown）。
- 以「[知识点]」开头：从笔记中自动抽取的结构化要点；第一行是知识点名称，后面是简短说明。
- 角色为 user 或 assistant、且没有上述前缀：本会话近期对话，用于理解用户意图与承接话题。

【回答风格】
- 精炼：先给结论或直接答案，少铺垫、不重复材料原文。
- 复杂内容用简短分点，每点尽量一句话。
- 材料不足以回答、或问题需要更多背景时，明确说明，并引导用户补充关键信息（例如具体场景、报错、相关笔记主题）。`

// buildRAGContext 构建 RAG 上下文
func (s *ChatService) buildRAGContext(ctx context.Context, userQuery string, sessionID, userID int64) ([]aiclient.ChatMessage, error) {
	const (
		MaxTokens            = 4000
		RecentMessageCount   = 12
		InitialTopK          = 5
		MinTopK              = 2
	)

	// 向量化用户查询
	queryVector, err := s.aiClient.GetEmbedding(ctx, userQuery)
	if err != nil {
		return nil, err
	}

	// 获取本会话最近消息（DAO 按时间倒序；大模型需要按对话时间正序）
	recentMsgs, err := s.chatDao.GetRecentMessages(ctx, sessionID, RecentMessageCount)
	if err != nil {
		return nil, err
	}
	slices.Reverse(recentMsgs)

	// 语义搜索（笔记与知识点；聊天消息不再建向量）
	topK := InitialTopK
	noteResults, _ := s.searchDao.SearchEmbeddingsByVector(ctx, queryVector, "note", userID, topK)
	kpResults, _ := s.searchDao.SearchEmbeddingsByVector(ctx, queryVector, "knowledge", userID, topK)

	// 合并上下文（recentMsgs 已含本轮用户消息，见 SendMessage 先落库再 buildRAGContext）
	contextMsgs := s.mergeContext(ctx, recentMsgs, noteResults, kpResults)

	// 动态裁剪
	totalTokens := s.calculateTotalTokens(contextMsgs)
	for totalTokens > MaxTokens && topK > MinTopK {
		topK--
		noteResults, _ = s.searchDao.SearchEmbeddingsByVector(ctx, queryVector, "note", userID, topK)
		kpResults, _ = s.searchDao.SearchEmbeddingsByVector(ctx, queryVector, "knowledge", userID, topK)
		contextMsgs = s.mergeContext(ctx, recentMsgs, noteResults, kpResults)
		totalTokens = s.calculateTotalTokens(contextMsgs)
	}

	// 如果仍超限，截断知识库
	if totalTokens > MaxTokens {
		contextMsgs = s.truncateKnowledgeContext(contextMsgs, MaxTokens)
	}

	return append([]aiclient.ChatMessage{
		{Role: "system", Content: ragChatSystemPrompt},
	}, contextMsgs...), nil
}

// mergeContext 合并上下文。recentMsgs 须为时间正序，且已包含本轮用户消息（与 SendMessage 中先 CreateMessage 再 buildRAGContext 一致）。
func (s *ChatService) mergeContext(ctx context.Context, recentMsgs []*model.ChatMessage, noteResults, kpResults []dao.SearchResult) []aiclient.ChatMessage {
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

	// 最近消息（user/assistant 交替，时间正序）
	for _, msg := range recentMsgs {
		result = append(result, aiclient.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

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
