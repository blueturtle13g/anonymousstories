package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"golang.org/x/crypto/bcrypt"
	"log"
)

func SingleWriter(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	}
	parameter := ps.ByName("id")
	var writer Writer
	if writerId, err := strconv.Atoi(parameter); err != nil {
		writer = getWriterByName(parameter)
	} else {
		writer = getWriterById(writerId)
	}
	gotFlash := make(map[string]interface{})

	gotFlash["WriterName"] = cWriter.Name
	gotFlash["Writer"] = writer
	gotFlash["Title"] = writer.Name
	gotFlash["Cats"] = getWritersCats(writer.Id)
	gotFlash["Stories"] = sortStoriesByDate(getWritersStories(writer.Id))


	var topTen []Story
	if len(getWritersStories(writer.Id)) > 10{
		for i:= 9; i >= 0; i--{
			topTen = append(topTen, sortStoriesByRate(getWritersStories(writer.Id))[i])
		}
	}else{
		topTen = sortStoriesByRate(getWritersStories(writer.Id))
	}
	gotFlash["TopTenStories"] = topTen
	if err := tpl.ExecuteTemplate(w, "writer.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func Writers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	if gotFlash["Writers"] == nil {
		gotFlash["Writers"] = getAllWriters("", "")
	}
	gotFlash["Title"] = "Writers"
	gotFlash["WriterName"] = cWriter.Name
	var topTen []Writer
	if len(getAllWriters("rate", "")) > 10{
		for i:= 9; i >= 0; i--{
			topTen = append(topTen, getAllWriters("rate", "")[i])
		}
	}else{
		topTen = getAllWriters("rate", "")
	}
	gotFlash["TopTenWriters"] = topTen

	if err := tpl.ExecuteTemplate(w, "writers.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func WritersProcess(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session := sessionManager.Load(r)
	sentFlash := make(map[string]interface{})
	// if the request was searching
	if word := r.FormValue("search"); len(word) > 0 {
		writers := searchWriters(word)
		if len(writers) == 0 {
			sentFlash["Err"] = "No Writer Found For This Word."
			sentFlash["Writers"] = getAllWriters("", "")
		} else {
			sentFlash["Writers"] = writers
			sentFlash["LenW"] = len(writers)
		}
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/writers", 302)
		return
	}
	// if the request was ordering
	order := r.FormValue("order")
	priority := r.FormValue("priority")
	sentFlash["Writers"] = getAllWriters(order, priority)
	if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, "/writers", 302)
}

func Profile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	// the writer that is logged in
	var cWriter Writer
	// the writer that its details is passed through url
	var psWriter Writer
	parameter := ps.ByName("id")
	writerId, err := strconv.Atoi(parameter)
	if err != nil {
		psWriter = getWriterByName(parameter)
	} else {
		psWriter = getWriterById(writerId)
	}

	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	} else if WriterId < 1 || WriterId != psWriter.Id {
		http.Redirect(w, r, "/", 303)
		return
	}
	gotFlash := make(map[string]interface{})

	if cong, err := session.PopString(w, "Cong"); err != nil {
		fmt.Println(err)
	} else {
		gotFlash["Cong"] = cong
	}
	gotFlash["Stories"] = sortStoriesByDate(getWritersStories(cWriter.Id))
	gotFlash["Writer"], gotFlash["Title"], gotFlash["WriterName"] = cWriter, cWriter.Name, cWriter.Name
	gotFlash["Cats"] = getWritersCats(cWriter.Id)

	var topTen []Story
	if len(getWritersStories(cWriter.Id)) > 10{
		for i:= 9; i >= 0; i--{
			topTen = append(topTen, sortStoriesByDate(getWritersStories(cWriter.Id))[i])
		}
	}else{
		topTen = sortStoriesByDate(getWritersStories(cWriter.Id))
	}
	gotFlash["TopTenStories"] = topTen
	if err := tpl.ExecuteTemplate(w, "profile.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func EditProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	var psWriter Writer
	parameter := ps.ByName("id")
	writerId, err := strconv.Atoi(parameter)
	if err != nil {
		psWriter = getWriterByName(parameter)
	} else {
		psWriter = getWriterById(writerId)
	}
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	} else if WriterId < 1 || WriterId != psWriter.Id {
		http.Redirect(w, r, "/", 303)
		return
	}
	gotFlash := make(map[string]interface{})
	if err := session.PopObject(w, "sentFlash", &gotFlash); err != nil {
		fmt.Println(err)
	}

	gotFlash["Writer"], gotFlash["Title"], gotFlash["WriterName"], gotFlash["UploadPic"] = cWriter, cWriter.Name, cWriter.Name, true

	if err := tpl.ExecuteTemplate(w, "editProfile.gohtml", gotFlash); err != nil {
		fmt.Println(err)
	}
}

