package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3/manage"
	_ "oauth2/common"
	"oauth2/config"
	mylog "oauth2/log"
	"oauth2/router"
)

var manager *manage.Manager

func main() {
	gin.DefaultWriter = mylog.GetLogFile()
	r := gin.Default()
	r = router.CollectRoute(r)
	panic(r.Run(fmt.Sprintf(":%d", config.Get().Server.Port))) // listen and serve on 0.0.0.0:8080
}

