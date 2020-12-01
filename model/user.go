package model

import "time"

//type User struct {
//	ID int `gorm:"primary_key" json:"id"`
//	Name string `json:"name"`
//	Password string `json:"password"`
//}

type User struct {
	ID        uint `gorm:"primarykey"`
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

//func (u *User) GetUserIDByPwd(username, password string) (userID string) {
//	// use the db conn
//	// write your own user authentication logic
//	// like:
//	// db.Where("name = ? AND password = ?", username, password).First(u)
//	// userID = u.ID
//	// test account: admin admin
//	if username == "admin" && password == "admin" {
//		userID = "admin"
//	}
//	return
//}
