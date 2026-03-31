package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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
	// 上下文文本处理
	ChatWithContext(ctx context.Context, messages []ChatMessage) (string, error)
	// 上下文文本流式处理
	ChatStream(ctx context.Context, message []ChatMessage, onDelta func(string) bool) error
	// 全局思维导图 - 分析多个顶节点之间的关联关系
	BuildGlobalMindMap(ctx context.Context, nodes []GlobalNode) ([]GlobalRelation, error)
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
	chatClient       *openai.Client // 对话专用
	embeddingBaseURL string         // 本地向量服务地址
	chatModel        string
	embeddingModel   string
}

const maxRetries = 3

// NewOpenAIClientFromConfig 从拆分后的配置创建客户端
func NewOpenAIClientFromConfig(cfg *config.Config) *OpenAIClient {
	client := &OpenAIClient{
		chatModel:        cfg.AI.Chat.Model,
		embeddingModel:   cfg.AI.Embedding.Model,
		embeddingBaseURL: cfg.AI.Embedding.BaseURL,
	}

	chatConfig := openai.DefaultConfig(cfg.AI.Chat.APIKey)
	chatConfig.BaseURL = cfg.AI.Chat.BaseURL
	client.chatClient = openai.NewClientWithConfig(chatConfig)

	return client
}

// isRetryable 判断是否为可重试的错误（rate limit 或服务端 5xx）
func isRetryable(err error) bool {
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		return apiErr.HTTPStatusCode == 429 || apiErr.HTTPStatusCode >= 500
	}
	return false
}

// 通用文本处理
func (c *OpenAIClient) ProcessText(ctx context.Context, prompt string, content string) (string, error) {
	var lastErr error
	for i := range maxRetries {
		resp, err := c.chatClient.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model: c.chatModel,
				Messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleSystem, Content: prompt},
					{Role: openai.ChatMessageRoleUser, Content: content},
				},
				Temperature: 0.3,
				MaxTokens:   2000,
			},
		)
		if err == nil {
			if len(resp.Choices) == 0 {
				return "", fmt.Errorf("AI返回空结果")
			}
			log.Println("大模型返回结果：", resp.Choices[0].Message.Content)
			return strings.TrimSpace(resp.Choices[0].Message.Content), nil
		}
		if !isRetryable(err) {
			return "", fmt.Errorf("AI处理失败: %w", err)
		}
		lastErr = err
		wait := time.Duration(1<<i) * time.Second // 1s, 2s, 4s
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(wait):
		}
	}
	return "", fmt.Errorf("AI处理失败（重试%d次）: %w", maxRetries, lastErr)
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
	prompt := `
请从技术笔记中抽取知识结构。

【硬性约束（必须严格遵守）】

总节点数 ≤ 20（超过视为失败）
层级 ≤ 3
每个节点最多 3 个 children
description ≤ 20 字

【选择策略（必须执行）】
如果信息过多：

优先保留抽象概念（如“布局系统”）
删除具体实现（如“justify-between”）
删除低重要性节点（importanceScore < 0.7）

【生成要求】

在输出前先规划节点数量
确保不会超过限制
只返回完整 JSON
返回JSON：

{
  "knowledgePoints": {
    "name": "",
    "description": "",
    "importanceScore": 0.5,
    "children": [
      {
        "name": "",
        "description": "",
        "importanceScore": 0.9,
        "children": []
      }
    ]
  }
}

技术笔记如下：
`

	result, err := c.ProcessText(ctx, prompt, content)
	if err != nil {
		return nil, err
	}

	result = cleanJSONResponse(result)

	// 解析JSON结果
	var response struct {
		KnowledgePoints KnowledgePoint `json:"knowledgePoints"`
	}

	err = json.Unmarshal([]byte(result), &response)
	if err != nil {
		return nil, err
	}

	return []KnowledgePoint{response.KnowledgePoints}, nil
}

