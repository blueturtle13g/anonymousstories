package main

import (
	"fmt"
	"regexp"
	"strings"
)

func (writer Writer) Validate(pwd, confirm string) (warnings []string) {
	if affect := alreadyName(writer.Name, 0); affect != 0 {
		warnings = append(warnings, "This username already exists.")
	}
	if affect := alreadyEmail(writer.Email, 0); affect != 0 {
		warnings = append(warnings, "This email already exists.")
	}
	if er := mailValidation(writer.Email); er != "" {
		warnings = append(warnings, er)
	}
	switch {
	case len(writer.Name) < 3:
		warnings = append(warnings, "Username cannot be less than 3.")
	case len(writer.Name) > 20:
		warnings = append(warnings, "Username cannot be more than 20 characters.")
	case strings.Contains(writer.Name, " "):
		warnings = append(warnings, "Username cannot have space between.")
	case len(pwd) < 8:
		warnings = append(warnings, "Password cannot be less than 8 characters.")
	case len(pwd) > 50:
		warnings = append(warnings, "Password cannot be more than 50 characters.")
	case pwd != confirm:
		warnings = append(warnings, "Password does not match with confirm-password.")
	}
	return warnings
}
func (writer Writer) UpValidate() (warnings []string) {
	if affect := alreadyName(writer.Name, writer.Id); affect > 0 {
		warnings = append(warnings, "This username already exists.")
	}
	if affect := alreadyEmail(writer.Email, writer.Id); affect > 0 {
		warnings = append(warnings, "This email already exists.")
	}
	if er := mailValidation(writer.Email); er != "" {
		warnings = append(warnings, er)
	}
	if er := nameValidation(writer.Name); er != "" {
		warnings = append(warnings, er)
	}
	switch {
	case len(writer.Name) < 3:
		warnings = append(warnings, "Username Cannot Be Less Than 3 Characters.")
	case len(writer.Quote) > 300:
		warnings = append(warnings, "The Length Of Your Quote Is Too Long(300).")
	case len(writer.Name) > 20:
		warnings = append(warnings, "Username Cannot Be More Than 20 Characters.")
	case strings.Contains(writer.Name, " "):
		warnings = append(warnings, "Username Cannot Have Space Between.")
	}
	return warnings
}

func (story Story) Validate(selectedCat, madeCat, condition string, tags []string) (warnings []string, cat string) {
	var storyId int
	if condition == "up" {
		storyId = story.Id
	} else {
		storyId = 0
	}
	if affect := alreadyTitle(story.Title, storyId); affect != 0 {
		warnings = append(warnings, "This Title already exists, please choose another title for your story.")
	}
	if affect := alreadyBody(story.Title, storyId); affect != 0 {
		warnings = append(warnings, "This Title already exists, please choose another title for your story.")
	}
	switch {
	case len(story.Title) < 2 || len(story.Title) > 40:
		warnings = append(warnings, "Title can't be less than 2 or more than 40 characters.")
	case (len(madeCat) < 2 || len(madeCat) > 40) && len(selectedCat) < 1:
		warnings = append(warnings, "The Name Of Category can't be less than 2 or more than 40 characters.")
	case len(story.Body) < 100:
		warnings = append(warnings, "Story can't be less than 100 characters.")
	case len(selectedCat) > 0 && len(madeCat) > 0:
		warnings = append(warnings, "Make your own category or choose from previous ones, you can't put your story in multiple categories.")
	case len(selectedCat) == 0 && len(madeCat) == 0:
		warnings = append(warnings, "Please choose a category or make one.")
	}
	if len(selectedCat) > 0 {
		return warnings, selectedCat
	} else {
		return warnings, madeCat
	}
}

func (msg Message) Validate() (warnings []string) {
	if len(msg.Name) < 3 || len(msg.Name) > 20 {
		warnings = append(warnings, "The length of your name can't be less than 3 or more than 20 characters.")
	}
	if len(msg.Text) == 0 {
		warnings = append(warnings, "You Are Sending Us An Empty Message, Please Check.")
	}
	if er := mailValidation(msg.Email); er != "" {
		warnings = append(warnings, er)
	}
	if er := nameValidation(msg.Name); er != "" {
		warnings = append(warnings, er)
	}
	return warnings
}

func mailValidation(email string) (er string) {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if re.MatchString(email) {
		return ""
	}
	return "you've entered an invalid email address, please check it."
}

func nameValidation(name string) (er string) {
	if valid, err := regexp.MatchString("^[a-zA-Z0]+([_ -]?[a-zA-Z0-9])*$", name); !valid {
		return "Please Insert A Valid Username Consist of English Letters."
	} else if err != nil {
		fmt.Println(err)
	}
	return ""
}
