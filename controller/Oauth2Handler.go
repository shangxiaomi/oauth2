package controller

import (
	"gopkg.in/oauth2.v3/errors"
	"log"
	"net/http"
	"oauth2/config"
	"oauth2/pkg/session"
	"oauth2/service"
)

func internalErrorHandler(err error) (re *errors.Response) {
	log.Println("Internal Error:", err.Error())
	return
}
func responseErrorHandler(re *errors.Response) {
	log.Println("Response Error:", re.Error.Error())
}
func authorizeScopeHandler(w http.ResponseWriter, r *http.Request) (scope string, err error) {
	if r.Form == nil {
		r.ParseForm()
	}
	s := config.ScopeFilter(r.Form.Get("client_id"), r.Form.Get("scope"))
	if s == nil {
		http.Error(w, "Invalid Scope", http.StatusBadRequest)
		return
	}
	scope = config.ScopeJoin(s)
	return
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	// 获取会话的UserId
	v, _ := session.Get(r, "LoggedInUserID")
	// 会话不存在
	if v == nil {
		if r.Form == nil {
			r.ParseForm()
		}
		session.Set(w, r, "RequestForm", r.Form)
		// 进行登录验证
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
	userID = v.(string)
	return
}
func passwordAuthorizationHandler(email string, password string) (string, error) {
	userService := service.NewUserService()
	userId, err := userService.GetUserIdByPwd(email, password)
	if err != nil {
		return "", err
	}
	return userId, nil
	//return fmt.Sprintf("%s%s", email, "hello"), nil
}
