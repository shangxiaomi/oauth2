package model

import "time"

type User struct {
	ID        int64 `gorm:"type:bigint;primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string `json:"name" gorm:"type:varchar(100);not null"`
	Email     string `json:"email" gorm:"type:varchar(255);not null;unique"`
	Password  string `json:"password" gorm:"type:varchar(255);not null;"`
	Active    int    `json:"active;not null;"`
	ActiveUrl string `json:"activeUrl" gorm:"type:varchar(255)"`
}

func (u *User) TableName() string {
	return "user"
}