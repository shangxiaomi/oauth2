package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"html/template"
	"net/http"
	"net/url"
	"oauth2/common"
	"oauth2/config"
	mylog "oauth2/log"
	"oauth2/pkg/session"
	"oauth2/service"
	"time"
)

type IAuthController interface {
	AuthorizeHandler(ctx *gin.Context)
	RegisterHandler(ctx *gin.Context)
	GetRegisterHandler(ctx *gin.Context)
	LoginHandler(ctx *gin.Context)
	LogoutHandler(ctx *gin.Context)
	TokenHandler(ctx *gin.Context)
	TestHandler(ctx *gin.Context)
}

type AuthController struct {
	srv     *server.Server
	manager *manage.Manager
}

type TplData struct {
	Client config.Client
	// 用户申请的合规scope
	Scope []config.Scope
	Error string
}

func NewAuthController() IAuthController {
	controller := AuthController{
		srv:     common.GetServer(),
		manager: common.GetManager(),
	}
	controller.srv.SetPasswordAuthorizationHandler(passwordAuthorizationHandler)
	controller.srv.SetUserAuthorizationHandler(userAuthorizeHandler)
	controller.srv.SetAuthorizeScopeHandler(authorizeScopeHandler)
	controller.srv.SetInternalErrorHandler(internalErrorHandler)
	controller.srv.SetResponseErrorHandler(responseErrorHandler)
	return controller
}

func (this AuthController) GetRegisterHandler(ctx *gin.Context) {
	panic("未实现的方法")
}

func (this AuthController) AuthorizeHandler(ctx *gin.Context) {
	var form url.Values
	if v, _ := session.Get(ctx.Request, "RequestForm"); v != nil {
		err := ctx.Request.ParseForm()
		if err != nil {
			mylog.Warn.Println("参数解析失败", ctx.Request)
			http.Error(ctx.Writer, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if ctx.Request.Form.Get("client_id") == "" {
			form = v.(url.Values)
		}
	}
	ctx.Request.Form = form

	// 一次新的登录，需要将session中旧的"RequestForm"删除掉
	if err := session.Delete(ctx.Writer, ctx.Request, "RequestForm"); err != nil {
		mylog.Error.Println(err)
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := this.srv.HandleAuthorizeRequest(ctx.Writer, ctx.Request); err != nil {
		mylog.Error.Println(err)
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (this AuthController) LoginHandler(ctx *gin.Context) {

	form, err := session.Get(ctx.Request, "RequestForm")
	if err != nil {
		mylog.Error.Println("获取session失败" + err.Error())
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	// 因为会从 userAuthorizeHandler 等中跳转过来，过意，一定会存在对应的参数
	if form == nil {
		mylog.Warn.Println("登录失败，session会话为空")
		http.Error(ctx.Writer, "Invalid Request", http.StatusBadRequest)
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
		http.Error(ctx.Writer, "Invalid Scope", http.StatusBadRequest)
		return
	}

	if ctx.Request.Method == "POST" {
		if ctx.Request.Form == nil {
			if err := ctx.Request.ParseForm(); err != nil {
				mylog.Error.Println("登陆失败" + err.Error())
				http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		var userId string
		// 账号密码登录
		if ctx.Request.Form.Get("type") == "password" {
			// 自己实现验证逻辑
			userService := service.NewUserService()
			userId, err = userService.GetUserIdByPwd(ctx.Request.Form.Get("email"), ctx.Request.Form.Get("password"))
			if err != nil {
				panic(fmt.Sprintf("数据查询时出错:%v", err.Error()))
			}

			if userId == "" {

				t, err := template.ParseFiles("tpl/index.html")
				if err != nil {
					mylog.Warn.Println("html模板渲染错误")
					http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
					return
				}
				data.Error = "用户名密码错误!"
				t.Execute(ctx.Writer, data)
				mylog.Warn.Println("登陆失败用户名或密码错误" + ctx.Request.Form.Get("email"))
				return
			}
		}
		if err := session.Set(ctx.Writer, ctx.Request, "LoggedInUserID", userId); err != nil {
			mylog.Error.Println("session设置失败:" + err.Error())
			http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		ctx.Writer.Header().Set("Location", "/authorize")
		ctx.Writer.WriteHeader(http.StatusFound)
		mylog.Info.Println("登陆成功" + ctx.Request.Form.Get("email"))
		return
	}

	t, err := template.ParseFiles("tpl/index.html")
	if err != nil {
		mylog.Error.Println(err)
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: 改成Gin的模板方法
	t.Execute(ctx.Writer, data)
}

func (this AuthController) RegisterHandler(ctx *gin.Context) {

	// TODO 注释中的方法无法获取参数
	/*
		//	var requestMap = make(map[string]string)
		//	json.NewDecoder(ctx.Request.Body).Decode(&requestMap)
		//	name := requestMap["username"]
		//	tele := requestMap["telephone"]
		//	password := requestMap["password"]
	*/
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")
	username := ctx.PostForm("username")
	userService := service.NewUserService()
	_, msg, code, err := userService.CreateUser(email, password, username)
	if err != nil {
		mylog.Error.Println(err.Error())
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.Writer.WriteHeader(code)
	ctx.Writer.Write([]byte(msg))
	// TODO 注册逻辑需要更改，如果注册失败需要返回原来的页面
}

func (this AuthController) LogoutHandler(ctx *gin.Context) {
	if ctx.Request.Form == nil {
		if err := ctx.Request.ParseForm(); err != nil {
			http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	redirectURI := ctx.Request.Form.Get("redirect_uri")
	if _, err := url.Parse(redirectURI); err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusBadRequest)
	}
	if err := session.Delete(ctx.Writer, ctx.Request, "LoggedInUserID"); err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx.Writer.Header().Set("Location", redirectURI)
	ctx.Writer.WriteHeader(http.StatusFound)
}

func (this AuthController) TokenHandler(ctx *gin.Context) {
	err := this.srv.HandleTokenRequest(ctx.Writer, ctx.Request)
	if err != nil {
		mylog.Error.Println("TokenHandler error:" + err.Error())
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (this AuthController) TestHandler(ctx *gin.Context) {
	token, err := this.srv.ValidationBearerToken(ctx.Request)
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusBadRequest)
		return
	}
	cli, err := this.manager.GetClient(token.GetClientID())
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
		"user_id":    token.GetUserID(),
		"client_id":  token.GetClientID(),
		"scope":      token.GetScope(),
		"domain":     cli.GetDomain(),
	}
	e := json.NewEncoder(ctx.Writer)
	e.SetIndent("", "  ")
	e.Encode(data)
}
