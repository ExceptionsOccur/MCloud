package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// AIExtractService 申请表单提取编排（T-068）：解析文件 → 调大模型分类+提取 → 校验归一化。
// 不写业务库；文件落盘由 AIFileStore 负责。

type AIExtractService struct {
	parser     *AIDocParser
	llm        *AILLMClient
	settingSvc *AISettingService
	fileStore  *AIFileStore
}

func NewAIExtractService() *AIExtractService {
	return &AIExtractService{
		parser:     NewAIDocParser(),
		llm:        NewAILLMClient(),
		settingSvc: NewAISettingService(),
		fileStore:  NewAIFileStore(),
	}
}

// AIExtractResult 提取结果（返回前端预览）
type AIExtractResult struct {
	DocType     string                 `json:"doc_type"`
	DocTypeName string                 `json:"doc_type_name"`
	Fields      map[string]interface{} `json:"fields"`
	Warnings    []string               `json:"warnings"`
	TextPreview string                 `json:"text_preview"`
	FileToken   string                 `json:"file_token"`
}

// llmOutput 大模型返回 JSON 结构
type llmOutput struct {
	DocType  string                 `json:"doc_type"`
	Fields   map[string]interface{} `json:"fields"`
	Warnings []string               `json:"warnings"`
}

// Extract 编排：解析 → 提取 → 归一化，并将原文件暂存 incoming 返回 file_token
func (s *AIExtractService) Extract(filename string, data []byte) (*AIExtractResult, error) {
	apiURL, apiKey, model, err := s.settingSvc.GetDecrypted()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(apiURL) == "" || strings.TrimSpace(model) == "" {
		return nil, ErrLLMNotConfigured
	}

	text, err := s.parser.Parse(filename, data)
	if err != nil {
		return nil, err
	}

	prompt := buildExtractionPrompt(text)
	content, err := s.llm.Complete(apiURL, apiKey, model, prompt)
	if err != nil {
		return nil, err
	}

	out, err := parseLLMOutput(content)
	if err != nil {
		return nil, err
	}

	// 暂存原文件（供确认入库后归档）
	token, err := s.fileStore.SaveIncoming(data, filename)
	if err != nil {
		return nil, fmt.Errorf("文件暂存失败: %w", err)
	}

	docTypeName := "未知类型"
	if t, ok := docTypeByKey(out.DocType); ok {
		docTypeName = t.Label
	}

	warnings := out.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	if out.DocType == "unknown" {
		warnings = append(warnings, "未能识别文档类型，请确认是否为受支持的申请表单")
	}

	return &AIExtractResult{
		DocType:     out.DocType,
		DocTypeName: docTypeName,
		Fields:      out.Fields,
		Warnings:    warnings,
		TextPreview: truncateText(text),
		FileToken:   token,
	}, nil
}

// Archive 将确认入库成功的文件归档
func (s *AIExtractService) Archive(token string) (string, error) {
	return s.fileStore.Archive(token)
}

// parseLLMOutput 解析大模型返回的 JSON（容忍 ```json 代码块包裹）
func parseLLMOutput(content string) (*llmOutput, error) {
	content = strings.TrimSpace(content)
	content = stripCodeFence(content)

	var out llmOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, errors.New("大模型返回内容不是合法 JSON")
	}
	if out.DocType == "" {
		out.DocType = "unknown"
	}
	if out.Fields == nil {
		out.Fields = map[string]interface{}{}
	}
	return &out, nil
}

// stripCodeFence 去掉 ```json ... ``` 包裹
func stripCodeFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimPrefix(s, "json")
	s = strings.TrimPrefix(s, "JSON")
	if idx := strings.LastIndex(s, "```"); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}
