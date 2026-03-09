package aiclient

// KnowledgePoint 表示从笔记中抽取的核心知识单元
type KnowledgePoint struct {
	Name            string  `json:"name"`            // 知识点名称
	Description     string  `json:"description"`     // 知识点说明
	ImportanceScore float64 `json:"importanceScore"` // 重要性评分
}

// ExtractResult 知识抽取结果
type ExtractResult struct {
	KnowledgePoints []KnowledgePoint `json:"knowledgePoints"`
	OverallScore    float64          `json:"overallScore"` // 整体重要性评分
}

// EmbeddingResult 向量生成结果
type EmbeddingResult struct {
	TargetType string    `json:"targetType"` // note / knowledge
	TargetID   int64     `json:"targetId"`   // 对应对象ID
	Vector     []float32 `json:"vector"`     // 向量数据
}