// 文本嵌入 - 调用本地 embedding-service
func (c *OpenAIClient) GetEmbedding(ctx context.Context, content string) ([]float32, error) {
	type reqBody struct {
		Texts []string `json:"texts"`
	}
	type respBody struct {
		Vectors [][]float32 `json:"vectors"`
	}

	const maxRetries = 3
	var lastErr error
	for i := range maxRetries {
		body, _ := json.Marshal(reqBody{Texts: []string{content}})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			c.embeddingBaseURL+"/embed", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("构建向量请求失败: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
		} else {
			if resp.StatusCode >= 500 {
				resp.Body.Close()
				lastErr = fmt.Errorf("向量服务返回 %d", resp.StatusCode)
			} else if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, fmt.Errorf("向量服务返回 %d", resp.StatusCode)
			} else {
				var result respBody
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					resp.Body.Close()
					return nil, fmt.Errorf("解析向量响应失败: %w", err)
				}
				resp.Body.Close()
				if len(result.Vectors) == 0 {
					return nil, fmt.Errorf("嵌入向量返回空结果")
				}
				return result.Vectors[0], nil
			}
		}

		wait := time.Duration(1<<i) * time.Second
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
	return nil, fmt.Errorf("获取嵌入向量失败（重试%d次）: %w", maxRetries, lastErr)
}

func cleanJSONResponse(result string) string {
	result = strings.TrimSpace(result)

	start := strings.Index(result, "{")
	end := strings.LastIndex(result, "}")

	if start != -1 && end != -1 {
		result = result[start : end+1]
	}

	return result
}

// BuildGlobalMindMap 分析多个顶节点之间的关联关系，生成全局思维导图连接
func (c *OpenAIClient) BuildGlobalMindMap(ctx context.Context, nodes []GlobalNode) ([]GlobalRelation, error) {
	if len(nodes) < 2 {
		return nil, nil
	}

	// 将顶节点列表序列化为 JSON 作为输入
	nodeListJSON, err := json.Marshal(nodes)
	if err != nil {
		return nil, fmt.Errorf("序列化节点列表失败: %w", err)
	}

	prompt := `你是一个知识图谱分析助手。给定以下知识主题列表（每个主题来自不同笔记的核心概念），请分析它们之间存在的有意义的关联关系。

要求：
1. 只建立真实存在的知识关联，不要强行连接
2. 关系类型举例：包含、依赖、对比、扩展、应用
3. 每对节点最多建立一条关系
4. 如果两个节点没有明显关联，不要建立关系

返回 JSON 格式（只返回 JSON，不要其他内容）：
{
  "relations": [
    {"from_id": 1, "to_id": 2, "label": "关系描述"},
    ...
  ]
}`

	result, err := c.ProcessText(ctx, prompt, string(nodeListJSON))
	if err != nil {
		return nil, fmt.Errorf("全局思维导图生成失败: %w", err)
	}

	result = cleanJSONResponse(result)

	var response struct {
		Relations []GlobalRelation `json:"relations"`
	}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("解析全局关系失败: %w", err)
	}

	return response.Relations, nil
}

func (c *OpenAIClient) ChatWithContext(ctx context.Context, messages []ChatMessage) (string, error) {
	var lastErr error

	var openaiMessages []openai.ChatCompletionMessage
	for _, m := range messages {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	for i := 0; i < maxRetries; i++ {
		resp, err := c.chatClient.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:       c.chatModel,
				Messages:    openaiMessages,
				Temperature: 0.3,
				MaxTokens:   300,
			},
		)
		if err == nil {
			if len(resp.Choices) == 0 {
				return "", fmt.Errorf("AI返回空结果")
			}
			return strings.TrimSpace(resp.Choices[0].Message.Content), nil
		}

		if !isRetryable(err) {
			return "", fmt.Errorf("AI处理失败: %w", err)
		}

		lastErr = err
		wait := time.Duration(1<<i) * time.Second
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(wait):
		}
	}
	return "", fmt.Errorf("AI处理失败（重试%d次）: %w", maxRetries, lastErr)
}

func (c *OpenAIClient) ChatStream(
	ctx context.Context,
	message []ChatMessage,
	onDelta func(string) bool,
) error {
	var openaiMessage []openai.ChatCompletionMessage
	for _, m := range message {
		openaiMessage = append(openaiMessage, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	stream, err := c.chatClient.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:       c.chatModel,
			Messages:    openaiMessage,
			Temperature: 0.3,
			MaxTokens:   300,
			Stream:      true,
		},
	)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		if len(response.Choices) == 0 {
			continue
		}
		delta := response.Choices[0].Delta.Content
		if delta != "" {
			continue
		}
		if !onDelta(delta) {
			return nil
		}
	}
	return nil
}
