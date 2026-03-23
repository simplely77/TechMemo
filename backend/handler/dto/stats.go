package dto

type GetOverviewStatsResp struct {
	TotalNotes           int64 `json:"total_notes"`
	TotalKnowledgePoints int64 `json:"total_knowledge_point"`
	TotalCategories      int64 `json:"total_categories"`
	TotalTags            int64 `json:"total_tags"`
}

type CategoryStats struct {
	CategoryID     int64  `json:"category_id"`
	CategoryName   string `json:"category_name"`
	NoteCount      int64  `json:"note_count"`
	KnowledgeCount int64  `json:"knowledge_count"`
}

type GetCategoriesStatsResp struct {
	Categories []CategoryStats `json:"categories"`
}
