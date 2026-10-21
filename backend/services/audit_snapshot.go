package services

import "mcloud/models"

// 资产变更快照：before/after 使用同一套字段键，供前端时间线逐字段对比展示（T-066）。
// 键名与模型字段一致；前端维护中文标签映射。

// hostSnapshot 组装主机快照（含一对一申请信息），字段键统一。
func hostSnapshot(h models.Host, app *models.HostApplication) map[string]interface{} {
	snap := map[string]interface{}{
		"region":       h.Region,
		"instance_id":  h.InstanceID,
		"name":         h.Name,
		"private_ip":   h.PrivateIP,
		"asset_type":   h.AssetType,
		"os":           h.OS,
		"cpu":          h.CPU,
		"cpu_arch":     h.CPUArch,
		"memory":       h.Memory,
		"system_disk":  h.SystemDisk,
		"data_disk":    h.DataDisk,
		"env_type":     h.EnvType,
		"is_db_server": h.IsDBServer,
		"status":       h.Status,
		"tags":         h.Tags,
		"person_id":    personIDValue(h.PersonID),
	}
	if app != nil {
		snap["apply_unit"] = app.ApplyUnit
		snap["applicant"] = app.Applicant
		snap["applicant_contact"] = app.ApplicantContact
		snap["project"] = app.Project
		snap["apply_reason"] = app.ApplyReason
		snap["apply_config"] = app.ApplyConfig
		snap["apply_time"] = app.ApplyTime
		snap["remark"] = app.Remark
	}
	return snap
}

// zeroTrustSnapshot 组装零信任快照（targets 为派生配对串，字段键统一）。
func zeroTrustSnapshot(z models.ZeroTrust, pairs []ZeroTrustTarget) map[string]interface{} {
	return map[string]interface{}{
		"apply_unit":      z.ApplyUnit,
		"account_name":    z.AccountName,
		"account_name_zh": z.AccountNameZh,
		"applicant":       z.Applicant,
		"contact":         z.Contact,
		"exit_location":   z.ExitLocation,
		"targets":         joinTargetPairs(pairs),
		"system_name":     z.SystemName,
		"apply_time":      z.ApplyTime,
		"remark":          z.Remark,
	}
}

// portMappingSnapshot 组装端口映射快照，字段键统一。
func portMappingSnapshot(m models.PortMapping) map[string]interface{} {
	return map[string]interface{}{
		"public_ip":          m.PublicIP,
		"host_id":            m.HostID,
		"external_ports":     m.ExternalPorts,
		"external_protocols": m.ExternalProtocols,
		"internal_ports":     m.InternalPorts,
		"internal_protocols": m.InternalProtocols,
		"domain":             m.Domain,
		"remark":             m.Remark,
	}
}

func personIDValue(p *uint) interface{} {
	if p == nil {
		return nil
	}
	return *p
}
