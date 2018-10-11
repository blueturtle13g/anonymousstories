package main

import (
	"fmt"
	_ "github.com/lib/pq"
)

type Story struct {
	Id, RateCount, ViewCount                                   int
	Rate                                                       float64
	Title, Body, By, CreatedOn, UpdatedOn, DeletedOn, Cat, Pic string
}

func alreadyTitle(title string, storyId int) int64 {
	rows, err := DB.Query("select id from stories where title = $1", title)
	if err != nil {
		fmt.Println(err)
	}
	var id int
	for rows.Next() {
		rows.Scan(&id)
	}
	if id == 0 || storyId == id {
		return 0
	}

	return 1
}

func alreadyBody(body string, storyId int) int64 {

	rows, err := DB.Query("select id from stories where body = $1", body)
	if err != nil {
		fmt.Println(err)
	}
	var id int
	for rows.Next() {
		rows.Scan(&id)
	}
	if id == 0 || storyId == id {
		return 0
	}

	return 1
}

func updateStory(upStory Story) int64 {
	// first we delete previous relations to put new ones.
	stmt, err := DB.Prepare("delete from tagrel where storyId= $1")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(upStory.Id)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	if affect < 1 {
		fmt.Println("tagrelation didn't get cleared.")
	}
	stmt, err = DB.Prepare("delete from catrel where storyId= $1")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(upStory.Id)
	if err != nil {
		fmt.Println(err)
	}

	affect, err = res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	if affect < 1 {
		fmt.Println("catrelation didn't get cleared.")
	}

	stmt, err = DB.Prepare("update stories set title= $1, body= $2, by= $3, pic= $4, updatedOn= $5 where id= $6")
	if err != nil {
		fmt.Println(err)
	}
	res, err = stmt.Exec(upStory.Title, upStory.Body, upStory.By, upStory.Pic, getNow(), upStory.Id)
	if err != nil {
		fmt.Println(err)
	}

	affect, err = res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}

	return affect
}

func searchStories(word string) (stories []Story) {
	word = "%" + word + "%"

	rows, err := DB.Query("SELECT * FROM stories WHERE body LIKE $1 or title like $2", word, word)
	if err != nil {
		fmt.Println(err)
	}
	var story Story
	for rows.Next() {
		rows.Scan(&story.Id, &story.Title, &story.Body, &story.By, &story.Pic, &story.ViewCount, &story.RateCount, &story.CreatedOn, &story.UpdatedOn, &story.DeletedOn)
		ItsCat := getCatByStoryId(story.Id)
		story.Cat = ItsCat.Name
		ItsRates := getStoryRates(story.Id)
		story.Rate = ItsRates.Overall
		stories = append(stories, story)
	}

	return stories
}

func getStoriesByIds(ids []int) (stories []Story) {
	var story Story
	for _, id := range ids {
		rows, err := DB.Query("select * from stories where id = $1", id)
		if err != nil {
			fmt.Println(err)
		}
		for rows.Next() {
			rows.Scan(&story.Id, &story.Title, &story.Body, &story.By, &story.Pic, &story.ViewCount, &story.RateCount, &story.CreatedOn, &story.UpdatedOn, &story.DeletedOn)
			ItsCat := getCatByStoryId(story.Id)
			story.Cat = ItsCat.Name
			ItsRate := getStoryRates(story.Id)
			story.Rate = ItsRate.Overall
		}

		stories = append(stories, story)
	}
	return stories
}

func getAllStories(order, priority string) (stories []Story) {
	rows, err := DB.Query("select * from stories order by id desc")
	if err != nil {
		fmt.Println(err)
	}
	var story Story
	for rows.Next() {
		rows.Scan(&story.Id, &story.Title, &story.Body, &story.By, &story.Pic, &story.ViewCount, &story.RateCount, &story.CreatedOn, &story.UpdatedOn, &story.DeletedOn)
		ItsCat := getCatByStoryId(story.Id)
		story.Cat = ItsCat.Name
		ItsRate := getStoryRates(story.Id)
		story.Rate = ItsRate.Overall
		stories = append(stories, story)
	}
	switch order {
	case "rate":
		stories = sortStoriesByRate(stories)
		if priority == "asc" {
			for i, j := 0, len(stories)-1; i < j; i, j = i+1, j-1 {
				stories[i], stories[j] = stories[j], stories[i]
			}
		}
		return stories
	case "date":
		if priority == "asc" {
			for i, j := 0, len(stories)-1; i < j; i, j = i+1, j-1 {
				stories[i], stories[j] = stories[j], stories[i]
			}
		}
		return stories
	default:
		return stories
	}
}

func getStoryById(storyId int) (story Story, ItsComments []Com, ItsTags []Tag) {
	rows, err := DB.Query("SELECT * FROM stories where id = $1", storyId)
	if err != nil {
		fmt.Println(err)
		return
	}
	for rows.Next() {
		rows.Scan(&story.Id, &story.Title, &story.Body, &story.By, &story.Pic, &story.ViewCount, &story.RateCount, &story.CreatedOn, &story.UpdatedOn, &story.DeletedOn)
	}

	ItsCat := getCatByStoryId(story.Id)
	story.Cat = ItsCat.Name
	ItsRate := getStoryRates(story.Id)
	story.Rate = ItsRate.Overall
	ItsComments = getStoriesComments(storyId)
	ItsTags = getStoriesTags(storyId)
	return story, ItsComments, ItsTags
}

func insertStory(story Story) (ItsId int) {
	err := DB.QueryRow(
		"INSERT INTO stories(title,body,by,createdOn) VALUES($1,$2,$3,$4) returning id;",
		story.Title, story.Body, story.By, getNow()).Scan(&ItsId)
	if err != nil {
		fmt.Println(err)
	}

	err = DB.QueryRow(
		"INSERT INTO storyrates VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) returning storyId;",
		ItsId, 0, 0, 0, 0, 0, 0, 0, 0, 0).Scan(&ItsId)
	if err != nil {
		fmt.Println(err)
	}

	return ItsId
}
