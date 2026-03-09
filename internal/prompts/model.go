package prompts

import "time"

type Prompt struct {
	ID            string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name          string    `gorm:"column:name" json:"name"`
	Description   string    `gorm:"column:description" json:"description"`
	OwnerID       string    `gorm:"column:owner_id" json:"owner_id"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	OwnerUsername string    `gorm:"-" json:"owner_username,omitempty"`
	Tags          []string  `gorm:"-" json:"tags,omitempty"`
}

func (Prompt) TableName() string {
	return "prompts"
}

type PromptTag struct {
	ID       string `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	PromptID string `gorm:"column:prompt_id"`
	Tag      string `gorm:"column:tag"`
}

func (PromptTag) TableName() string {
	return "prompt_tags"
}
