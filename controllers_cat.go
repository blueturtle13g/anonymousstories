package main

import (
	"net/http"
	"strconv"

	"fmt"
	"github.com/julienschmidt/httprouter"
)

func SingleCat(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	parameter := ps.ByName("id")
	var cat Cat
	CatId, err := strconv.Atoi(parameter)
	if err != nil {
		cat = getCatByName(parameter)
	} else {
		cat = getCatById(CatId)
	}
	gotFlash := make(map[string]interface{})
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	}
	gotFlash["Cat"] = cat
	stories := getStoriesByCatId(cat.Id)
	writers := getWritersByCatId(cat.Id)
	gotFlash["Stories"] = sortStoriesByDate(stories)
	gotFlash["Writers"] = writers
	gotFlash["Creator"] = getWriterByName(cat.By)
	gotFlash["Title"] = cat.Name
	gotFlash["WriterName"] = cWriter.Name

	var topTenStories []Story
	soredStories := sortStoriesByRate(stories)
	if len(soredStories) > 10{
		for i:= 9; i >= 0; i--{
			topTenStories = append(topTenStories, soredStories[i])
		}
	}else{
		topTenStories = soredStories
	}
	gotFlash["TopTenStories"] = topTenStories

	var topTenWriters []Writer
	sortedWriters := sortWritersByRate(writers)
	if len(sortedWriters) > 10{
		for i:= 9; i >= 0; i--{
			topTenWriters = append(topTenWriters, sortedWriters[i])
		}
	}else{
		topTenWriters = sortedWriters
	}

	gotFlash["TopTenWriters"] = topTenWriters

	if err := tpl.ExecuteTemplate(w, "cat.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func Cats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	gotFlash := make(map[string]interface{})
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	}
	if err := session.PopObject(w, "sentFlash", gotFlash); err != nil {
		fmt.Println(err)
	}
	gotFlash["Title"] = "Categories"
	gotFlash["WriterName"] = cWriter.Name
	order := gotFlash["Order"]
	priority := gotFlash["Priority"]
	// if any result for search was found we get it, if not we show all categories.
	if gotFlash["LenC"] == nil {
		gotFlash["Cats"] = getAllCats("", "")
	}
	if order != nil {
		gotFlash["Cats"] = getAllCats(order.(string), priority.(string))
	}
	var topTen []Cat
	if len(getAllCats("rate", "")) > 10{
		for i:= 9; i >= 0; i--{
			topTen = append(topTen, getAllCats("rate", "")[i])
		}
	}else{
		topTen = getAllCats("rate", "")
	}

	gotFlash["TopTenCats"] = topTen
	if err := tpl.ExecuteTemplate(w, "cats.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func CatsProcess(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	sentFlash := make(map[string]interface{})
	// if request is search.
	if word := r.FormValue("search"); len(word) > 0 {
		cats := searchCats(word)
		if len(cats) == 0 {
			sentFlash["Err"] = "No Category Found For This Word."
			if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
				fmt.Println(err)
			}
			return
		}
		// if something was found.
		sentFlash["Cats"] = cats
		sentFlash["LenC"] = len(cats)
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/cats", 302)
		return
	}
	// if request was ordering and not searching.
	sentFlash["Order"] = r.FormValue("order")
	sentFlash["Priority"] = r.FormValue("priority")
	if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, "/cats", 302)
}
