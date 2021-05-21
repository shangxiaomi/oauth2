package router

import (
	"oauth2/controller"
	"oauth2/middleware"

	"github.com/gin-gonic/gin"
)

func CollectRoute(g *gin.Engine) *gin.Engine {
	authController := controller.NewAuthController()
	sso := g.Use(middleware.CORSMiddleware())
	{ //sso := g.Group("/sso")
		sso.GET("/authorize", authController.AuthorizeHandler)
		sso.Any("/login", authController.LoginHandler)
		sso.POST("/token", authController.TokenHandler)
		sso.GET("/test", authController.TestHandler)
		sso.GET("/logout", authController.LogoutHandler)
		sso.POST("/register", authController.RegisterHandler)
		//sso.GET("/register", authController.GetRegisterHandler)
		// TODO 这里用的是相对于main.go文件所在的相对位置，需要考察一下,为什么不是本文件的相对位置
		sso.Static("/static", "./static")
	}
	return g
}
