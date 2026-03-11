package identities

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	FindByProviderUserID(ctx context.Context, provider, providerUserID string) (*UserIdentity, error)
	Create(ctx context.Context, identity *UserIdentity) (*UserIdentity, error)
	FindByUserAndProvider(ctx context.Context, userID, provider string) (*UserIdentity, error)
}

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) FindByProviderUserID(ctx context.Context, provider, providerUserID string) (*UserIdentity, error) {
	var identity UserIdentity
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&identity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *GORMRepository) Create(ctx context.Context, identity *UserIdentity) (*UserIdentity, error) {
	if err := r.db.WithContext(ctx).Create(identity).Error; err != nil {
		return nil, err
	}
	return identity, nil
}

func (r *GORMRepository) FindByUserAndProvider(ctx context.Context, userID, provider string) (*UserIdentity, error) {
	var identity UserIdentity
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&identity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &identity, nil
}
