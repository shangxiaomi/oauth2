package main

import (
	"gopkg.in/oauth2.v3/manage"
	"log"
	"net/http"
	"oauth2/controller"
)

var manager *manage.Manager

func main() {

	// 添加路由
	http.HandleFunc("/authorize", controller.AuthorizeHandler)
	http.HandleFunc("/register",controller.RegisterHandler)
	http.HandleFunc("/login", controller.LoginHandler)
	http.HandleFunc("/logout", controller.LogoutHandler)
	http.HandleFunc("/token", controller.TokenHandler)
	http.HandleFunc("/test", controller.TestHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	log.Println("Server is running at 9096 port.")
	log.Fatal(http.ListenAndServe(":9096", nil))
}

