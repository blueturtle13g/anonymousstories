package main

import (
	"net/http"
	"strconv"

	"fmt"
	"github.com/julienschmidt/httprouter"
)

func SingleStory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	}
	parameter := ps.ByName("id")
	storyId, _ := strconv.Atoi(parameter)
	if affect := incrementViewCount(storyId); affect < 1 {
		fmt.Println("view count didn't increment.")
	}
	gotFlash := make(map[string]interface{})
	wholeStory, ItsComments, ItsTags := getStoryById(storyId)
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}
	gotFlash["WriterName"] = cWriter.Name
	gotFlash["StoryWriter"] = getWriterByName(wholeStory.By)
	gotFlash["Title"] = wholeStory.Title
	gotFlash["Story"] = wholeStory
	gotFlash["Coms"] = ItsComments
	gotFlash["Cat"] = getCatByStoryId(storyId)
	gotFlash["Tags"] = getUniqueTag(ItsTags)
	gotFlash["Rate"] = getStoryRates(storyId)
	if err := tpl.ExecuteTemplate(w, "story.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func SingleStoryProcess(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	parameter := ps.ByName("id")
	storyId, _ := strconv.Atoi(parameter)
	sentFlash := make(map[string]interface{})
	// if user is not registered and is trying to comment,
	// we ask them to first log in
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	} else {
		w.Write([]byte("Please First Register"))
		return
	}
	cStory, _, _ := getStoryById(storyId)
	submit := r.FormValue("submit")
	text := r.FormValue("text")
	storyCat := getCatByStoryId(storyId)

	if submit == "rate" {
		if cWriter.Name == cStory.By {
			w.Write([]byte("Sorry, You Can't Rate To Your Own Story"))
			return
		}
		RatePrem, err := strconv.Atoi(r.FormValue("RatePrem"))
		if err != nil {
			w.Write([]byte("Please rate to all fields"))
			return
		}
		RatePres, err := strconv.Atoi(r.FormValue("RatePres"))
		if err != nil {
			w.Write([]byte("Please rate to all fields"))
			return
		}
		RateStr, err := strconv.Atoi(r.FormValue("RateStr"))
		if err != nil {
			w.Write([]byte("Please rate to all fields"))
			return
		}
		RateChar, err := strconv.Atoi(r.FormValue("RateChar"))
		if err != nil {
			w.Write([]byte("Please rate to all fields"))
			return
		}
		RateTheme, err := strconv.Atoi(r.FormValue("RateTheme"))
		if err != nil {
			w.Write([]byte("Please rate to all fields"))
			return
		}
		RateStyle, err := strconv.Atoi(r.FormValue("RateStyle"))
		if err != nil {
			w.Write([]byte("Please rate to all fields"))
			return
		}
		RateCom, err := strconv.Atoi(r.FormValue("RateCom"))
		if err != nil {
			w.Write([]byte("Please rate to all fields"))
			return
		}

		rates := []float64{float64(RatePrem), float64(RatePres),
			float64(RateStr), float64(RateChar), float64(RateTheme),
			float64(RateStyle), float64(RateCom)}

		RateOverAll := getAverage(rates)
		newRate := Rate{StoryId: cStory.Id, RaterId: cWriter.Id, Overall: RateOverAll, Premise: RatePrem, Presentation: RatePres, Structure: RateStr, Characters: RateChar, Theme: RateTheme, Style: RateStyle, Commercial: RateCom}
		// if there is a record in rates, with this raterId it means, the guy has already rated, so we update it.
		if affect := alreadyRate(cWriter.Id, cStory.Id); affect > 0 {
			if affect := updateRates(cWriter.Id, storyId, newRate); affect < 1 {
				fmt.Println("updating rates, failed.")
				sentFlash["Err"] = "updating rates, failed."
			} else {
				// if updating a rate was successfully done, we should update its dependent rates too.
				updateDependentRates(cStory, storyCat, "up")
				sentFlash["Cong"] = "Your Previous Rate Was Updated."
			}
			if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
				fmt.Println(err)
			}
			w.Write([]byte("done"))
			return
		}

		if rateId := insertRates(cWriter.Id, storyId, newRate); rateId < 1 {
			fmt.Println("inserting rates, failed.")
			sentFlash["Err"] = "inserting rates, failed."
		} else {
			sentFlash["Cong"] = "Your Rate Was Successfully Counted."
		}
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		// after each rate we update all dependent rates.
		updateDependentRates(cStory, storyCat, "")
		w.Write([]byte("done"))
		return
	}
	// if the request was commenting and not rating.
	if submit == "com" {
		if len(text) < 3 || len(text) > 500 {
			w.Write([]byte("The length of your comment can't be less than 3 or more than 500 characters."))
			return
		}
		newComment := Com{By: cWriter.Name, StoryId: storyId, Text: text}
		if affect := insertCom(newComment); affect < 0 {
			fmt.Println("inserting comment failed.")
		}
		w.Write([]byte("done"))
		return
	}
	if submit == "resCom" {
		comId, err := strconv.Atoi(r.FormValue("comId"))
		if err != nil {
			fmt.Println(err)
		}
		if len(text) < 3 || len(text) > 500 {
			w.Write([]byte("The length of your comment can't be less than 3 or more than 500 characters"))
			return
		}
		newComReply := Com{By: cWriter.Name, ComId: comId, StoryId: cStory.Id, Text: text}
		if resComId := insertCom(newComReply); resComId < 1 {
			fmt.Println("inserting resCom failed.")
		}
		w.Write([]byte("done"))
		return
	}
	if submit == "EditCom" {
		comId, err := strconv.Atoi(r.FormValue("comId"))
		if err != nil {
			fmt.Println(err)
		}
		if len(text) < 3 || len(text) > 500 {
			w.Write([]byte("The length of your comment can't be less than 3 or more than 500 characters"))
			return
		}
		// if we encounter an internal problem
		if affect := updateCom(comId, text); affect < 1 {
			fmt.Println("Updating Failed, Please Try Again.")
			return
		}
		w.Write([]byte("done"))
		return
	}
	if submit == "DeleteCom" {
		comId, err := strconv.Atoi(r.FormValue("comId"))
		if err != nil {
			fmt.Println(err)
		}
		if affect := deleteComment(comId); affect < 1 {
			w.Write([]byte("The Deletion Of Your Comment Failed, Please Try Again"))
			return
		}
		w.Write([]byte("done"))
		return
	}
}

