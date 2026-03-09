package users

import "time"

type User struct {
	ID        string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username  string    `gorm:"column:username" json:"username"`
	Email     string    `gorm:"column:email" json:"email,omitempty"`
	AvatarURL string    `gorm:"column:avatar_url" json:"avatar_url,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}
