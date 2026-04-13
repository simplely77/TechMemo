package dto

import "time"

// GetSearchHistoryReq 获取搜索历史请求
type GetSearchHistoryReq struct {
	Page       int64  `form:"page"`
	PageSize   int64  `form:"page_size"`
	SearchType string `form:"search_type"`
	TargetType string `form:"target_type"`
}

// GetSearchHistoryResp 获取搜索历史响应
type GetSearchHistoryResp struct {
	Items    []SearchHistoryItem `json:"items"`
	Total    int64               `json:"total"`
	Page     int64               `json:"page"`
	PageSize int64               `json:"page_size"`
}

// SearchHistoryItem 单条搜索历史
type SearchHistoryItem struct {
	ID             int64     `json:"id"`
	Keyword        string    `json:"keyword"`
	SearchType     string    `json:"search_type"`
	TargetType     string    `json:"target_type"`
	LastSearchedAt time.Time `json:"last_searched_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// SemanticSearchReq 语义搜索请求（混合检索：向量 + 关键词）
type SemanticSearchReq struct {
	Query      string `json:"query" binding:"required"`
	SearchType string `json:"search_type" binding:"required,oneof=note knowledge"`
	TopK       int    `json:"top_k" binding:"required,min=1,max=100"`
}

// SemanticSearchResp 语义搜索响应
type SemanticSearchResp struct {
	Results []SearchResultItem `json:"results"`
	Query   string             `json:"query"`
	Total   int                `json:"total"`
}

// SearchResultItem 搜索结果项
type SearchResultItem struct {
	ID         int64   `json:"id"`
	Type       string  `json:"type"` // "note" 或 "knowledge"
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"` // 无 rerank 时为 RRF 融合分约 0–1；有 rerank 时为 sigmoid(原始分) 约 0–1
	RerankScore *float64 `json:"rerank_score,omitempty"` // CrossEncoder 原始分，仅启用 rerank 且成功时返回

	// Note 特有字段
	NoteType string        `json:"note_type,omitempty"`
	Category *NoteCategory `json:"category,omitempty"`

	// Knowledge 特有字段
	SourceNoteID    int64   `json:"source_note_id,omitempty"`
	SourceNoteTitle string  `json:"source_note_title,omitempty"`
	ImportanceScore float64 `json:"importance_score,omitempty"`

	CreatedAt string `json:"created_at"`
}
