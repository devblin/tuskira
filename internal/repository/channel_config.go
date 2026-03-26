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
	err := r.db.Where("channel = ? AND user_id = ?", cfg.Channel, cfg.UserID).First(&existing).Error
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

func (r *ChannelConfigRepository) FindByChannel(channel model.Channel, userID uint) (*model.ChannelConfig, error) {
	var cfg model.ChannelConfig
	if err := r.db.Where("channel = ? AND user_id = ?", channel, userID).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *ChannelConfigRepository) FindAll(userID uint) ([]model.ChannelConfig, error) {
	var configs []model.ChannelConfig
	err := r.db.Where("user_id = ?", userID).Order("channel asc").Find(&configs).Error
	return configs, err
}

func (r *ChannelConfigRepository) Delete(channel model.Channel, userID uint) error {
	return r.db.Where("channel = ? AND user_id = ?", channel, userID).Delete(&model.ChannelConfig{}).Error
}
