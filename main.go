package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3/manage"
	mylog "oauth2/log"
	"oauth2/router"
)

var manager *manage.Manager

func main() {
	gin.DefaultWriter = mylog.GetLogFile()
	r := gin.Default()
	r = router.CollectRoute(r)
	panic(r.Run(":9096")) // listen and serve on 0.0.0.0:8080
}

