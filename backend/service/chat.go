package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"slices"
	"sort"
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

// ragChatSystemPrompt 置于 RAG 消息最前；RAG 正文先经 fitRAGContextWithinBudget 再与此前缀合并。
const ragChatSystemPrompt = `你是 TechMemo 的智能问答助手，帮助用户理解和运用其个人技术笔记库中的内容。

【上下文里各类内容的含义】
- 以「[笔记]」开头：用户原始笔记。第一行是标题，后面是正文（Markdown）。
- 以「[知识点]」开头：从笔记中自动抽取的结构化要点；第一行是知识点名称，后面是简短说明。
- 角色为 user 或 assistant、且没有上述前缀：本会话近期对话，用于理解用户意图与承接话题。

【回答风格】
- 精炼：先给结论或直接答案，少铺垫、不重复材料原文。
- 复杂内容用简短分点，每点尽量一句话。
- 材料不足以回答、或问题需要更多背景时，明确说明，并引导用户补充关键信息（例如具体场景、报错、相关笔记主题）。`

// ragSystemContentWithStats 在系统提示中附带全库真实统计，避免纯 RAG 只看见检索片段时误把「片段条数」当成「用户共有多少条」。
func (s *ChatService) ragSystemContentWithStats(ctx context.Context, userID int64) string {
	n, errN := s.noteDao.CountNotesByUid(ctx, userID)
	k, errK := s.kpDao.CountKnowledgePointsByUid(ctx, userID)
	if errN != nil {
		log.Printf("chat rag: CountNotesByUid: %v", errN)
	}
	if errK != nil {
		log.Printf("chat rag: CountKnowledgePointsByUid: %v", errK)
	}
	if errN != nil || errK != nil {
		return ragChatSystemPrompt + `

【说明】下方以「[笔记]」「[知识点]」出现的内容，仅是与当前问题相关的**部分检索结果**，不是全库。若用户询问「我有多少条笔记/知识点」等**总量**问题，当前无法从本回复中给出准确数字，请直接说明情况，并建议用户到笔记列表或知识点页查看；不要根据下方片段的条数猜测总量。`
	}
	return ragChatSystemPrompt + fmt.Sprintf(`

【知识库统计（数据库为准）】当前账号下共有 %d 条有效笔记（不含已删除）、%d 个知识点。

【与检索材料的关系】下方 [笔记]、[知识点] 块只包含与本次问题相关的一部分内容，**其条数不等于**全库规模。当用户问「我有多少笔记/知识点」等**整体数量**时，必须以本段【知识库统计】中的数字回答。当用户问「某主题/标签下有多少」等**带条件**的数量时，若检索片段不能覆盖全部，应说明这是基于已给出的材料、可能不完整，并建议用搜索或列表筛选进一步确认。`, n, k)
}

const (
	ragMaxContextTokens   = 4000
	ragMaxDialogueTokens  = 2000
	ragRecentMessageCount = 12
	// 单次多取若干条候选，在内存中按 RRF 分数与 token 预算选入，避免为降预算反复调低 topK 重查库
	ragRetrievalTopK = 20
)

// kbRAGSlot 将检索结果展开为可排序候选（Distance 越小 = RRF 融合后越相关）
type kbRAGSlot struct {
	dist  float64
	tie   int64
	tok   int32
	msg   aiclient.ChatMessage
}

// buildRAGContext 构建 RAG 上下文
func (s *ChatService) buildRAGContext(ctx context.Context, userQuery string, sessionID, userID int64) ([]aiclient.ChatMessage, error) {
	queryVector, err := s.aiClient.GetEmbedding(ctx, userQuery)
	if err != nil {
		return nil, err
	}

	recentMsgs, err := s.chatDao.GetRecentMessages(ctx, sessionID, ragRecentMessageCount)
	if err != nil {
		return nil, err
	}
	slices.Reverse(recentMsgs)

	// 对话占固定上限，从最早一条开始裁，保护近期多轮与本轮用户问
	recentMsgs = trimDialogueToTokenBudget(s, recentMsgs, ragMaxDialogueTokens)
	dialogueTokens := dialogueTokenTotal(s, recentMsgs)

	noteResults, err := s.searchDao.HybridSearchEmbeddings(ctx, queryVector, userID, userQuery, "note", ragRetrievalTopK)
	if err != nil {
		return nil, err
	}
	kpResults, err := s.searchDao.HybridSearchEmbeddings(ctx, queryVector, userID, userQuery, "knowledge", ragRetrievalTopK)
	if err != nil {
		return nil, err
	}

	slots, err := s.collectKnowledgeSlots(ctx, noteResults, kpResults)
	if err != nil {
		return nil, err
	}
	sortKBSlotsByRelevance(slots)
	kbBudget := ragMaxContextTokens - dialogueTokens
	if kbBudget < 0 {
		kbBudget = 0
	}
	kbMsgs := s.pickKnowledgeWithinBudget(slots, kbBudget)
	contextMsgs := s.composeRAGMessages(kbMsgs, recentMsgs)
	contextMsgs = s.shrinkRAGToTokenBudget(contextMsgs, ragMaxContextTokens)

	return append([]aiclient.ChatMessage{
		{Role: "system", Content: s.ragSystemContentWithStats(ctx, userID)},
	}, contextMsgs...), nil
}

