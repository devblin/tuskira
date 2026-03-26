package repository

import (
	"github.com/devblin/tuskira/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

func (r *NotificationRepository) Save(n *model.Notification) error {
	return r.db.Save(n).Error
}

func (r *NotificationRepository) FindByID(id uint) (*model.Notification, error) {
	var n model.Notification
	if err := r.db.Preload("Template").First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepository) FindByRecipient(recipient string, userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Where("recipient = ? AND user_id = ?", recipient, userID).Order("created_at desc").Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) FindPendingScheduled(userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Where("status = ? AND user_id = ?", model.StatusScheduled, userID).Order("schedule_at asc").Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) FindPending(userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Where("status = ? AND user_id = ?", model.StatusPending, userID).Order("created_at asc").Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) FindSent(userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Where("status IN ? AND user_id = ?", []model.Status{model.StatusSent, model.StatusFailed}, userID).Order("created_at desc").Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) FindPendingByRecipientAndChannel(recipient string, channel model.Channel) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.Where("recipient = ? AND channel = ? AND status IN ?", recipient, channel, []model.Status{model.StatusPending, model.StatusFailed}).
		Order("created_at asc").Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) FindTemplateByID(id uint) (*model.Template, error) {
	var t model.Template
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
