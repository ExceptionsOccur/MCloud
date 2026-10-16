package services

import (
	"errors"
	"strings"

	"mcloud/config"
	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"

	"gorm.io/gorm"
)

// settingsRowID AI 配置为单条全局配置，主键恒为 1
const settingsRowID = 1

type AISettingService struct{}

func NewAISettingService() *AISettingService {
	return &AISettingService{}
}

// AISettingResponse 对外返回：api_key 仅回显脱敏值，绝不返回明文
type AISettingResponse struct {
	APIURL       string `json:"api_url"`
	APIKeyMasked string `json:"api_key_masked"`
	APIKeySet    bool   `json:"api_key_set"`
	Model        string `json:"model"`
}

// SaveAISettingRequest 保存请求：api_key 为空表示「保留原值」（前端脱敏回显无法回传明文）
type SaveAISettingRequest struct {
	APIURL string `json:"api_url" binding:"required"`
	APIKey string `json:"api_key"`
	Model  string `json:"model" binding:"required"`
}

func (s *AISettingService) loadOrEmpty() (*models.AISetting, error) {
	var setting models.AISetting
	err := database.DB.First(&setting, settingsRowID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.AISetting{ID: settingsRowID}, nil
		}
		return nil, err
	}
	return &setting, nil
}

// Get 返回脱敏配置（未配置时返回空配置）
func (s *AISettingService) Get() (*AISettingResponse, error) {
	setting, err := s.loadOrEmpty()
	if err != nil {
		return nil, err
	}

	plain, err := utils.DecryptAPIKey(setting.APIKeyEnc, config.AppConfig.AISecretKey)
	if err != nil {
		// 密文损坏时不阻断读取：按「未设置 key」处理
		plain = ""
	}

	return &AISettingResponse{
		APIURL:       setting.APIURL,
		APIKeyMasked: utils.MaskAPIKey(plain),
		APIKeySet:    plain != "",
		Model:        setting.Model,
	}, nil
}

// Save 保存配置（upsert）；api_key 为空保留原值
func (s *AISettingService) Save(req SaveAISettingRequest) error {
	req.APIURL = strings.TrimSpace(req.APIURL)
	req.Model = strings.TrimSpace(req.Model)
	req.APIKey = strings.TrimSpace(req.APIKey)
	if req.APIURL == "" {
		return errors.New("API 地址不能为空")
	}
	if req.Model == "" {
		return errors.New("模型名称不能为空")
	}

	setting, err := s.loadOrEmpty()
	if err != nil {
		return err
	}

	// 仅当显式传入非空 key 时更新，否则保留已存密文
	if req.APIKey != "" {
		enc, err := utils.EncryptAPIKeyWithWarning(req.APIKey, config.AppConfig.AISecretKey)
		if err != nil {
			return err
		}
		setting.APIKeyEnc = enc
	}

	setting.APIURL = req.APIURL
	setting.Model = req.Model

	if err := database.DB.Save(setting).Error; err != nil {
		return err
	}
	return nil
}

// GetDecrypted 内部调用：返回 API 地址、模型与解密后的明文 Key（供 LLM 客户端使用）
func (s *AISettingService) GetDecrypted() (apiURL, apiKey, model string, err error) {
	setting, err := s.loadOrEmpty()
	if err != nil {
		return "", "", "", err
	}
	plain, err := utils.DecryptAPIKey(setting.APIKeyEnc, config.AppConfig.AISecretKey)
	if err != nil {
		return "", "", "", err
	}
	return setting.APIURL, plain, setting.Model, nil
}
