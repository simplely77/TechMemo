package aiclient

// KnowledgePoint 表示从笔记中抽取的核心知识单元
type KnowledgePoint struct {
	Name            string           `json:"name"`            // 知识点名称
	Description     string           `json:"description"`     // 知识点说明
	ImportanceScore float64          `json:"importanceScore"` // 重要性评分
	Children        []KnowledgePoint `json:"children"`        //子节点
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

// GlobalNode 用于全局思维导图的顶节点输入
type GlobalNode struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GlobalRelation 全局思维导图中顶节点之间的关联关系
type GlobalRelation struct {
	FromID int64  `json:"from_id"`
	ToID   int64  `json:"to_id"`
	Label  string `json:"label"` // 关系描述，如"包含"、"依赖"
}

type ChatMessage struct {
	Role    string // "user"/"assistant"/"system"
	Content string
}
