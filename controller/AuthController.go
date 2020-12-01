package controller

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"oauth2/config"
	"oauth2/model"
	"oauth2/pkg/session"
	"time"
)

var srv *server.Server
var manager *manage.Manager

func init() {	// 加载配置文件
	config.Setup()
	// 设置session的相关配置
	session.Setup()
	// config oauth2 server
	manager = manage.NewDefaultManager()
	// 授权码配置
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	// 设置token存储
	manager.MustTokenStorage(store.NewMemoryTokenStore())
	// 设置生成token的方法
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate([]byte("shangxiaomisercert"), jwt.SigningMethodHS512))
	// 生成客户端存储
	clientStore := store.NewClientStore()
	for _, v := range config.Get().OAuth2.Client {
		clientStore.Set(v.ID, &models.Client{
			ID:     v.ID,
			Secret: v.Secret,
			Domain: v.Domain,
		})
	}
	manager.MapClientStorage(clientStore)
	srv = server.NewServer(server.NewConfig(), manager)

	// 用来获取根据用户名和密码获取用户id
	srv.SetPasswordAuthorizationHandler(passwordAuthorizationHandler)
	srv.SetUserAuthorizationHandler(userAuthorizeHandler)
	srv.SetAuthorizeScopeHandler(authorizeScopeHandler)
	srv.SetInternalErrorHandler(internalErrorHandler)
	srv.SetResponseErrorHandler(responseErrorHandler)
	// 可以添加额外的载荷
	//srv.SetExtensionFieldsHandler()
}

// 授权处理,获取授权code
func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := srv.HandleAuthorizeRequest(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

type TplData struct {
	Client config.Client
	// 用户申请的合规scope
	Scope []config.Scope
	Error string
}

// 登录
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	form, err := session.Get(r, "RequestForm")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 因为会从 userAuthorizeHandler 等中跳转过来，过意，一定会存在对应的参数
	if form == nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	clientId := form.(url.Values).Get("client_key")
	scope := form.(url.Values).Get("scope")

	// 页面数据
	data := TplData{
		Client: config.GetClient(clientId),
		Scope:  config.ScopeFilter(clientId, scope),
	}

	if data.Scope == nil {
		http.Error(w, "Invalid Scope", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		var userId string
		// 账号密码登录
		if r.Form.Get("type") == "password" {
			// 自己实现验证逻辑
			var user model.User
			userId = user.GetUserIDByPwd(r.Form.Get("username"), r.Form.Get("password"))
			if userId == "" {
				t, err := template.ParseFiles("tpl/login.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				data.Error = "用户名密码错误!"
				t.Execute(w, data)

				return
			}
		}
	}

	t, err := template.ParseFiles("tpl/login.html")
	if err != nil {
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
	v, _ := session.Get(r, "LoggedInUserId")
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
func passwordAuthorizationHandler(username string, password string) (string, error) {
	var user model.User
	userId := user.GetUserIDByPwd(username, password)
	return userId, nil
}
