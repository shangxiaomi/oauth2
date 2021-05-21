package common

import (
	"oauth2/config"
	mylog "oauth2/log"
	"oauth2/pkg/session"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

var srv *server.Server
var manager *manage.Manager

func init() {
	// 一定要在db前进行调用
	// 加载配置文件
	config.Setup()
	// 进行数据库初始化
	InitDB()
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
			// 设置客户端的id
			ID:     v.ID,
			Secret: v.Secret,
			Domain: v.Domain,
		})
	}
	manager.MapClientStorage(clientStore)
	srv = server.NewServer(server.NewConfig(), manager)
	mylog.Info.Println("oauth2.0 service init success")
	// 可以添加额外的载荷
}

func GetServer() *server.Server {
	return srv
}

func GetManager() *manage.Manager {
	return manager
}
