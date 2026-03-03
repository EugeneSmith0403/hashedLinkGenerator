package link

import (
	"adv/go-http/internal/models"
	"adv/go-http/pkg/db"
)

type LinkRepository struct {
	db *db.Db
}

func NewLinkRepository(db *db.Db) *LinkRepository {
	return &LinkRepository{
		db,
	}
}

func (r LinkRepository) Create(link *models.Link) (*models.Link, error) {

	result := r.db.DB.Create(link)

	if result.Error != nil {
		return nil, result.Error
	}

	return link, nil

}

func (r LinkRepository) Update(link *models.Link) (*models.Link, error) {

	result := r.db.Updates(link)

	if result.Error != nil {
		return nil, result.Error
	}

	return link, nil

}

func (r LinkRepository) Delete(link *models.Link) (*models.Link, error) {
	result := r.db.Delete(link)

	if result.Error != nil {
		return nil, result.Error
	}

	return link, nil
}

func (r LinkRepository) GetByHash(hash string) (*models.Link, error) {
	var link models.Link
	result := r.db.DB.First(&link, "hash = ?", hash)

	if result.Error != nil {
		return nil, result.Error
	}

	return &link, nil

}

func (r *LinkRepository) GetByUserID(userID uint) ([]*models.Link, error) {
	var links []*models.Link
	result := r.db.DB.Where("user_id = ?", userID).Find(&links)
	if result.Error != nil {
		return nil, result.Error
	}
	return links, nil
}

func (r *LinkRepository) getById(id uint) (*models.Link, error) {
	var link models.Link
	result := r.db.DB.First(&link, id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &link, nil

}
