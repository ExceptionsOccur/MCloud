package services

import (
	"errors"
	"fmt"
	"strings"

	"mcloud/database"
	"mcloud/models"
)

type PersonService struct{}

func NewPersonService() *PersonService {
	return &PersonService{}
}

type PersonRequest struct {
	Name    string `json:"name" binding:"required"`
	Contact string `json:"contact"`
	Unit    string `json:"unit"`
}

type PersonItem struct {
	models.Person
	HostCount int64 `json:"host_count"`
}

// ErrPersonReferenced 人员已被主机引用，禁止删除
var ErrPersonReferenced = errors.New("该人员已被主机引用")

// ErrPersonNotFound 人员不存在
var ErrPersonNotFound = errors.New("人员不存在")

func normalizePerson(req PersonRequest) (PersonRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Contact = strings.TrimSpace(req.Contact)
	req.Unit = strings.TrimSpace(req.Unit)
	if req.Name == "" {
		return req, errors.New("姓名不能为空")
	}
	return req, nil
}

func (s *PersonService) List(keyword string) ([]PersonItem, error) {
	db := database.DB.Model(&models.Person{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name ILIKE ? OR contact ILIKE ? OR unit ILIKE ?", like, like, like)
	}

	var persons []models.Person
	if err := db.Order("id ASC").Find(&persons).Error; err != nil {
		return nil, err
	}

	items := make([]PersonItem, 0, len(persons))
	for _, p := range persons {
		var count int64
		database.DB.Model(&models.Host{}).Where("person_id = ?", p.ID).Count(&count)
		items = append(items, PersonItem{Person: p, HostCount: count})
	}
	return items, nil
}

func (s *PersonService) GetByID(id uint) (*models.Person, error) {
	var person models.Person
	if err := database.DB.First(&person, id).Error; err != nil {
		return nil, ErrPersonNotFound
	}
	return &person, nil
}

func (s *PersonService) Create(req PersonRequest) (uint, error) {
	req, err := normalizePerson(req)
	if err != nil {
		return 0, err
	}

	var count int64
	database.DB.Model(&models.Person{}).
		Where("name = ? AND contact = ? AND unit = ?", req.Name, req.Contact, req.Unit).
		Count(&count)
	if count > 0 {
		return 0, errors.New("该人员已存在")
	}

	person := models.Person{Name: req.Name, Contact: req.Contact, Unit: req.Unit}
	if err := database.DB.Create(&person).Error; err != nil {
		return 0, err
	}
	return person.ID, nil
}

func (s *PersonService) Update(id uint, req PersonRequest) error {
	req, err := normalizePerson(req)
	if err != nil {
		return err
	}

	var person models.Person
	if err := database.DB.First(&person, id).Error; err != nil {
		return ErrPersonNotFound
	}

	var count int64
	database.DB.Model(&models.Person{}).
		Where("name = ? AND contact = ? AND unit = ? AND id != ?", req.Name, req.Contact, req.Unit, id).
		Count(&count)
	if count > 0 {
		return errors.New("该人员已存在")
	}

	person.Name = req.Name
	person.Contact = req.Contact
	person.Unit = req.Unit
	return database.DB.Save(&person).Error
}

func (s *PersonService) Delete(id uint) error {
	var person models.Person
	if err := database.DB.First(&person, id).Error; err != nil {
		return ErrPersonNotFound
	}

	var count int64
	database.DB.Model(&models.Host{}).Where("person_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("%w，%d 台主机正在使用", ErrPersonReferenced, count)
	}
	return database.DB.Delete(&person).Error
}
