package services

import (
	"mcloud/database"
	"mcloud/models"
)

// 区域资源使用聚合（T-087）：资源统计页原先前端按区域串行分页拉全量主机明细再汇总，
// 改为后端单条 SQL 一次聚合，避免「区域数 × 分页数」次请求。

// RegionUsageStat 单区域按状态拆分的资源用量
type RegionUsageStat struct {
	// 运行中
	RunningCPU       int64 `json:"running_cpu"`
	RunningMemory    int64 `json:"running_memory"`
	RunningStorage   int64 `json:"running_storage"`
	RunningBareMetal int64 `json:"running_bare_metal"`
	// 已停止（含「已关机」，与统计口径一致）
	StoppedCPU       int64 `json:"stopped_cpu"`
	StoppedMemory    int64 `json:"stopped_memory"`
	StoppedStorage   int64 `json:"stopped_storage"`
	StoppedBareMetal int64 `json:"stopped_bare_metal"`
}

// regionUsageRow SQL 聚合结果行
type regionUsageRow struct {
	Region           string
	RunningCPU       int64
	RunningMemory    int64
	RunningStorage   int64
	RunningBareMetal int64
	StoppedCPU       int64
	StoppedMemory    int64
	StoppedStorage   int64
	StoppedBareMetal int64
}

// RegionUsage 各区域资源用量（键为区域名）。
// 口径与前端原有实现一致：排除云外主机（T-078）；CPU/内存/存储只计虚拟机
// （裸金属服务器与物理机不计，T-079）；裸金属台数单独统计。
func (s *StatsService) RegionUsage() (map[string]RegionUsageStat, error) {
	var rows []regionUsageRow
	err := database.DB.Model(&models.Host{}).
		Select(`
			region,
			COALESCE(SUM(CASE WHEN status = '运行中' AND COALESCE(asset_type,'') NOT IN ('裸金属服务器','物理机') THEN cpu ELSE 0 END), 0) AS running_cpu,
			COALESCE(SUM(CASE WHEN status = '运行中' AND COALESCE(asset_type,'') NOT IN ('裸金属服务器','物理机') THEN memory ELSE 0 END), 0) AS running_memory,
			COALESCE(SUM(CASE WHEN status = '运行中' AND COALESCE(asset_type,'') NOT IN ('裸金属服务器','物理机') THEN COALESCE(system_disk,0) + COALESCE(data_disk,0) ELSE 0 END), 0) AS running_storage,
			COALESCE(SUM(CASE WHEN status = '运行中' AND asset_type = '裸金属服务器' THEN 1 ELSE 0 END), 0) AS running_bare_metal,
			COALESCE(SUM(CASE WHEN status IN ('已停止','已关机') AND COALESCE(asset_type,'') NOT IN ('裸金属服务器','物理机') THEN cpu ELSE 0 END), 0) AS stopped_cpu,
			COALESCE(SUM(CASE WHEN status IN ('已停止','已关机') AND COALESCE(asset_type,'') NOT IN ('裸金属服务器','物理机') THEN memory ELSE 0 END), 0) AS stopped_memory,
			COALESCE(SUM(CASE WHEN status IN ('已停止','已关机') AND COALESCE(asset_type,'') NOT IN ('裸金属服务器','物理机') THEN COALESCE(system_disk,0) + COALESCE(data_disk,0) ELSE 0 END), 0) AS stopped_storage,
			COALESCE(SUM(CASE WHEN status IN ('已停止','已关机') AND asset_type = '裸金属服务器' THEN 1 ELSE 0 END), 0) AS stopped_bare_metal
		`).
		Where("is_external = ?", false).
		Group("region").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make(map[string]RegionUsageStat, len(rows))
	for _, r := range rows {
		out[r.Region] = RegionUsageStat{
			RunningCPU: r.RunningCPU, RunningMemory: r.RunningMemory,
			RunningStorage: r.RunningStorage, RunningBareMetal: r.RunningBareMetal,
			StoppedCPU: r.StoppedCPU, StoppedMemory: r.StoppedMemory,
			StoppedStorage: r.StoppedStorage, StoppedBareMetal: r.StoppedBareMetal,
		}
	}
	return out, nil
}
