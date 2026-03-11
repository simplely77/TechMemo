package dto

type GetKnowledgePointsReq struct {
	SourceNoteID  int64   `form:"source_note_id" json:"source_note_id"`
	Keyword       string  `form:"keyword" json:"keyword"`
	MinImportance float64 `form:"min_importance" json:"min_importance"`
	Page          int64   `form:"page" json:"page"`
	PageSize      int64   `form:"page_size" json:"page_size"`
}

type GetKnowledgePointsResp struct {
	KnowledgePoints []KnowledgePointItem `json:"knowledge_points"`
	Total           int64                `json:"total"`
	Page            int64                `json:"page"`
	PageSize        int64                `json:"page_size"`
}

type KnowledgePointItem struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	SourceNoteID    int64   `json:"source_note_id"`
	SourceNoteTitle string  `json:"source_note_title,omitempty"`
	ImportanceScore float64 `json:"importance_score"`
	CreatedAt       string  `json:"created_at"`
}

type GetKnowledgePointResp struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	SourceNoteID     int64             `json:"source_note_id"`
	SourceNoteTitle  string            `json:"source_note_title,omitempty"`
	ImportanceScore  float64           `json:"importance_score"`
	RelatedKnowledge []RelatedKnowledge `json:"related_knowledge,omitempty"`
	CreatedAt        string            `json:"created_at"`
}

type RelatedKnowledge struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	RelationType string `json:"relation_type"`
}

type UpdateKnowledgePointReq struct {
	Name            string  `json:"name" binding:"required"`
	Description     string  `json:"description"`
	ImportanceScore float64 `json:"importance_score" binding:"min=1,max=10"`
}

// GetMindMapReq 获取思维导图请求
type GetMindMapReq struct {
	NoteID int64 `form:"note_id" binding:"required"`
}

// MindMapNode 思维导图节点
type MindMapNode struct {
	ID              int64         `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	ImportanceScore float64       `json:"importance_score"`
	Children        []MindMapNode `json:"children,omitempty"`
}

// GetMindMapResp 思维导图响应
type GetMindMapResp struct {
	NoteID int64         `json:"note_id"`
	Nodes  []MindMapNode `json:"nodes"`
}