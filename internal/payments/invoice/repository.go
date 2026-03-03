package invoice

import (
	"adv/go-http/pkg/db"
	"errors"

	"gorm.io/gorm"
)

type InvoiceRepository struct {
	db *db.Db
}

func NewInvoiceRepository(db *db.Db) *InvoiceRepository {
	return &InvoiceRepository{db}
}

func (r *InvoiceRepository) Create(inv *Invoice) (*Invoice, error) {
	result := r.db.DB.Create(inv)
	if result.Error != nil {
		return nil, result.Error
	}
	return inv, nil
}

func (r *InvoiceRepository) GetByID(id uint) (*Invoice, error) {
	var inv Invoice
	result := r.db.DB.First(&inv, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &inv, nil
}

func (r *InvoiceRepository) Update(inv *Invoice) (*Invoice, error) {
	result := r.db.DB.Save(inv)
	return inv, result.Error
}

func (r *InvoiceRepository) GetByBillingID(billingID string) (*Invoice, error) {
	var inv Invoice
	result := r.db.DB.First(&inv, "billing_id = ?", billingID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &inv, nil
}
