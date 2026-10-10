package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

// parsePersonID 将批量更新中的 person_id 原始值转换为 *uint，nil 表示解除关联
func parsePersonID(v interface{}) (*uint, error) {
	if v == nil {
		return nil, nil
	}
	var id uint
	switch n := v.(type) {
	case float64:
		id = uint(n)
	case int:
		id = uint(n)
	case int64:
		id = uint(n)
	case uint:
		id = n
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(n), 10, 64)
		if err != nil {
			return nil, errors.New("无效的人员ID")
		}
		id = uint(parsed)
	default:
		return nil, errors.New("无效的人员ID")
	}
	return &id, nil
}

type BatchCreateItem struct {
	Region            string `json:"region" binding:"required"`
	InstanceID        string `json:"instance_id"`
	Name              string `json:"name" binding:"required"`
	PrivateIP         string `json:"private_ip" binding:"required"`
	AssetType         string `json:"asset_type"`
	OS                string `json:"os"`
	CPU               int    `json:"cpu"`
	CPUArch           string `json:"cpu_arch"`
	Memory            int    `json:"memory"`
	Disk              int    `json:"disk"`
	SystemDisk        int    `json:"system_disk"`
	DataDisk          int    `json:"data_disk"`
	EnvType           string `json:"env_type"`
	IsDBServer        *bool  `json:"is_db_server"`
	Status            string `json:"status"`
	OpenPorts         string `json:"open_ports"`
	Tags              string `json:"tags"`
	Applicant         string `json:"applicant"`
	Project           string `json:"project"`
	ApplyUnit         string `json:"apply_unit"`
	ApplyReason       string `json:"apply_reason"`
	ApplyConfig       string `json:"apply_config"`
	ApplyTime         string `json:"apply_time"`
	ObjectStorageSize string `json:"object_storage_size"`
	Remark            string `json:"remark"`
	ApplicantContact  string `json:"applicant_contact"`
}

type BatchCreateRequest struct {
	Hosts []BatchCreateItem `json:"hosts" binding:"required"`
}

type BatchCreateResponse struct {
	Success    int      `json:"success"`
	Skipped    int      `json:"skipped"`
	Errors     int      `json:"errors"`
	LineErrors []string `json:"line_errors,omitempty"`
}

// appendLineError 记录失败行明细，最多保留 10 条
func appendLineError(list []string, msg string) []string {
	if len(list) >= 10 {
		return list
	}
	return append(list, msg)
}

type BatchCreateTextRequest struct {
	Text string `json:"text" binding:"required"`
}

// BatchCreateFromText 纯文本批量添加：每行一条记录，逗号分隔，列顺序与 CSV 模板一致（最多26列）
func (s *HostService) BatchCreateFromText(text string, op Operator, requestID string) (*BatchCreateResponse, error) {
	resp := &BatchCreateResponse{}
	headerSkipped := false

	for i, raw := range strings.Split(text, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		row, err := parseCSVLine(line)
		if err != nil {
			resp.Errors++
			resp.LineErrors = appendLineError(resp.LineErrors, fmt.Sprintf("第%d行: 格式错误(%v)", lineNo, err))
			continue
		}

		// 跳过表头行（允许直接粘贴带表头的内容）
		if !headerSkipped && len(row) > 0 && strings.TrimSpace(row[0]) == utils.CSVHeaders[0] {
			headerSkipped = true
			continue
		}

		if len(row) > len(utils.CSVHeaders) {
			resp.Errors++
			resp.LineErrors = appendLineError(resp.LineErrors,
				fmt.Sprintf("第%d行: 列数超出 %d 列", lineNo, len(utils.CSVHeaders)))
			continue
		}
		if len(row) < 3 {
			resp.Errors++
			resp.LineErrors = appendLineError(resp.LineErrors,
				fmt.Sprintf("第%d行: 至少需要3列(区域,主机名称,内网IP)", lineNo))
			continue
		}
		for len(row) < len(utils.CSVHeaders) {
			row = append(row, "")
		}

		req, err := s.ParseCSVRowToCreateHost(row)
		if err != nil {
			resp.Errors++
			resp.LineErrors = appendLineError(resp.LineErrors, fmt.Sprintf("第%d行: %v", lineNo, err))
			continue
		}

		if _, err := s.Create(req, op, requestID); err != nil {
			if strings.Contains(err.Error(), "已存在") {
				resp.Skipped++
			} else {
				resp.Errors++
				resp.LineErrors = appendLineError(resp.LineErrors, fmt.Sprintf("第%d行: %v", lineNo, err))
			}
		} else {
			resp.Success++
		}
	}

	return resp, nil
}

