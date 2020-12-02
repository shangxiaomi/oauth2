package controller

import (
	"encoding/json"
	"fmt"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"oauth2/common"
	"oauth2/config"
	mylog "oauth2/log"
	"oauth2/pkg/session"
	"oauth2/service"
	"time"
)

var srv *server.Server
var manager *manage.Manager

func init() {
	// 加载配置文件
	srv = common.GetServer()
	manager = common.GetManager()
	// 用来获取根据用户名和密码获取用户id
	srv.SetPasswordAuthorizationHandler(passwordAuthorizationHandler)
	srv.SetUserAuthorizationHandler(userAuthorizeHandler)
	srv.SetAuthorizeScopeHandler(authorizeScopeHandler)
	srv.SetInternalErrorHandler(internalErrorHandler)
	srv.SetResponseErrorHandler(responseErrorHandler)
}

// 授权处理,获取授权code
func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	if CORS(w,r) {
		return
	}
	var form url.Values
	if v, _ := session.Get(r, "RequestForm"); v != nil {
		r.ParseForm()
		if r.Form.Get("client_id") == "" {
			form = v.(url.Values)
		}
	}
	r.Form = form

	// 一次新的登录，需要将session中旧的"RequestForm"删除掉
	if err := session.Delete(w, r, "RequestForm"); err != nil {
		mylog.Error.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := srv.HandleAuthorizeRequest(w, r); err != nil {
		mylog.Error.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

type TplData struct {
	Client config.Client
	// 用户申请的合规scope
	Scope []config.Scope
	Error string
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if CORS(w,r) {
		return
	}
	if r.Method != "POST" {
		mylog.Error.Println("不支持的方法" + r.Method)
		http.Error(w, "不支持的方法", 405)
		return
	}
	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			mylog.Error.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	username := r.Form.Get("username")
	userService := service.NewUserService()
	_, msg, code, err := userService.CreateUser(email, password, username)
	if err != nil {
		mylog.Error.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

// 登录
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if CORS(w,r) {
		return
	}

	form, err := session.Get(r, "RequestForm")
	if err != nil {
		mylog.Error.Println("获取session失败" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 因为会从 userAuthorizeHandler 等中跳转过来，过意，一定会存在对应的参数
	if form == nil {
		mylog.Warn.Println("登录失败，session会话为空")
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	clientId := form.(url.Values).Get("client_id")
	scope := form.(url.Values).Get("scope")

	// 页面数据
	data := TplData{
		Client: config.GetClient(clientId),
		Scope:  config.ScopeFilter(clientId, scope),
	}

	if data.Scope == nil {
		mylog.Warn.Println("登陆失败" + "Invalid Scope")
		http.Error(w, "Invalid Scope", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				mylog.Error.Println("登陆失败" + err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		var userId string
		// 账号密码登录
		if r.Form.Get("type") == "password" {
			// 自己实现验证逻辑
			userService := service.NewUserService()
			userId, err = userService.GetUserIdByPwd(r.Form.Get("email"), r.Form.Get("password"))
			if err != nil {
				panic(fmt.Sprintf("数据查询时出错:%v", err.Error()))
			}

			if userId == "" {

				t, err := template.ParseFiles("tpl/index.html")
				if err != nil {
					mylog.Warn.Println("html模板渲染错误")
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				data.Error = "用户名密码错误!"
				t.Execute(w, data)
				mylog.Warn.Println("登陆失败用户名或密码错误" + r.Form.Get("email"))
				return
			}
		}
		if err := session.Set(w, r, "LoggedInUserID", userId); err != nil {
			mylog.Error.Println("session设置失败:" + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", "/authorize")
		w.WriteHeader(http.StatusFound)
		mylog.Info.Println("登陆成功" + r.Form.Get("email"))
		return
	}

	t, err := template.ParseFiles("tpl/index.html")
	if err != nil {
		mylog.Error.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

// 进行登出
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	redirectURI := r.Form.Get("redirect_uri")
	if _, err := url.Parse(redirectURI); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if err := session.Delete(w, r, "LoggedInUserID"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", redirectURI)
	w.WriteHeader(http.StatusFound)
}

// 进行token发放
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	err := srv.HandleTokenRequest(w, r)
	if err != nil {
		mylog.Error.Println("TokenHandler error:" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	token, err := srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cli, err := manager.GetClient(token.GetClientID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
		"user_id":    token.GetUserID(),
		"client_id":  token.GetClientID(),
		"scope":      token.GetScope(),
		"domain":     cli.GetDomain(),
	}
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(data)
}

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

func CORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if r.Method == http.MethodOptions {
		w.WriteHeader(200)
		return true
	}
	return false
}
