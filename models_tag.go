package main

import (
	"fmt"
	_ "github.com/lib/pq"
)

type Tag struct {
	Id              int
	Name, CreatedOn string
}
type TagRelations struct {
	TagId, StoryId, WriterId int
}

func getStoriesByTagId(TagId int) (stories []Story) {

	rows, err := DB.Query("select storyId from tagRel where tagId = $1", TagId)
	if err != nil {
		fmt.Println(err)
	}

	var ids []int

	for rows.Next() {
		var storyId int
		rows.Scan(&storyId)
		ids = append(ids, storyId)
	}

	ids = getUniqueInt(ids)
	stories = getStoriesByIds(ids)

	return stories
}

func getTagById(TagId int) (tag Tag) {

	rows, err := DB.Query("select * from tags where id = $1", TagId)
	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		rows.Scan(&tag.Id, &tag.Name, &tag.CreatedOn)
	}

	return tag
}

func searchTags(word string) (tags []Tag) {
	word = "%" + word + "%"

	rows, err := DB.Query("SELECT * FROM tags WHERE name LIKE $1", word)
	if err != nil {
		fmt.Println(err)
	}
	var tag Tag
	for rows.Next() {
		rows.Scan(&tag.Id, &tag.Name, &tag.CreatedOn)
		tags = append(tags, tag)
	}

	return tags
}

func getStoriesTags(storyId int) (ItsTags []Tag) {

	rows, err := DB.Query("SELECT * FROM tagRel where storyId = $1", storyId)
	if err != nil {
		fmt.Println(err)
	}
	var tagRel TagRelations
	var ItsTagRels []TagRelations
	for rows.Next() {
		err = rows.Scan(&tagRel.TagId, &tagRel.StoryId, &tagRel.WriterId)
		if err != nil {
			fmt.Println(err)
		}
		ItsTagRels = append(ItsTagRels, tagRel)
	}
	for _, v := range ItsTagRels {
		rows, err := DB.Query("SELECT * FROM tags where id = $1", v.TagId)
		if err != nil {
			fmt.Println(err)
		}
		var tag Tag
		for rows.Next() {
			err = rows.Scan(&tag.Id, &tag.Name, &tag.CreatedOn)
			if err != nil {
				fmt.Println(err)
			}
			ItsTags = append(ItsTags, tag)
		}
	}

	return ItsTags
}

func alreadyTag(tagName string) (ItsId int) {

	rows, err := DB.Query("select id from tags where name = $1", tagName)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		err = rows.Scan(&ItsId)
		if err != nil {
			fmt.Println(err)
		}
	}

	return ItsId
}

func insertTag(tagName string) (ItsId int) {

	err := DB.QueryRow(
		"insert into tags(name,createdOn) VALUES($1,$2) returning id;",
		tagName, getNow()).Scan(&ItsId)
	if err != nil {
		fmt.Println(err)
	}

	return ItsId
}

func insertTagRel(tagId, writerId, storyId int) (ItsId int) {

	err := DB.QueryRow(
		"insert into tagRel(tagId, storyId, writerId) values($1, $2, $3) returning tagId;",
		tagId, storyId, writerId).Scan(&ItsId)
	if err != nil {
		fmt.Println(err)
	}

	return ItsId
}
