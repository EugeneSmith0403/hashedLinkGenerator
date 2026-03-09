package payment

import (
	"errors"

	paymentmodels "adv/go-http/internal/payments/models"
	"adv/go-http/pkg/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *db.Db
}

func NewPaymentRepository(db *db.Db) *PaymentRepository {
	return &PaymentRepository{db}
}

func (r *PaymentRepository) Create(p *paymentmodels.Payment) (*paymentmodels.Payment, error) {
	result := r.db.DB.Create(p)
	if result.Error != nil {
		return nil, result.Error
	}
	return p, nil
}

func (r *PaymentRepository) Save(p *paymentmodels.Payment) (*paymentmodels.Payment, error) {
	existing, err := r.GetByUuid(p.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		result := r.db.DB.Create(p)
		return p, result.Error
	}
	result := r.db.DB.Save(p)
	return p, result.Error
}

func (r *PaymentRepository) LinkInvoice(paymentID uuid.UUID, invoiceID uint) error {
	return r.db.DB.Model(&paymentmodels.Payment{}).
		Where("id = ?", paymentID).
		Update("invoice_id", invoiceID).Error
}

func (r *PaymentRepository) LinkSubscriptionByPI(piID string, subscriptionID uint) error {
	return r.db.DB.Model(&paymentmodels.Payment{}).
		Where("payment_intent_id = ?", piID).
		Update("subscription_id", subscriptionID).Error
}

func (r *PaymentRepository) GetByUuid(id uuid.UUID) (*paymentmodels.Payment, error) {
	var p paymentmodels.Payment
	result := r.db.DB.First(&p, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &p, nil
}

func (r *PaymentRepository) GetByPaymentIntentID(piID string) (*paymentmodels.Payment, error) {
	var p paymentmodels.Payment
	result := r.db.DB.First(&p, "payment_intent_id = ?", piID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &p, nil
}

func (r *PaymentRepository) CancelBySubscriptionID(subscriptionID uint) error {
	return r.db.DB.Model(&paymentmodels.Payment{}).
		Where("subscription_id = ?", subscriptionID).
		Update("status", "canceled").Error
}

func (r *PaymentRepository) GetByAccountID(accountID uint) ([]paymentmodels.Payment, error) {
	var payments []paymentmodels.Payment
	result := r.db.DB.Where("account_id = ?", accountID).Order("created_at DESC").Find(&payments)
	if result.Error != nil {
		return nil, result.Error
	}
	return payments, nil
}