func EditProfileProcess(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	session := sessionManager.Load(r)
	var cWriter Writer
	var psWriter Writer
	parameter := ps.ByName("id")
	if WriterId, err := strconv.Atoi(parameter); err != nil {
		psWriter = getWriterByName(parameter)
	} else {
		psWriter = getWriterById(WriterId)
	}
	if WriterId, err := session.GetInt("WriterId"); err != nil {
		fmt.Println(err)
	} else if WriterId > 0 {
		cWriter = getWriterById(WriterId)
	} else if WriterId < 1 || cWriter != psWriter {
		http.Redirect(w, r, "/", 303)
		return
	}
	sentFlash := make(map[string]interface{})
	submit := r.FormValue("submit")
	username := r.FormValue("username")
	email := r.FormValue("email")
	quote := r.FormValue("quote")
	per := r.FormValue("permission")

	var permission bool
	if per == "true" {
		permission = true
	}
	if submit == "Delete" {
		session.Destroy(w)
		// first we remove the profile pic
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		WriterPic := filepath.Join(wd, "static", "pic", "pros", strconv.Itoa(cWriter.Id), cWriter.Pic)
		if err := os.Remove(WriterPic); err != nil {
			fmt.Println(err)
		}
		if affect := deleteWriterById(cWriter.Id); affect < 1 {
			fmt.Println("Your Account Can't Be Deleted.")
		}
		// if successful
		if err := session.PutString(w, "Cong", "Your Account Is Deleted"); err != nil {
			fmt.Println(err)
		}
		w.Write([]byte("done"))
		return
	}
	fmt.Println("submit: ", submit)
	if submit == "DeleteImg" {
		fmt.Println("got to delete img")
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		WriterPic := filepath.Join(wd, "static", "pic", "pros", strconv.Itoa(cWriter.Id), cWriter.Pic)
		if err := os.Remove(WriterPic); err != nil {
			fmt.Println(err)
		}
		fmt.Println("writerid: ", cWriter.Id)
		if affect := deleteWritersPic(cWriter.Id); affect < 1 {
			fmt.Println("Your Picture Can't Be Deleted.")
		}
		// if successful
		w.Write([]byte("done"))
		return
	}

	if submit == "Update Password"{
		cPass := r.FormValue("cPass")
		newPass := r.FormValue("newPass")
		confirmPass := r.FormValue("confirmPass")
		hashed := cWriter.Password
		// check to see if the password is correct
		if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(cPass)); err != nil {
			sentFlash["Err"] = "Your Current Password Is Wrong."
			if err := session.PutObject(w, "sentFlash",sentFlash); err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/profile/"+cWriter.Name+"/edit", 303)
			return
		}

		if newPass != confirmPass {
			sentFlash["Err"] = "Your New Password Doesn't Match With The Confirm Password."
			if err := session.PutObject(w, "sentFlash",sentFlash); err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/profile/"+cWriter.Name+"/edit", 303)
			return
		}
		if len(newPass) < 8 || len(newPass) > 50 {
			sentFlash["Err"] = "The Length Of Your Password Should Be Between 8 And 50 Characters."
			if err := session.PutObject(w, "sentFlash",sentFlash); err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/profile/"+cWriter.Name+"/edit", 303)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}
		if affect := updatePass(cWriter.Id, string(hash)); affect < 1 {
			sentFlash["Err"] = "Your New Password Didn't Change, Please Try Again."
			if err := session.PutObject(w, "sentFlash",sentFlash); err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/profile/"+cWriter.Name+"/edit", 303)
			return
		}

		if err := session.PutString(w, "Cong", "Your Password Has Been Updated."); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/profile/"+cWriter.Name, 303)
		return
	}

	var maxFileSize int64 = 4 * 1000 * 1000 //limit upload file to 10m
	if r.ContentLength > maxFileSize {
		sentFlash["Err"] = "The Maximum Size Of Profile Picture Is 4M."
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/profile/"+strconv.Itoa(cWriter.Id)+"/edit", 303)
		return
	}

	file, FileHeader, err := r.FormFile("pic")
	if err != nil {
		fmt.Println(err)
	}

	var picName string
	if err != http.ErrMissingFile && len(FileHeader.Filename) != 0{
		if valid := detectFileType(file); !valid {
			sentFlash["Err"] = "Please Upload An Image, Other Types Are Not Supported."
			if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/profile/"+strconv.Itoa(cWriter.Id)+"/edit", 303)
			return
		}
		picName = processProPic(file, cWriter)
		defer file.Close()
	}

	upWriter := Writer{Id: cWriter.Id, Name: strings.TrimSpace(username), Email: strings.TrimSpace(email), Quote: strings.TrimSpace(quote), Permission: permission}
	if picName != "" {
		upWriter.Pic = picName
	} else {
		upWriter.Pic = cWriter.Pic
	}
	// if validation of form failed
	// we give it just a password, to avoid validation from returning err for password being empty
	if warnings := upWriter.UpValidate(); len(warnings) > 0 {
		sentFlash["Permission"] = permission
		sentFlash["Quote"] = quote
		sentFlash["Errs"] = warnings
		sentFlash["UserName"] = username
		sentFlash["Email"] = email
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/profile/"+cWriter.Name+"/edit", 303)
		return
	}

	// if validation was passed
	if affect := updateWriter(upWriter); affect < 1 {
		sentFlash["Errs"] = "Your Account Didn't Updated."
		if err := session.PutObject(w, "sentFlash", sentFlash); err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/profile/"+cWriter.Name+"/edit", 303)
		return
	}
	if err := session.PutString(w, "Cong", "Your Profile Has Been Updated."); err != nil {
		fmt.Println(err)
	}
	// in case user has changed username, we don't use the previous one.
	http.Redirect(w, r, "/profile/"+username, 302)
}
