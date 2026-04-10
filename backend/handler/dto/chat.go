package dto

type CreateSessionReq struct {
	Title string `json:"title"` // 可选；不传则服务端生成「新会话 N」
}

type UpdateSessionReq struct {
	Title string `json:"title" binding:"required,min=1,max=200"`
}

type SendMessageReq struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

type CreateSessionResp struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ChatSessionListResp struct {
	Sessions []CreateSessionResp `json:"sessions"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

type ChatMessageResp struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type ChatMessageListResp struct {
	Messages []ChatMessageResp `json:"messages"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}