func EditStory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	parameter := ps.ByName("id")
	storyId, err := strconv.Atoi(parameter)
	if err != nil {
		fmt.Println("Please pass an id")
	}
	psStory, _, _ := getStoryById(storyId)

	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	} else if WriterId < 1 || cWriter.Name != psStory.By {
		http.Redirect(w, r, "/", 303)
		return
	}
	gotFlash := make(map[string]interface{})
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}

	gotFlash["Cat"], gotFlash["Story"], gotFlash["Cats"], gotFlash["Title"], gotFlash["WriterName"], gotFlash["UploadPic"] = getCatByStoryId(psStory.Id), psStory, getAllCats("rate", "desc"), psStory.Title, cWriter.Name, true

	if err := tpl.ExecuteTemplate(w, "editStory.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func EditStoryProcess(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	parameter := ps.ByName("id")
	storyId, err := strconv.Atoi(parameter)
	if err != nil {
		fmt.Println("Please pass an id")
	}
	psStory, _, _ := getStoryById(storyId)

	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	} else if WriterId < 1 || cWriter.Name != psStory.By {
		http.Redirect(w, r, "/", 303)
		return
	}
	sentFlash := make(map[string]interface{})
	if r.FormValue("submit") == "Delete" {
		if affect := deleteStory(psStory.Id); affect < 0 {
			fmt.Println("deletion of story didn't happen")
		}
		if err := session.PutString(w, "Cong", "Your Story Is Deleted."); err != nil {
			fmt.Println(err)
		}
		w.Write([]byte("done"))
		return
	}
	var maxFileSize int64 = 4 * 1000 * 1000 //limit upload file to 10m
	if r.ContentLength > maxFileSize {
		sentFlash["Err"] = "The Maximum Size Of Story Picture Is 4M."
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/story/"+strconv.Itoa(psStory.Id)+"/edit", 303)
		return
	}

	file, multipartFileHeader, err := r.FormFile("pic")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	var picName string
	if multipartFileHeader.Filename != "" {
		if valid := detectFileType(file); !valid {
			sentFlash["Err"] = "Please Upload An Image, Other Types Are Not Supported."
			if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/story/"+strconv.Itoa(psStory.Id)+"/edit", 303)
			return
		}
		picName = processStoryPic(file, psStory)
	}
	title := r.FormValue("title")
	body := r.FormValue("body")
	tags := tagFinder(body)
	selectedCat := r.FormValue("selectCat")
	madeCat := r.FormValue("cat")

	var warnings []string
	var cat string

	upStory := Story{Id: storyId, Title: title, Body: body, By: cWriter.Name, Pic: picName}
	sentFlash["TitleStory"] = title
	sentFlash["Body"] = body
	sentFlash["SelectedCat"] = selectedCat
	sentFlash["Cat"] = madeCat
	sentFlash["Cats"] = getAllCats("", "")
	if warnings, cat = upStory.Validate(selectedCat, madeCat, "up", tags); len(warnings) > 0 {
		sentFlash["Errs"] = warnings
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/story/"+strconv.Itoa(psStory.Id)+"/edit", 303)
		return
	}

	if affect := updateStory(upStory); affect < 1 {
		sentFlash["Err"] = "There is something wrong with your data, Please check it and try again."
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/story/"+strconv.Itoa(psStory.Id)+"/edit", 303)
		return
	}
	// if the process finished successfully.
	processCat(cat, cWriter.Name, cWriter.Id, storyId)
	processTag(w, cWriter.Id, storyId, tags)
	http.Redirect(w, r, "/story/"+strconv.Itoa(psStory.Id), 302)
}