func (s *HostService) BatchCreate(req BatchCreateRequest, op Operator, requestID string) (*BatchCreateResponse, error) {
	resp := &BatchCreateResponse{}

	for i, item := range req.Hosts {
		var count int64
		database.DB.Model(&models.Host{}).Where("private_ip = ?", item.PrivateIP).Count(&count)
		if count > 0 {
			resp.Skipped++
			continue
		}

		req2 := CreateHostRequest{
			Region:            item.Region,
			InstanceID:        item.InstanceID,
			Name:              item.Name,
			PrivateIP:         item.PrivateIP,
			AssetType:         item.AssetType,
			OS:                item.OS,
			CPU:               item.CPU,
			CPUArch:           item.CPUArch,
			Memory:            item.Memory,
			Disk:              item.Disk,
			SystemDisk:        item.SystemDisk,
			DataDisk:          item.DataDisk,
			EnvType:           item.EnvType,
			IsDBServer:        item.IsDBServer,
			Status:            item.Status,
			OpenPorts:         item.OpenPorts,
			Tags:              item.Tags,
			Applicant:         item.Applicant,
			Project:           item.Project,
			ApplyUnit:         item.ApplyUnit,
			ApplyReason:       item.ApplyReason,
			ApplyConfig:       item.ApplyConfig,
			ApplyTime:         item.ApplyTime,
			ObjectStorageSize: item.ObjectStorageSize,
			Remark:            item.Remark,
			ApplicantContact:  item.ApplicantContact,
		}

		_, err := s.Create(req2, op, requestID)
		if err != nil {
			resp.Errors++
			resp.LineErrors = appendLineError(resp.LineErrors, fmt.Sprintf("第%d条: %v", i+1, err))
		} else {
			resp.Success++
		}
	}

	return resp, nil
}

type BatchUpdateRequest struct {
	IDS  []uint                 `json:"ids" binding:"required"`
	Data map[string]interface{} `json:"data" binding:"required"`
}

func (s *HostService) BatchUpdate(req BatchUpdateRequest, op Operator, requestID string) error {
	if len(req.IDS) == 0 {
		return errors.New("未选择任何主机")
	}

	// Fields that belong to host_applications table
	appFields := map[string]bool{
		"apply_unit": true, "applicant": true, "applicant_contact": true,
		"project": true, "apply_reason": true, "apply_config": true,
		"apply_time": true, "object_storage_size": true, "remark": true,
	}

	// Separate host fields and application fields
	hostUpdates := map[string]interface{}{}
	appUpdates := map[string]interface{}{}
	for k, v := range req.Data {
		if appFields[k] {
			appUpdates[k] = v
		} else {
			hostUpdates[k] = v
		}
	}

	// 申请时间严格 YYYY-MM-DD（空串合法），null 与省略同义；错误信息含格式提示
	if v, ok := appUpdates["apply_time"]; ok {
		if v == nil {
			delete(appUpdates, "apply_time")
		} else {
			sv, isStr := v.(string)
			if !isStr {
				return fmt.Errorf("%w: 申请时间格式应为 YYYY-MM-DD", utils.ErrInvalidDate)
			}
			sv = strings.TrimSpace(sv)
			if err := utils.ValidateDateOnly("申请时间", sv); err != nil {
				return err
			}
			appUpdates["apply_time"] = sv
		}
	}

	// person_id 归一化为 *uint（nil 表示解除关联），并校验人员存在
	if v, ok := hostUpdates["person_id"]; ok {
		personID, err := parsePersonID(v)
		if err != nil {
			return err
		}
		if err := validatePersonID(personID); err != nil {
			return err
		}
		hostUpdates["person_id"] = personID
	}

	tx := database.DB.Begin()

	if len(hostUpdates) > 0 {
		if err := tx.Model(&models.Host{}).Where("id IN ?", req.IDS).Updates(hostUpdates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(appUpdates) > 0 {
		tx.Model(&models.HostApplication{}).Where("host_id IN ?", req.IDS).Updates(appUpdates)
	}

	tx.Commit()
	NewAuditService().Record(op, "batch_update", "host", nil,
		DetailDiff{After: map[string]interface{}{"ids": req.IDS, "data": req.Data}}, requestID)
	return nil
}
