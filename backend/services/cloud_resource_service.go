package services

import (
	"mcloud/database"
	"mcloud/models"
)

type CloudResourceService struct{}

func NewCloudResourceService() *CloudResourceService {
	return &CloudResourceService{}
}

func (s *CloudResourceService) List() ([]models.CloudResource, error) {
	var resources []models.CloudResource
	result := database.DB.Order("id ASC").Find(&resources)
	if result.Error != nil {
		return nil, result.Error
	}
	return resources, nil
}

type UpdateCloudResourceRequest struct {
	Region        string `json:"region" binding:"required"`
	PhysicalCPU   int    `json:"physical_cpu"`
	Vcpu          int    `json:"vcpu"`
	Memory        int    `json:"memory"`
	Storage       int    `json:"storage"`
	BareMetal     int    `json:"bare_metal"`
	GpuCardCount  int    `json:"gpu_card_count"`
	ObjectStorage int    `json:"object_storage"`
}

func (s *CloudResourceService) Update(req UpdateCloudResourceRequest, op Operator, requestID string) error {
	var resource models.CloudResource
	result := database.DB.Where("region = ?", req.Region).First(&resource)
	if result.Error != nil {
		resource = models.CloudResource{
			Region:        req.Region,
			PhysicalCPU:   req.PhysicalCPU,
			Vcpu:          req.Vcpu,
			Memory:        req.Memory,
			Storage:       req.Storage,
			BareMetal:     req.BareMetal,
			GpuCardCount:  req.GpuCardCount,
			ObjectStorage: req.ObjectStorage,
		}
		if err := database.DB.Create(&resource).Error; err != nil {
			return err
		}
		NewAuditService().Record(op, "create", "cloud_resource", &resource.ID, DetailDiff{After: req}, requestID)
		return nil
	}

	before := UpdateCloudResourceRequest{
		Region: resource.Region, PhysicalCPU: resource.PhysicalCPU, Vcpu: resource.Vcpu,
		Memory: resource.Memory, Storage: resource.Storage, BareMetal: resource.BareMetal,
		GpuCardCount: resource.GpuCardCount, ObjectStorage: resource.ObjectStorage,
	}
	resource.PhysicalCPU = req.PhysicalCPU
	resource.Vcpu = req.Vcpu
	resource.Memory = req.Memory
	resource.Storage = req.Storage
	resource.BareMetal = req.BareMetal
	resource.GpuCardCount = req.GpuCardCount
	resource.ObjectStorage = req.ObjectStorage
	if err := database.DB.Save(&resource).Error; err != nil {
		return err
	}
	NewAuditService().Record(op, "update", "cloud_resource", &resource.ID,
		DetailDiff{Before: before, After: req}, requestID)
	return nil
}
