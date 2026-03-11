package versions

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, pv *PromptVersion) (*PromptVersion, error)
	FindByPromptID(ctx context.Context, promptID string) ([]PromptVersion, error)
	FindByVersion(ctx context.Context, promptID, version string) (*PromptVersion, error)
	FindLatestByPromptID(ctx context.Context, promptID string) (*PromptVersion, error)
	RecordDownload(ctx context.Context, promptID, versionID string) error
}

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, pv *PromptVersion) (*PromptVersion, error) {
	if err := r.db.WithContext(ctx).Create(pv).Error; err != nil {
		return nil, err
	}
	return pv, nil
}

func (r *GORMRepository) FindByPromptID(ctx context.Context, promptID string) ([]PromptVersion, error) {
	rows := make([]PromptVersion, 0)
	err := r.db.WithContext(ctx).Where("prompt_id = ?", promptID).Order("created_at desc").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GORMRepository) FindByVersion(ctx context.Context, promptID, version string) (*PromptVersion, error) {
	var row PromptVersion
	err := r.db.WithContext(ctx).Where("prompt_id = ? and version = ?", promptID, version).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *GORMRepository) FindLatestByPromptID(ctx context.Context, promptID string) (*PromptVersion, error) {
	var row PromptVersion
	err := r.db.WithContext(ctx).Where("prompt_id = ?", promptID).Order("created_at desc").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *GORMRepository) RecordDownload(ctx context.Context, promptID, versionID string) error {
	return r.db.WithContext(ctx).Create(&Download{PromptID: promptID, VersionID: versionID}).Error
}
