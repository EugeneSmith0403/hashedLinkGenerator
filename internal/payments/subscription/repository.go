package subscription

import (
	"adv/go-http/pkg/db"
	"errors"
	"time"

	"gorm.io/gorm"
)

type SubscriptionUpdate struct {
	Status             SubscriptionStatus `gorm:"column:status"`
	CurrentPeriodStart time.Time          `gorm:"column:current_period_start"`
	CurrentPeriodEnd   time.Time          `gorm:"column:current_period_end"`
	CancelAt           *time.Time         `gorm:"column:cancel_at"`
	CanceledAt         *time.Time         `gorm:"column:canceled_at"`
	TrialStart         *time.Time         `gorm:"column:trial_start"`
	TrialEnd           *time.Time         `gorm:"column:trial_end"`
}

type SubscriptionRepository struct {
	db *db.Db
}

func NewSubscriptionRepository(db *db.Db) *SubscriptionRepository {
	return &SubscriptionRepository{db}
}

func (r *SubscriptionRepository) Create(s *Subscription) (*Subscription, error) {
	result := r.db.DB.Create(s)
	if result.Error != nil {
		return nil, result.Error
	}
	return s, nil
}

func (r *SubscriptionRepository) GetByID(id uint) (*Subscription, error) {
	var subscription Subscription
	result := r.db.DB.First(&subscription, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) GetByBillingID(billingID string) (*Subscription, error) {
	var subscription Subscription
	result := r.db.DB.First(&subscription, "billing_id = ?", billingID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) GetActiveByUserID(userID uint) (*Subscription, error) {
	var subscription Subscription
	result := r.db.DB.
		Where("user_id = ? AND status IN ?", userID, []SubscriptionStatus{
			SubscriptionStatusActive,
			SubscriptionStatusTrialing,
		}).
		First(&subscription)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) Update(id uint, upd *SubscriptionUpdate) (*Subscription, error) {
	result := r.db.DB.Model(&Subscription{}).Where("id = ?", id).
		Select("status", "current_period_start", "current_period_end", "cancel_at", "canceled_at", "trial_start", "trial_end").
		Updates(upd)
	if result.Error != nil {
		return nil, result.Error
	}

	return r.GetByID(id)
}

func (r *SubscriptionRepository) UpdateByBillingId(billingId string, upd *SubscriptionUpdate) (*Subscription, error) {
	result := r.db.DB.Model(&Subscription{}).Where("billing_id = ?", billingId).
		Select("status", "current_period_start", "current_period_end", "cancel_at", "canceled_at", "trial_start", "trial_end").
		Updates(upd)
	if result.Error != nil {
		return nil, result.Error
	}

	return r.GetByBillingID(billingId)
}

func (r *SubscriptionRepository) GetSubscriptionByUserId(userId uint) (*Subscription, error) {
	var sub Subscription

	result := r.db.DB.
		Where("user_id = ? AND status IN ?", userId, []SubscriptionStatus{
			SubscriptionStatusActive,
			SubscriptionStatusTrialing,
		}).
		First(&sub)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &sub, nil
}

func (r *SubscriptionRepository) GetSubscriptionByEmail(email string) (*Subscription, error) {
	var sub Subscription

	result := r.db.DB.
		Model(&Subscription{}).
		Joins("JOIN users ON users.id = subscriptions.user_id").
		Where("users.email = ?", email).
		Where("subscriptions.status = ?", SubscriptionStatusActive).
		First(&sub)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &sub, nil
}
