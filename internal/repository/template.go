package repository

import (
	"github.com/devblin/tuskira/internal/model"
	"gorm.io/gorm"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(t *model.Template) error {
	return r.db.Create(t).Error
}

func (r *TemplateRepository) FindByID(id uint) (*model.Template, error) {
	var t model.Template
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepository) FindAll() ([]model.Template, error) {
	var templates []model.Template
	err := r.db.Find(&templates).Error
	return templates, err
}
