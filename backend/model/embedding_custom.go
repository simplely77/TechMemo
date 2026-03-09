package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// EmbeddingCustom 自定义的Embedding类型，用于处理向量存储
type EmbeddingCustom struct {
	ID         int64     `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   int64     `json:"target_id"`
	Vector     []float32 `json:"vector"`
	ModelName  string    `json:"model_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToDBFormat 将向量转换为数据库存储格式
func (e *EmbeddingCustom) ToDBFormat() (*Embedding, error) {
	// 将[]float32序列化为字符串
	vectorBytes, err := json.Marshal(e.Vector)
	if err != nil {
		return nil, fmt.Errorf("序列化向量失败: %w", err)
	}

	return &Embedding{
		ID:         e.ID,
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		Vector:     string(vectorBytes),
		ModelName:  e.ModelName,
		CreatedAt:  e.CreatedAt,
	}, nil
}

// FromDBFormat 从数据库格式解析向量
func (e *EmbeddingCustom) FromDBFormat(dbEmbedding *Embedding) error {
	e.ID = dbEmbedding.ID
	e.TargetType = dbEmbedding.TargetType
	e.TargetID = dbEmbedding.TargetID
	e.ModelName = dbEmbedding.ModelName
	e.CreatedAt = dbEmbedding.CreatedAt

	// 解析向量字符串
	if strings.TrimSpace(dbEmbedding.Vector) != "" {
		var vector []float32
		err := json.Unmarshal([]byte(dbEmbedding.Vector), &vector)
		if err != nil {
			return fmt.Errorf("解析向量失败: %w", err)
		}
		e.Vector = vector
	}

	return nil
}