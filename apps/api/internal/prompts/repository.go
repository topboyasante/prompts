package prompts

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, p *Prompt, tags []string) (*Prompt, error)
	FindByOwnerAndName(ctx context.Context, ownerUsername, name string) (*Prompt, error)
	List(ctx context.Context, limit, offset int) ([]Prompt, error)
	Search(ctx context.Context, query string, limit, offset int) ([]Prompt, error)
	FindByOwnerUsername(ctx context.Context, ownerUsername string) ([]Prompt, error)
	FindByID(ctx context.Context, id string) (*Prompt, error)
	GetTags(ctx context.Context, promptID string) ([]string, error)
	Delete(ctx context.Context, id string) error
}

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, p *Prompt, tags []string) (*Prompt, error) {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		for _, tag := range tags {
			if err := tx.Create(&PromptTag{PromptID: p.ID, Tag: tag}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	p.Tags = tags
	return p, nil
}

func (r *GORMRepository) FindByOwnerAndName(ctx context.Context, ownerUsername, name string) (*Prompt, error) {
	var prompt Prompt
	err := r.db.WithContext(ctx).
		Table("prompts p").
		Select("p.id, p.name, p.description, p.owner_id, p.created_at, u.username as owner_username").
		Joins("join users u on u.id = p.owner_id").
		Where("u.username = ? and p.name = ?", ownerUsername, name).
		First(&prompt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	tags, err := r.GetTags(ctx, prompt.ID)
	if err != nil {
		return nil, err
	}
	prompt.Tags = tags
	return &prompt, nil
}

func (r *GORMRepository) List(ctx context.Context, limit, offset int) ([]Prompt, error) {
	rows := make([]Prompt, 0)
	err := r.db.WithContext(ctx).
		Table("prompts p").
		Select("p.id, p.name, p.description, p.owner_id, p.created_at, u.username as owner_username").
		Joins("join users u on u.id = p.owner_id").
		Order("p.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GORMRepository) Search(ctx context.Context, query string, limit, offset int) ([]Prompt, error) {
	rows := make([]Prompt, 0)
	sql := `
SELECT p.id, p.name, p.description, p.owner_id, p.created_at, u.username AS owner_username
FROM prompts p JOIN users u ON u.id = p.owner_id
WHERE p.name ILIKE '%' || ? || '%'
   OR p.description ILIKE '%' || ? || '%'
   OR EXISTS (SELECT 1 FROM prompt_tags pt WHERE pt.prompt_id = p.id AND pt.tag ILIKE '%' || ? || '%')
ORDER BY similarity(p.name, ?) DESC
LIMIT ? OFFSET ?;`
	err := r.db.WithContext(ctx).Raw(sql, query, query, query, query, limit, offset).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GORMRepository) FindByOwnerUsername(ctx context.Context, ownerUsername string) ([]Prompt, error) {
	rows := make([]Prompt, 0)
	err := r.db.WithContext(ctx).
		Table("prompts p").
		Select("p.id, p.name, p.description, p.owner_id, p.created_at, u.username as owner_username").
		Joins("join users u on u.id = p.owner_id").
		Where("u.username = ?", ownerUsername).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GORMRepository) FindByID(ctx context.Context, id string) (*Prompt, error) {
	var prompt Prompt
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&prompt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &prompt, nil
}

func (r *GORMRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Prompt{}).Error
}

func (r *GORMRepository) GetTags(ctx context.Context, promptID string) ([]string, error) {
	rows := make([]PromptTag, 0)
	err := r.db.WithContext(ctx).Where("prompt_id = ?", promptID).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	tags := make([]string, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, row.Tag)
	}
	return tags, nil
}
