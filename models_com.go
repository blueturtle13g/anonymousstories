package main

import (
	"fmt"
	_ "github.com/lib/pq"
)

type Com struct {
	Id, ComId, StoryId                        int
	Text, By, CreatedOn, UpdatedOn, DeletedOn string
}

func getWritersComs(WriterName string) (coms []Com) {
	rows, err := DB.Query("select * from coms where by= $1", WriterName)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		var com Com
		rows.Scan(&com.Id, &com.ComId, &com.StoryId, &com.Text, &com.By, &com.CreatedOn, &com.UpdatedOn, &com.DeletedOn)
		coms = append(coms, com)
	}
	return coms
}

func getStoriesComments(storyId int) (ItsComments []Com) {
	rows, err := DB.Query("SELECT * FROM coms where storyId = $1 order by id desc", storyId)
	if err != nil {
		fmt.Println(err)
	}
	var com Com
	for rows.Next() {
		err = rows.Scan(&com.Id, &com.ComId, &com.StoryId, &com.Text, &com.By, &com.CreatedOn, &com.UpdatedOn, &com.DeletedOn)
		if err != nil {
			fmt.Println(err)
		}
		ItsComments = append(ItsComments, com)
	}
	return ItsComments
}

func insertCom(com Com) (comId int) {
	err := DB.QueryRow("INSERT INTO coms(comId, storyId, text, by, createdOn) VALUES($1,$2,$3,$4,$5) returning id;", com.ComId, com.StoryId, com.Text, com.By, getNow()).Scan(&comId)
	if err != nil {
		fmt.Println(err)
	}
	return comId
}

func updateCom(id int, text string) int64 {
	stmt, err := DB.Prepare("update coms set text= $1, updatedOn=$2 where id= $3")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(text, getNow(), id)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}

func deleteStory(storyId int) int64 {
	stmt, err := DB.Prepare("delete from catRel where storyId=$1")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(storyId)
	if err != nil {
		fmt.Println(err)
	}

	catRelAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	if catRelAffect < 1 {
		fmt.Println("catRel didn't get changed.")
	}
	stmt, err = DB.Prepare("delete from tagRel where storyId=$1")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(storyId)
	if err != nil {
		fmt.Println(err)
	}

	tagRelAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	if tagRelAffect < 1 {
		fmt.Println("tagRel didn't get changed.")
	}
	stmt, err = DB.Prepare("delete from comrel where storyid=$1")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(storyId)
	if err != nil {
		fmt.Println(err)
	}

	comRelAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	if comRelAffect < 1 {
		fmt.Println("comRel didn't get changed.")
	}
	stmt, err = DB.Prepare("delete from stories where id=$1")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(storyId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}

func deleteComment(comId int) int64 {
	stmt, err := DB.Prepare("delete from coms where comId=$1")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(comId)
	if err != nil {
		fmt.Println(err)
	}

	dependentAffects, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	if dependentAffects == 0{
		fmt.Println("deletiong of rescoms failed.")
	}
	stmt, err = DB.Prepare("delete from coms where id=$1")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(comId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}
