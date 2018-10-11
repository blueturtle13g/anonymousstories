package main

import (
	"encoding/gob"
	"github.com/alexedwards/scs"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"time"
)

var (
	tpl            *template.Template
	sessionManager = scs.NewCookieManager("cPfu>HIUkVBA1M7W/gNo+ZEjtp0}Yz-~Gv4lmdXyTOQ{$9r3Rs2^#nwqC8i6JK5D")
	DB             = dbConn()
)

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	sessionManager.Lifetime(time.Hour * 6) // Set the maximum session lifetime to 1 hour.
	sessionManager.Persist(false)          // Persist the session after a user has closed their browser.
}

func Routes() {
	router := httprouter.New()
	router.ServeFiles("/static/*filepath", http.Dir("static"))
	gob.Register([]Story{})
	gob.Register([]Cat{})
	gob.Register([]Tag{})
	router.GET("/out", LogOut)
	router.GET("/about", About)
	router.GET("/charts", Charts)
	router.GET("/tag/:id", SingleTag)
	router.GET("/cat/:id", SingleCat)
	router.GET("/writer/:id", SingleWriter)
	router.GET("/profile/:id", Profile)

	router.GET("/profile/:id/edit", EditProfile)
	router.POST("/profile/:id/edit", EditProfileProcess)

	router.GET("/story/:id/edit", EditStory)
	router.POST("/story/:id/edit", EditStoryProcess)

	router.GET("/profile/:id/editPassword", EditPassword)
	router.POST("/profile/:id/editPassword", EditPasswordProcess)

	router.GET("/", Index)
	router.POST("/", IndexProcess)

	router.GET("/in", LogIn)
	router.POST("/in", LogInProcess)

	router.GET("/cats", Cats)
	router.POST("/cats", CatsProcess)

	router.GET("/contact", Contact)
	router.POST("/contact", ContactProcess)

	router.GET("/writers", Writers)
	router.POST("/writers", WritersProcess)

	router.GET("/tellStory", TellStory)
	router.POST("/tellStory", TellStoryProcess)

	router.GET("/register", Register)
	router.POST("/register", RegisterProcess)

	router.GET("/story/:id", SingleStory)
	router.POST("/story/:id", SingleStoryProcess)

	http.ListenAndServe(":8080", sessionManager.Use(router))
}
