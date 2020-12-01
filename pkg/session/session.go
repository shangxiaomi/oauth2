package session

import (
	"encoding/gob"
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
	"net/url"
	"oauth2/config"
)

var store *sessions.CookieStore

func Setup() {
	// 传输的时候，需要对接口进行注册
	gob.Register(url.Values{})
	store = sessions.NewCookieStore([]byte(config.Get().Session.SecretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 20,
		HttpOnly: true,
	}
}

func Get(r *http.Request, name string) (val interface{}, err error) {
	fmt.Println("GET" ,r)
	session, err := store.Get(r, config.Get().Session.Name)
	if err != nil {
		return
	}
	val = session.Values[name]
	return
}

func Set(w http.ResponseWriter, r *http.Request, name string, val interface{}) (err error) {
	fmt.Println("Set" ,r)
	// Get a session.
	session, err := store.Get(r, config.Get().Session.Name)
	if err != nil {
		return
	}
	session.Values[name] = val
	err = session.Save(r, w)
	return
}


func Delete(w http.ResponseWriter, r *http.Request, name string) (err error) {
	// Get a session.
	session, err := store.Get(r, config.Get().Session.Name)
	if err != nil {
		return
	}

	delete(session.Values, name)
	err = session.Save(r, w)

	return
}