func TellStory(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	} else {
		http.Redirect(w, r, "/", 303)
		return
	}
	gotFlash := make(map[string]interface{})
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}
	gotFlash["Cats"], gotFlash["Title"], gotFlash["WriterName"], gotFlash["UploadPic"] = getAllCats("rate", "desc"), "Tell Your Story", cWriter.Name, true

	if err := tpl.ExecuteTemplate(w, "tellStory.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func TellStoryProcess(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	} else {
		http.Redirect(w, r, "/", 302)
		return
	}
	sentFlash := make(map[string]interface{})
	title := r.FormValue("title")
	body := r.FormValue("body")
	tags := tagFinder(body)
	selectedCat := r.FormValue("selectCat")
	madeCat := r.FormValue("cat")
	var maxFileSize int64 = 4 * 1000 * 1000 //limit upload file to 10m
	if r.ContentLength > maxFileSize {
		sentFlash["Err"] = "The Maximum Size Of Story Picture Is 4M."
		sentFlash["TitleStory"] = title
		sentFlash["Body"] = body
		sentFlash["SelectedCat"] = selectedCat
		sentFlash["Cat"] = madeCat
		sentFlash["Cats"] = getAllCats("", "")
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/tellStory", 303)
		return
	}

	file, multipartFileHeader, err := r.FormFile("pic")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	if multipartFileHeader.Filename != "" {
		if valid := detectFileType(file); !valid {
			sentFlash["Err"] = "Please Upload An Image, Other Types Are Not Supported."
			sentFlash["TitleStory"] = title
			sentFlash["Body"] = body
			sentFlash["SelectedCat"] = selectedCat
			sentFlash["Cat"] = madeCat
			sentFlash["Cats"] = getAllCats("", "")
			if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/tellStory", 303)
			return
		}
	}

	var warnings []string
	var cat string

	newStory := Story{Title: title, Body: body, By: cWriter.Name}
	sentFlash["TitleStory"] = title
	sentFlash["Body"] = body
	sentFlash["SelectedCat"] = selectedCat
	sentFlash["Cat"] = madeCat
	sentFlash["Cats"] = getAllCats("", "")
	if warnings, cat = newStory.Validate(selectedCat, madeCat, "", tags); len(warnings) > 0 {
		sentFlash["Errs"] = warnings
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/tellStory", 302)
		return
	}

	var storyId int
	if storyId = insertStory(newStory); storyId < 1 {
		sentFlash["Err"] = "There is something wrong with your data, Please check it and try again."
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/tellStory", 302)
		return
	}
	if multipartFileHeader.Filename != "" {
		story, _, _ := getStoryById(storyId)
		picName := processStoryPic(file, story)
		story.Pic = picName
		if affect := updateStory(story); affect < 1 {
			fmt.Println("story didn't get updated to get its pic")
		}
	}
	processCat(cat, cWriter.Name, cWriter.Id, storyId)
	processTag(w, cWriter.Id, storyId, tags)
	http.Redirect(w, r, "/story/"+strconv.Itoa(storyId), 302)
}
