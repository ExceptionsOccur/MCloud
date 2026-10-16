package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AILLMClient OpenAI 兼容大模型客户端（T-068）：仅用标准库 net/http，无新依赖。

// ErrLLMNotConfigured 未配置 AI 接入
var ErrLLMNotConfigured = errors.New("尚未配置 AI 接口，请先在「AI 接入设置」中填写 API 地址、密钥与模型")

// ErrLLMRequestFailed 调用大模型失败
var ErrLLMRequestFailed = errors.New("调用大模型失败")

// llmTimeout 单次请求超时
const llmTimeout = 60 * time.Second

type AILLMClient struct{}

func NewAILLMClient() *AILLMClient {
	return &AILLMClient{}
}

// chatMessage OpenAI chat 消息
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest OpenAI 兼容请求体
type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

// responseFormat 强制 JSON 输出
type responseFormat struct {
	Type string `json:"type"`
}

// chatResponse OpenAI 兼容响应体
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete 调用 {apiURL}/chat/completions，返回模型输出文本（JSON 字符串）
func (c *AILLMClient) Complete(apiURL, apiKey, model, prompt string) (string, error) {
	apiURL = strings.TrimSpace(apiURL)
	if apiURL == "" || model == "" {
		return "", ErrLLMNotConfigured
	}

	endpoint := strings.TrimRight(apiURL, "/") + "/chat/completions"

	body, err := json.Marshal(chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature:    0,
		ResponseFormat: &responseFormat{Type: "json_object"},
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrLLMRequestFailed, err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrLLMRequestFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: llmTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrLLMRequestFailed, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrLLMRequestFailed, err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("%w: 响应解析失败（HTTP %d）", ErrLLMRequestFailed, resp.StatusCode)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("%w: %s", ErrLLMRequestFailed, parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: HTTP %d", ErrLLMRequestFailed, resp.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("%w: 模型未返回内容", ErrLLMRequestFailed)
	}

	return parsed.Choices[0].Message.Content, nil
}
