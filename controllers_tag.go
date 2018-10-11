package main

import (
	"net/http"
	"strconv"

	"fmt"
	"github.com/julienschmidt/httprouter"
)

func SingleTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else {
		cWriter = getWriterById(WriterId)
	}
	parameter := ps.ByName("id")
	TagId, _ := strconv.Atoi(parameter)
	gotFlash := make(map[string]interface{})
	tag := getTagById(TagId)
	gotFlash["Title"] = "#" + tag.Name
	gotFlash["WriterName"] = cWriter.Name
	gotFlash["Stories"] = getStoriesByTagId(TagId)
	gotFlash["Tag"] = tag

	if err := tpl.ExecuteTemplate(w, "tag.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}
