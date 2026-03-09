package aiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"techmemo/backend/config"

	"github.com/sashabaranov/go-openai"
)

// AIClient 专注于README中描述的核心AI处理功能
type AIClient interface {
	// 笔记类型判断 - 智能判断笔记内容类型
	ClassifyNoteType(ctx context.Context, content string) (string, error)
	// 知识抽取 - 从笔记中抽取核心知识点
	ExtractKnowledgePoints(ctx context.Context, content string) ([]KnowledgePoint, error)
	// 文本嵌入 - 生成向量表示，支持"粗到细"的语义检索策略
	GetEmbedding(ctx context.Context, content string) ([]float32, error)
	// 通用文本处理
	ProcessText(ctx context.Context, prompt string, content string) (string, error)
	// 获取当前使用的模型名称
	GetChatModelName() string
	// 获取当前使用的向量模型名称
	GetEmbeddingModelName() string
}

func (a *OpenAIClient) GetChatModelName() string {
	return a.chatModel
}

func (a *OpenAIClient) GetEmbeddingModelName() string {
	return a.embeddingModel
}

type OpenAIClient struct {
	chatClient      *openai.Client // 对话专用
	embeddingClient *openai.Client // 向量专用
	chatModel       string
	embeddingModel  string
}

// NewOpenAIClientFromConfig 从拆分后的配置创建客户端
func NewOpenAIClientFromConfig(cfg *config.Config) *OpenAIClient {
	client := &OpenAIClient{
		chatModel:      cfg.AI.Chat.Model,
		embeddingModel: cfg.AI.Embedding.Model,
	}

	// 1. 初始化对话客户端
	chatConfig := openai.DefaultConfig(cfg.AI.Chat.APIKey)
	chatConfig.BaseURL = cfg.AI.Chat.BaseURL
	client.chatClient = openai.NewClientWithConfig(chatConfig)

	// 2. 初始化向量客户端
	embedConfig := openai.DefaultConfig(cfg.AI.Embedding.APIKey)
	embedConfig.BaseURL = cfg.AI.Embedding.BaseURL
	client.embeddingClient = openai.NewClientWithConfig(embedConfig)

	return client
}

// 通用文本处理
func (c *OpenAIClient) ProcessText(ctx context.Context, prompt string, content string) (string, error) {
	resp, err := c.chatClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.chatModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
			Temperature: 0.3,
			MaxTokens:   2000,
		},
	)

	if err != nil {
		return "", fmt.Errorf("AI处理失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("AI返回空结果")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// ClassifyNoteType 智能判断笔记是否值得进行后续 AI 处理
func (c *OpenAIClient) ClassifyNoteType(ctx context.Context, content string) (string, error) {
	// 内容过短，直接忽略
	if len(strings.TrimSpace(content)) < 50 {
		return "ignore", nil
	}

	prompt := `
请判断以下笔记内容是否值得进行后续 AI 知识处理，并返回最合适的类型：

- knowledge：内容包含明确、可总结、可抽取的知识点，适合生成结构化知识
- reference：内容主要是资料、命令、配置、链接或参考信息，仅用于检索
- ignore：内容是个人记录、情绪、草稿，或不具备知识价值

请只返回 knowledge、reference 或 ignore，不要输出任何其他内容。
`

	result, err := c.ProcessText(ctx, prompt, content)
	if err != nil {
		return "ignore", err
	}

	result = strings.TrimSpace(strings.ToLower(result))

	switch result {
	case "knowledge", "reference", "ignore":
		return result, nil
	default:
		// LLM 偶尔不听话，兜底
		return "ignore", nil
	}
}

// 知识抽取 - 专注于README中描述的知识点抽取功能
func (c *OpenAIClient) ExtractKnowledgePoints(ctx context.Context, content string) ([]KnowledgePoint, error) {
	prompt := `请从以下技术笔记中抽取核心知识点，要求：
1. 识别3-8个核心知识点
2. 每个知识点包含：名称、简要说明、重要性评分(1-10分)
3. 返回JSON格式：{"knowledgePoints": [{"name": "", "description": "", "importanceScore": 0}]}
4. 只返回JSON，不要其他内容`

	result, err := c.ProcessText(ctx, prompt, content)
	if err != nil {
		return nil, err
	}

	// 解析JSON结果
	var response struct {
		KnowledgePoints []KnowledgePoint `json:"knowledgePoints"`
	}

	err = json.Unmarshal([]byte(result), &response)
	if err != nil {
		return nil, err
	}

	return response.KnowledgePoints, nil
}

// 文本嵌入 - 支持"粗到细"的语义检索策略
func (c *OpenAIClient) GetEmbedding(ctx context.Context, content string) ([]float32, error) {
	resp, err := c.embeddingClient.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Model: openai.EmbeddingModel(c.embeddingModel),
			Input: []string{content},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("获取嵌入向量失败: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("嵌入向量返回空结果")
	}

	return resp.Data[0].Embedding, nil
}
