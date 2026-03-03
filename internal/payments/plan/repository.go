package plan

import (
	"adv/go-http/pkg/db"
	"errors"

	"gorm.io/gorm"
)

type PlanRepository struct {
	db *db.Db
}

func NewPlanRepository(db *db.Db) *PlanRepository {
	return &PlanRepository{db}
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
