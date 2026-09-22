package models

type HostApplication struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	HostID            uint   `json:"host_id" gorm:"not null;uniqueIndex"`
	ApplyUnit         string `json:"apply_unit" gorm:"type:varchar(128)"`
	Applicant         string `json:"applicant" gorm:"type:varchar(64)"`
	ApplicantContact  string `json:"applicant_contact" gorm:"type:varchar(64)"`
	Project           string `json:"project" gorm:"type:varchar(128)"`
	ApplyReason       string `json:"apply_reason" gorm:"type:text"`
	ApplyConfig       string `json:"apply_config" gorm:"type:text"`
	ApplyTime         string `json:"apply_time" gorm:"type:varchar(32)"`
	ObjectStorageSize string `json:"object_storage_size" gorm:"type:varchar(32)"`
	Remark            string `json:"remark" gorm:"type:text"`
}
