package repository

import (
	"github.com/devblin/tuskira/internal/model"
	"gorm.io/gorm"
)

type ChannelConfigRepository struct {
	db *gorm.DB
}

func NewChannelConfigRepository(db *gorm.DB) *ChannelConfigRepository {
	return &ChannelConfigRepository{db: db}
}

func (r *ChannelConfigRepository) Upsert(cfg *model.ChannelConfig) error {
	var existing model.ChannelConfig
	err := r.db.Where("channel = ?", cfg.Channel).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(cfg).Error
	}
	if err != nil {
		return err
	}
	existing.Enabled = cfg.Enabled
	existing.Config = cfg.Config
	return r.db.Save(&existing).Error
}

func (r *ChannelConfigRepository) FindByChannel(channel model.Channel) (*model.ChannelConfig, error) {
	var cfg model.ChannelConfig
	if err := r.db.Where("channel = ?", channel).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *ChannelConfigRepository) FindAll() ([]model.ChannelConfig, error) {
	var configs []model.ChannelConfig
	err := r.db.Order("channel asc").Find(&configs).Error
	return configs, err
}

func (r *ChannelConfigRepository) Delete(channel model.Channel) error {
	return r.db.Where("channel = ?", channel).Delete(&model.ChannelConfig{}).Error
}
