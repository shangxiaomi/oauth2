package service

import (
	"fmt"
	"net/http"
	mylog "oauth2/log"

	"golang.org/x/crypto/bcrypt"
	"oauth2/model"
	"oauth2/repository"
)

type IUserService interface {
	GetUserIdByPwd(email string, password string) (string, error)
	ValidatePassword(email string, password string) bool
	CreateUser(email string, password string, name string) (user *model.User, msg string, code int, err error)
	GetUserInfoByEmail(email string) (*model.User, error)
}

type UserService struct {
	repo repository.IUserRepo
}

func (u UserService) GetUserInfoByEmail(email string) (*model.User, error) {
	return u.repo.SelectByEmail(email)
}

// 需要保证传入的参数是合法的
func (u UserService) CreateUser(email string, password string, name string) (user *model.User, msg string, code int, err error) {
	// 生成加密的密码
	hashPassword, err := getPassword(password)
	if err != nil {
		mylog.Error.Println(fmt.Sprintf("密码加密时出错 %v", err))
		return nil, "内部系统错误", http.StatusInternalServerError, err
	}
	// 创建新的用户对象
	newUser := model.User{
		Name:     name,
		Email:    email,
		Password: hashPassword,
		Active:   0,
	}
	// 将用户插入到数据库中
	user, err = u.repo.SelectByEmail(email)

	if user.ID != 0 {
		return nil, "此email已经被注册", http.StatusUnprocessableEntity, nil
	}

	user, err = u.repo.Create(newUser)
	if err != nil {
		mylog.Error.Println(fmt.Sprintf("数据创建时出错: %v", err))
		return nil, "内部系统错误", http.StatusInternalServerError, err
	}

	return user, "创建成功", http.StatusOK, nil
}

func (u UserService) ValidatePassword(email string, password string) bool {
	user, err := u.repo.SelectByEmail(email)
	if err != nil {
		mylog.Warn.Println(fmt.Sprintf("查询数据时出错 %v", err.Error()))
		return false
	}
	return validatePassword(user, password)
}

func (u UserService) GetUserIdByPwd(email string, password string) (string, error) {
	// 验证用户数据的正确性
	flag := u.ValidatePassword(email, password)
	if flag == false {
		return "", nil
	}
	user, err := u.repo.SelectByEmail(email)
	if user == nil {
		mylog.Info.Println("根据email查询到的user为空")
		// TODO 一个保底逻辑，看是不是抛出异常
		return "", nil
	}
	return fmt.Sprintf("%d", user.ID), err
}

func getPassword(password string) (string, error) {
	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bcryptPassword), nil
}

// 验证密码的正确正确性，返回 true 说明用户名和密码正确
func validatePassword(user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func NewUserService() IUserService {
	return UserService{
		repo: repository.NewUserRepo(),
	}
}
