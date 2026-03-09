package versions

import "time"

type PromptVersion struct {
	ID         string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	PromptID   string    `gorm:"column:prompt_id" json:"prompt_id"`
	Version    string    `gorm:"column:version" json:"version"`
	TarballURL string    `gorm:"column:tarball_url" json:"tarball_url"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (PromptVersion) TableName() string {
	return "prompt_versions"
}

type Download struct {
	ID           string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	PromptID     string    `gorm:"column:prompt_id"`
	VersionID    string    `gorm:"column:version_id"`
	DownloadedAt time.Time `gorm:"column:downloaded_at"`
}

func (Download) TableName() string {
	return "downloads"
}
