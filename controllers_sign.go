package main

import (
	"log"
	"net/http"
	"strings"

	"fmt"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

func LogIn(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	gotFlash := make(map[string]interface{})
	gotFlash["Title"] = "Sign Page"
	if logged, err := session.Exists("WriterId"); err != nil {
		fmt.Println(err)
	} else if logged {
		http.Redirect(w, r, "/", 302)
		return
	}
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}

	if err := tpl.ExecuteTemplate(w, "in.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func LogInProcess(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	sentFlash := make(map[string]interface{})
	if logged, err := session.Exists("WriterId"); err != nil {
		fmt.Println(err)
	} else if logged {
		http.Redirect(w, r, "/", 302)
		return
	}
	uORe := r.FormValue("uORe")
	sentFlash["UORE"] = uORe
	password := r.FormValue("password")
	var WriterId int
	// check to see if the email or username is correct
	if WriterId = checkWriterName(uORe); WriterId < 1 {
		sentFlash["Err"] = "Your User Name Or Email Is Wrong."
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/in", 302)
		return
	}
	// username or email is correct
	cWriter := getWriterById(WriterId)
	hashed := cWriter.Password
	// check to see if the password is correct
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		sentFlash["Err"] = "your Password is wrong."
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/in", 303)
		return
	}
	// password is correct
	if err := session.PutInt(w, "WriterId", cWriter.Id); err != nil {
		fmt.Println(err)
	}
	if storyId, _ := strconv.Atoi(r.FormValue("from")); storyId > 1 {
		http.Redirect(w, r, "/story/"+strconv.Itoa(storyId), 302)
		return
	}
	http.Redirect(w, r, "/profile/"+cWriter.Name, 302)
}

func LogOut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	if writerId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else {
		if affect := updateLastLog(writerId); affect < 1 {
			fmt.Println("update last log didn't work")
		}
	}
	session.Destroy(w)
	w.Write([]byte("done"))
}

func Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	gotFlash := make(map[string]interface{})
	if logged, err := session.Exists("WriterId"); err != nil {
		fmt.Println(err)
	} else if logged {
		http.Redirect(w, r, "/", 302)
		return
	}
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}

	gotFlash["Title"] = "Sign Page"

	if err := tpl.ExecuteTemplate(w, "register.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func RegisterProcess(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	sentFlash := make(map[string]interface{})

	submit := r.FormValue("submit")
	username := r.FormValue("username")
	if submit == "CheckName" {
		if WriterId, err := strconv.Atoi(r.FormValue("writerId")); err != nil {
			fmt.Println(err)
		} else {
			if er := nameValidation(username); er != "" {
				w.Write([]byte("Invalid"))
				return
			}
			if affect := alreadyName(username, WriterId); affect > 0 {
				w.Write([]byte("Taken"))
			} else {
				w.Write([]byte("Available"))
			}
			return
		}
	}

	if logged, err := session.Exists("WriterId"); err != nil {
		fmt.Println(err)
	} else if logged {
		http.Redirect(w, r, "/", 302)
		return
	}

	email := r.FormValue("email")
	pwd := r.FormValue("password")
	confirm := r.FormValue("confirm")
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	password := string(hash)
	newWriter := Writer{Name: strings.TrimSpace(username), Email: strings.TrimSpace(email), Password: strings.TrimSpace(password)}
	warnings := newWriter.Validate(pwd, confirm)
	// if validation of form failed
	if len(warnings) > 0 {
		sentFlash["Errs"] = warnings
		sentFlash["UserName"] = username
		sentFlash["Email"] = email
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/register", 302)
		return
	}
	// if validation was passed
	WriterId := insertWriter(newWriter)
	if err := session.PutInt(w, "WriterId", WriterId); err != nil {
		fmt.Println(err)
	}
	if err := session.PutString(w, "Cong", "Congratulations, You Successfully Registered In Anonymous Stories."); err != nil {
		fmt.Println(err)
	}
	if storyId, _ := strconv.Atoi(r.FormValue("from")); storyId > 1 {
		http.Redirect(w, r, "/story/"+strconv.Itoa(storyId), 302)
		return
	}
	http.Redirect(w, r, "/profile/"+strconv.Itoa(WriterId), 302)
}
