package identities

import "time"

type UserIdentity struct {
	ID             string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID         string    `gorm:"column:user_id" json:"user_id"`
	Provider       string    `gorm:"column:provider" json:"provider"`
	ProviderUserID string    `gorm:"column:provider_user_id" json:"provider_user_id"`
	Email          string    `gorm:"column:email" json:"email,omitempty"`
	EmailVerified  bool      `gorm:"column:email_verified" json:"email_verified"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (UserIdentity) TableName() string {
	return "user_identities"
}
