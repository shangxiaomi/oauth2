package repository

import (
	"gorm.io/gorm"
	"oauth2/common"
	"oauth2/model"
)

type IUserRepo interface {
	Create(user model.User) (*model.User, error)
	Update(User model.User, updateParams map[string]interface{}) (*model.User, error)
	SelectById(id int) (*model.User, error)
	SelectByEmail(email string) (*model.User, error)
	DeleteById(id int) error
}

type UserRepo struct {
	DB *gorm.DB
}

func (u UserRepo) SelectByEmail(email string) (*model.User, error) {
	var user model.User
	tx := u.DB.Where("email = ?", email).First(&user)
	if tx.Error != nil {
		return &user, tx.Error
	}
	return &user, nil
}

func (u UserRepo) Create(user model.User) (*model.User, error) {
	create := u.DB.Create(&user)
	if create.Error != nil {
		return nil, create.Error
	}
	return &user, nil
}

func (u UserRepo) Update(user model.User, updateParams map[string]interface{}) (*model.User, error) {
	updates := u.DB.Model(&user).Updates(updateParams)
	if updates.Error != nil {
		return nil, updates.Error
	}
	return &user, nil
}

func (u UserRepo) SelectById(id int) (*model.User, error) {
	var user model.User
	tx := u.DB.Select(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func (u UserRepo) DeleteById(id int) error {
	panic("implement me")
}

func NewUserRepo() IUserRepo {
	return UserRepo{DB: common.GetDB()}
}
