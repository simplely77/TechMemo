package dao

import (
	"context"
	"errors"
	"techmemo/backend/model"
	"techmemo/backend/query"
	"time"

	"gorm.io/gorm"
)

type ChatDao struct {
	q *query.Query
}

// CreateSession 创建聊天会话
func (c *ChatDao) CreateSession(ctx context.Context, params CreateSessionParams) (*model.ChatSession, error) {
	session := &model.ChatSession{
		UserID:    params.UserID,
		Title:     params.Title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := c.q.ChatSession.WithContext(ctx).Create(session)
	return session, err
}

// GetSessionsByUserID 获取用户的所有会话
func (c *ChatDao) GetSessionsByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.ChatSession, int64, error) {
	q := c.q.ChatSession.WithContext(ctx).Where(c.q.ChatSession.UserID.Eq(userID))

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	sessions, err := q.
		Order(c.q.ChatSession.UpdatedAt.Desc()).
		Limit(limit).
		Offset(offset).
		Find()

	return sessions, total, err
}

// GetSessionByID 根据ID获取会话
func (c *ChatDao) GetSessionByID(ctx context.Context, sessionID int64) (*model.ChatSession, error) {
	session, err := c.q.ChatSession.
		WithContext(ctx).
		Where(c.q.ChatSession.ID.Eq(sessionID)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

// UpdateSessionTitle 更新会话标题
func (c *ChatDao) UpdateSessionTitle(ctx context.Context, sessionID int64, title string) error {
	_, err := c.q.ChatSession.
		WithContext(ctx).
		Where(c.q.ChatSession.ID.Eq(sessionID)).
		Update(c.q.ChatSession.Title, title)
	return err
}

// DeleteSession 删除会话（级联删除消息和 embedding）
func (c *ChatDao) DeleteSession(ctx context.Context, sessionID int64) error {
	// 查出该会话所有消息 ID
	var msgIDs []int64
	if err := c.q.ChatMessage.WithContext(ctx).
		Where(c.q.ChatMessage.SessionID.Eq(sessionID)).
		Pluck(c.q.ChatMessage.ID, &msgIDs); err != nil {
		return err
	}

	// 删除消息的 embedding
	if len(msgIDs) > 0 {
		if _, err := c.q.Embedding.WithContext(ctx).
			Where(c.q.Embedding.TargetType.Eq("chat_message")).
			Where(c.q.Embedding.TargetID.In(msgIDs...)).
			Delete(); err != nil {
			return err
		}
	}

	// 删除消息
	if _, err := c.q.ChatMessage.WithContext(ctx).
		Where(c.q.ChatMessage.SessionID.Eq(sessionID)).
		Delete(); err != nil {
		return err
	}

	// 删除会话
	_, err := c.q.ChatSession.WithContext(ctx).
		Where(c.q.ChatSession.ID.Eq(sessionID)).
		Delete()
	return err
}

// CreateMessage 创建消息
func (c *ChatDao) CreateMessage(ctx context.Context, params CreateMessageParams) (*model.ChatMessage, error) {
	msg := &model.ChatMessage{
		SessionID:  params.SessionID,
		UserID:     params.UserID,
		Role:       params.Role,
		Content:    params.Content,
		TokenCount: params.TokenCount,
		CreatedAt:  time.Now(),
	}
	err := c.q.ChatMessage.WithContext(ctx).Create(msg)
	return msg, err
}

// GetMessagesBySessionID 获取会话的所有消息（分页）
func (c *ChatDao) GetMessagesBySessionID(ctx context.Context, sessionID int64, limit, offset int) ([]*model.ChatMessage, int64, error) {
	q := c.q.ChatMessage.WithContext(ctx).Where(c.q.ChatMessage.SessionID.Eq(sessionID))

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	messages, err := q.
		Order(c.q.ChatMessage.CreatedAt.Asc()).
		Limit(limit).
		Offset(offset).
		Find()

	return messages, total, err
}

// GetRecentMessages 获取会话最近的N条消息
func (c *ChatDao) GetRecentMessages(ctx context.Context, sessionID int64, count int) ([]*model.ChatMessage, error) {
	return c.q.ChatMessage.
		WithContext(ctx).
		Where(c.q.ChatMessage.SessionID.Eq(sessionID)).
		Order(c.q.ChatMessage.CreatedAt.Desc()).
		Limit(count).
		Find()
}

// GetMessageByID 根据ID获取消息
func (c *ChatDao) GetMessageByID(ctx context.Context, messageID int64) (*model.ChatMessage, error) {
	msg, err := c.q.ChatMessage.
		WithContext(ctx).
		Where(c.q.ChatMessage.ID.Eq(messageID)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return msg, nil
}

// GetMessagesByIDs 根据ID列表获取消息
func (c *ChatDao) GetMessagesByIDs(ctx context.Context, messageIDs []int64) ([]*model.ChatMessage, error) {
	return c.q.ChatMessage.
		WithContext(ctx).
		Where(c.q.ChatMessage.ID.In(messageIDs...)).
		Order(c.q.ChatMessage.CreatedAt.Asc()).
		Find()
}

// GetMessagesBySessionIDAndRole 获取会话中特定角色的消息
func (c *ChatDao) GetMessagesBySessionIDAndRole(ctx context.Context, sessionID int64, role string) ([]*model.ChatMessage, error) {
	return c.q.ChatMessage.
		WithContext(ctx).
		Where(c.q.ChatMessage.SessionID.Eq(sessionID)).
		Where(c.q.ChatMessage.Role.Eq(role)).
		Order(c.q.ChatMessage.CreatedAt.Asc()).
		Find()
}

// CheckSessionBelongsToUser 检查会话是否属于用户
func (c *ChatDao) CheckSessionBelongsToUser(ctx context.Context, sessionID, userID int64) (bool, error) {
	count, err := c.q.ChatSession.
		WithContext(ctx).
		Where(c.q.ChatSession.ID.Eq(sessionID)).
		Where(c.q.ChatSession.UserID.Eq(userID)).
		Count()
	return count > 0, err
}

// CheckMessageBelongsToUser 检查消息是否属于用户
func (c *ChatDao) CheckMessageBelongsToUser(ctx context.Context, messageID, userID int64) (bool, error) {
	count, err := c.q.ChatMessage.
		WithContext(ctx).
		Where(c.q.ChatMessage.ID.Eq(messageID)).
		Where(c.q.ChatMessage.UserID.Eq(userID)).
		Count()
	return count > 0, err
}

func NewChatDao(q *query.Query) *ChatDao {
	return &ChatDao{q: q}
}
