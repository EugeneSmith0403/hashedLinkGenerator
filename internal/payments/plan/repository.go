package plan

import (
	"link-generator/pkg/db"
	"errors"

	"gorm.io/gorm"
)

type PlanRepository struct {
	db *db.Db
}

func NewPlanRepository(db *db.Db) *PlanRepository {
	return &PlanRepository{db}
}

func (r *PlanRepository) GetAll() ([]*Plan, error) {
	var plans []*Plan
	result := r.db.DB.Where("is_active = ?", true).Find(&plans)
	if result.Error != nil {
		return nil, result.Error
	}
	return plans, nil
}

func (r *PlanRepository) GetByID(id uint) (*Plan, error) {
	var plan Plan
	result := r.db.DB.First(&plan, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &plan, nil
}