func (s *ChatService) collectKnowledgeSlots(ctx context.Context, noteResults, kpResults []dao.SearchResult) ([]kbRAGSlot, error) {
	noteIDs := make([]int64, 0, len(noteResults))
	for i := range noteResults {
		noteIDs = append(noteIDs, noteResults[i].TargetID)
	}
	notes, err := s.noteDao.GetNotesByIDs(ctx, noteIDs)
	if err != nil {
		return nil, err
	}
	noteByID := make(map[int64]*model.Note, len(notes))
	for i := range notes {
		noteByID[notes[i].ID] = notes[i]
	}

	kpIDs := make([]int64, 0, len(kpResults))
	for i := range kpResults {
		kpIDs = append(kpIDs, kpResults[i].TargetID)
	}
	kps, err := s.kpDao.GetKnowledgePointsByIDs(ctx, kpIDs)
	if err != nil {
		return nil, err
	}
	kpByID := make(map[int64]*model.KnowledgePoint, len(kps))
	for i := range kps {
		kpByID[kps[i].ID] = kps[i]
	}

	var out []kbRAGSlot
	for _, nr := range noteResults {
		note := noteByID[nr.TargetID]
		if note == nil {
			continue
		}
		content := fmt.Sprintf("[笔记] %s\n%s", note.Title, note.ContentMd)
		out = append(out, kbRAGSlot{
			dist:  nr.Distance,
			tie:   note.ID,
			tok:   s.estimateTokens(content),
			msg:   aiclient.ChatMessage{Role: "system", Content: content},
		})
	}
	for _, kr := range kpResults {
		kp := kpByID[kr.TargetID]
		if kp == nil {
			continue
		}
		content := fmt.Sprintf("[知识点] %s\n%s", kp.Name, kp.Description)
		out = append(out, kbRAGSlot{
			dist:  kr.Distance,
			tie:   kp.ID,
			tok:   s.estimateTokens(content),
			msg:   aiclient.ChatMessage{Role: "system", Content: content},
		})
	}
	return out, nil
}

func sortKBSlotsByRelevance(slots []kbRAGSlot) {
	sort.Slice(slots, func(i, j int) bool {
		if slots[i].dist != slots[j].dist {
			return slots[i].dist < slots[j].dist
		}
		return slots[i].tie < slots[j].tie
	})
}

func (s *ChatService) pickKnowledgeWithinBudget(slots []kbRAGSlot, budget int32) []aiclient.ChatMessage {
	var out []aiclient.ChatMessage
	rem := budget
	for i := range slots {
		if slots[i].tok <= rem {
			out = append(out, slots[i].msg)
			rem -= slots[i].tok
		}
	}
	return out
}

func (s *ChatService) composeRAGMessages(kb []aiclient.ChatMessage, recent []*model.ChatMessage) []aiclient.ChatMessage {
	out := make([]aiclient.ChatMessage, 0, len(kb)+len(recent))
	out = append(out, kb...)
	for _, m := range recent {
		out = append(out, aiclient.ChatMessage{Role: m.Role, Content: m.Content})
	}
	return out
}

// trimDialogueToTokenBudget 自最早消息起丢弃，使对话块不超过 max。至少保留 1 条，即使仍略超（极端单条极长时由 shrinkRAG 再协调）。
func trimDialogueToTokenBudget(s *ChatService, recent []*model.ChatMessage, max int32) []*model.ChatMessage {
	if max <= 0 {
		return recent
	}
	for len(recent) > 0 {
		if dialogueTokenTotal(s, recent) <= max {
			return recent
		}
		if len(recent) <= 1 {
			return recent
		}
		recent = recent[1:]
	}
	return recent
}

func dialogueTokenTotal(s *ChatService, recent []*model.ChatMessage) int32 {
	var t int32
	for _, m := range recent {
		t += s.estimateTokens(m.Content)
	}
	return t
}

// shrinkRAGToTokenBudget 总长度仍超限时：先去掉排在末尾的知识 system 条（在按相关性从好到坏插入的前提下，即先丢相对最不相关），再丢对话旧轮。
func (s *ChatService) shrinkRAGToTokenBudget(msgs []aiclient.ChatMessage, max int32) []aiclient.ChatMessage {
	for s.calculateTotalTokens(msgs) > max && len(msgs) > 0 {
		n := ragLeadingSystemCount(msgs)
		if n > 0 {
			// 去掉本段「知识」里列在最后的一条，即已选入结果中相关度最弱的一条
			msgs = append(msgs[:n-1], msgs[n:]...)
			continue
		}
		if len(msgs) <= 1 {
			break
		}
		msgs = msgs[1:]
	}
	return msgs
}

func ragLeadingSystemCount(msgs []aiclient.ChatMessage) int {
	i := 0
	for i < len(msgs) && msgs[i].Role == "system" {
		i++
	}
	return i
}

// calculateTotalTokens 计算总 token 数
func (s *ChatService) calculateTotalTokens(msgs []aiclient.ChatMessage) int32 {
	var total int32
	for _, msg := range msgs {
		total += s.estimateTokens(msg.Content)
	}
	return total
}
