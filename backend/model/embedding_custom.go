package model

import "time"

// EmbeddingCustom 向量存储模型。
type EmbeddingCustom struct {
	ID         int64     `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   int64     `json:"target_id"`
	Vector     []float32 `json:"vector"`
	ModelName  string    `json:"model_name"`
	CreatedAt  time.Time `json:"created_at"`
}
