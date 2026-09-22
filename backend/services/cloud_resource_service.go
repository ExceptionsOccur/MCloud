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

func (s *CloudResourceService) Update(req UpdateCloudResourceRequest) error {
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
		return database.DB.Create(&resource).Error
	}

	resource.PhysicalCPU = req.PhysicalCPU
	resource.Vcpu = req.Vcpu
	resource.Memory = req.Memory
	resource.Storage = req.Storage
	resource.BareMetal = req.BareMetal
	resource.GpuCardCount = req.GpuCardCount
	resource.ObjectStorage = req.ObjectStorage
	return database.DB.Save(&resource).Error
}
