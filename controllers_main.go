package main

import (
	"net/http"
	"strings"

	"fmt"
	"github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	}
	gotFlash := make(map[string]interface{})
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}
	var err error
	if gotFlash["Cong"], err = session.PopString(w, "Cong"); err != nil {
		fmt.Println(err)
	}
	gotFlash["Title"] = "Anonymous Stories"
	gotFlash["WriterName"] = cWriter.Name
	if gotFlash["Stories"] == nil {
		gotFlash["Stories"] = getAllStories("", "")
	}

	if err := tpl.ExecuteTemplate(w, "index.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func IndexProcess(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	sentFlash := make(map[string]interface{})
	// if the request was searching
	if word := r.FormValue("search"); len(word) > 0 {
		tags := searchTags(word)
		stories := searchStories(word)
		if len(tags) == 0 && len(stories) == 0 {
			sentFlash["Err"] = "No Result Found For This Word."
			sentFlash["Stories"] = getAllStories("", "")
		} else {
			// if the request was successful
			sentFlash["Tags"] = tags
			sentFlash["Stories"] = stories
			sentFlash["LenS"] = len(stories)
			sentFlash["LenT"] = len(tags)
		}
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/", 302)
		return
	}
	// if the request was ordering
	order := r.FormValue("order")
	priority := r.FormValue("priority")
	sentFlash["Stories"] = getAllStories(order, priority)
	if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, "/", 302)
}

func About(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	gotFlash := make(map[string]interface{})
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	}
	gotFlash["Title"] = "About AStories"
	gotFlash["WriterName"] = cWriter.Name

	if err := tpl.ExecuteTemplate(w, "about.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func Contact(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	}
	gotFlash := make(map[string]interface{})
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}
	gotFlash["Title"] = "Contact Us"
	gotFlash["WriterName"] = cWriter.Name

	if err := tpl.ExecuteTemplate(w, "contact.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func ContactProcess(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	sentFlash := make(map[string]interface{})
	var name, email string
	// to see if the user is already logged in, we don't ask for their information.
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter := getWriterById(WriterId)
		name = cWriter.Name
		email = cWriter.Email
	} else {
		name = r.FormValue("name")
		email = r.FormValue("email")
	}
	text := r.FormValue("text")
	newMsg := Message{Name: strings.TrimSpace(name), Email: strings.TrimSpace(email), Text: strings.TrimSpace(text)}
	if warnings := newMsg.Validate(); len(warnings) > 0 {
		sentFlash["Errs"] = warnings
		sentFlash["Name"] = name
		sentFlash["Email"] = email
		sentFlash["Text"] = text
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/contact", 302)
		return
	}
	if msgId := insertMsg(newMsg); msgId > 0 {
		sentFlash["Cong"] = "Congratulations dear " + name + ", your message has been sent successfully."
	} else {
		sentFlash["Err"] = "there is something wrong with your message, please check it and try again."
	}
	if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, "/contact", 302)
}

func Charts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	}

	gotFlash := make(map[string]interface{})
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}

	gotFlash["Title"] = "Charts"
	gotFlash["WriterName"] = cWriter.Name

	var top250Stories []Story
	if len(getAllWriters("rate", "")) > 250{
		for i:= 9; i >= 0; i--{
			top250Stories = append(top250Stories, getAllStories("rate", "")[i])
		}
	}else{
		top250Stories = getAllStories("rate", "")
	}
	gotFlash["Top250Stories"] = top250Stories

	var top250Writers []Writer
	if len(getAllWriters("rate", "")) > 250{
		for i:= 9; i >= 0; i--{
			top250Writers = append(top250Writers, getAllWriters("rate", "")[i])
		}
	}else{
		top250Writers = getAllWriters("rate", "")
	}
	gotFlash["Top250Writers"] = top250Writers

	var top50Cats []Cat
	if len(getAllCats("rate", "")) > 50{
		for i:= 9; i >= 0; i--{
			top50Cats = append(top50Cats, getAllCats("rate", "")[i])
		}
	}else{
		top50Cats = getAllCats("rate", "")
	}
	gotFlash["Top50Cats"] = top50Cats

	if err := tpl.ExecuteTemplate(w, "charts.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}