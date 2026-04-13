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
	// 重排序 - 根据查询和文档列表，返回每个文档的得分
	Rerank(ctx context.Context, query string, documents []string) ([]float64, error)
	// RerankEnabled 当配置了 rerank.base_url 时为 true
	RerankEnabled() bool
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

func (c *OpenAIClient) RerankEnabled() bool {
	return strings.TrimSpace(c.rerankBaseURL) != ""
}

type OpenAIClient struct {
	chatClient       *openai.Client // 对话专用
	embeddingBaseURL string         // 本地向量服务地址
	chatModel        string
	embeddingModel   string
	rerankBaseURL    string
}

const maxRetries = 3

// NewOpenAIClientFromConfig 从拆分后的配置创建客户端
func NewOpenAIClientFromConfig(cfg *config.Config) *OpenAIClient {
	client := &OpenAIClient{
		chatModel:        cfg.AI.Chat.Model,
		embeddingModel:   cfg.AI.Embedding.Model,
		embeddingBaseURL: cfg.AI.Embedding.BaseURL,
		rerankBaseURL:    cfg.AI.Rerank.BaseURL,
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
请判断以下笔记内容是否值得进行 AI 知识抽取：

- extract：内容包含可复用的知识（概念、原理、实现方法、经验、配置等），适合结构化整理
- ignore：内容为随手记录、情绪、无结构信息或无复用价值

【判断标准（必须遵守）】
1. 只要包含以下任意一种，即判定为 extract：
   - 明确概念或定义
   - 技术原理或解释
   - 可复用的实现方式（代码、命令、配置）
   - 有通用价值的经验或总结

2. 以下情况判定为 ignore：
   - 纯个人记录（如 TODO、日记）
   - 零散片段且无法理解上下文
   - 无实际技术或知识价值

【输出要求】
- 只返回 extract 或 ignore
- 不要输出任何解释或其他内容

【笔记内容】
`

	result, err := c.ProcessText(ctx, prompt, content)
	if err != nil {
		return "ignore", err
	}

	result = strings.TrimSpace(strings.ToLower(result))

	switch result {
	case "extract", "ignore":
		return result, nil
	default:
		// LLM 偶尔不听话，兜底
		return "ignore", nil
	}
}

// 知识抽取 - 专注于README中描述的知识点抽取功能
func (c *OpenAIClient) ExtractKnowledgePoints(ctx context.Context, content string) ([]KnowledgePoint, error) {
	prompt := `
请从技术笔记中抽取结构化知识点。

【目标】
构建“可用于知识库沉淀”的知识结构，既包含抽象概念，也包含关键实现与实用技巧。

【约束（必须遵守）】
- 总节点数 ≤ 100
- 层级 ≤ 6
- 每个节点最多 8 个 children
- description ≤ 100 字

【抽取策略（必须执行）】
1. 优先抽取以下三类内容：
   - 核心概念（是什么）
   - 关键原理（为什么）
   - 实用实现（怎么做）

2. 不要删除具体实现，尤其是：
   - 常见写法（如 API / 代码模式）
   - 关键配置（如参数、命令）
   - 典型用法（如使用场景）

3. 删除或弱化以下内容：
   - 纯样式/低层细节（如无意义参数组合）
   - 重复信息
   - 重要性较低（importanceScore < 0.6）

4. 抽象与具体要平衡：
   - 每个抽象节点下，至少包含1个具体实现或示例

【生成要求】
- 在输出前先规划节点数量，确保不超过限制
- 结构要有逻辑层次（由抽象到具体）
- importanceScore 取值范围 [0,1]，保留两位小数
- 只返回完整 JSON，不要解释

【输出格式】
{
  "knowledgePoints": {
    "name": "",
    "description": "",
    "importanceScore": 0.8,
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

【技术笔记】
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
				MaxTokens:   3000,
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
			MaxTokens:   3000,
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
		log.Println("流式传输大模型返回数据：", delta)
		if delta == "" {
			continue
		}
		if !onDelta(delta) {
			return nil
		}
	}
	return nil
}

func (c *OpenAIClient) Rerank(ctx context.Context, query string, documents []string) ([]float64, error) {
	if len(documents) == 0 {
		return nil, nil
	}
	base := strings.TrimRight(strings.TrimSpace(c.rerankBaseURL), "/")
	if base == "" {
		return nil, fmt.Errorf("重排序服务未配置")
	}

	type reqBody struct {
		Query     string   `json:"query"`
		Documents []string `json:"documents"`
	}
	type respBody struct {
		Scores []float64 `json:"scores"`
	}

	const maxRetries = 3
	var lastErr error
	for i := range maxRetries {
		body, err := json.Marshal(reqBody{Query: query, Documents: documents})
		if err != nil {
			return nil, fmt.Errorf("序列化重排序请求失败: %w", err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/rerank", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("构建重排序请求失败: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
		} else {
			if resp.StatusCode >= 500 {
				resp.Body.Close()
				lastErr = fmt.Errorf("重排序服务返回 %d", resp.StatusCode)
			} else if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, fmt.Errorf("重排序服务返回 %d", resp.StatusCode)
			} else {
				var result respBody
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					resp.Body.Close()
					lastErr = fmt.Errorf("解析重排序响应失败: %w", err)
				} else {
					resp.Body.Close()
					if len(result.Scores) != len(documents) {
						return nil, fmt.Errorf("重排序结果条数与文档不一致: %d != %d", len(result.Scores), len(documents))
					}
					return result.Scores, nil
				}
			}
		}
		wait := time.Duration(1<<i) * time.Second
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
	return nil, fmt.Errorf("重排序失败（重试%d次）: %w", maxRetries, lastErr)
}
