package main

import (
	"fmt"
	_ "github.com/lib/pq"
)

type Writer struct {
	Id, RateCount                                                    int
	Rate                                                             float64
	Name, Email, Password, CreatedOn, DeletedOn, LastLog, Quote, Pic string
	Permission                                                       bool
}

func searchWriters(word string) (writers []Writer) {
	word = "%" + word + "%"

	rows, err := DB.Query("SELECT * FROM writers WHERE name LIKE $1", word)
	if err != nil {
		fmt.Println(err)

	}
	var writer Writer
	for rows.Next() {
		rows.Scan(&writer.Id, &writer.Name, &writer.Email, &writer.Password, &writer.Quote, &writer.Pic, &writer.Rate, &writer.RateCount, &writer.Permission, &writer.LastLog, &writer.CreatedOn, &writer.DeletedOn)
		writers = append(writers, writer)
	}
	return writers
}

func getWriterByName(Name string) (writer Writer) {

	rows, err := DB.Query("select * from writers where name = $1", Name)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		rows.Scan(&writer.Id, &writer.Name, &writer.Email, &writer.Password, &writer.Quote, &writer.Pic, &writer.Rate, &writer.RateCount, &writer.Permission, &writer.LastLog, &writer.CreatedOn, &writer.DeletedOn)
	}
	return writer
}

func getWritersByIds(ids []int) (writers []Writer) {
	var writer Writer
	for _, id := range ids {
		rows, err := DB.Query("select * from writers where id = $1", id)
		if err != nil {
			fmt.Println(err)
		}
		for rows.Next() {
			rows.Scan(&writer.Id, &writer.Name, &writer.Email, &writer.Password, &writer.Quote, &writer.Pic, &writer.Rate, &writer.RateCount, &writer.Permission, &writer.LastLog, &writer.CreatedOn, &writer.DeletedOn)
		}
		writers = append(writers, writer)
	}
	return writers
}

func insertWriter(writer Writer) (ItsId int) {
	err := DB.QueryRow(
		"INSERT INTO writers(name,email,password,createdon) VALUES($1,$2,$3,$4) returning id;",
		writer.Name, writer.Email, writer.Password, getNow()).Scan(&ItsId)
	if err != nil {
		fmt.Println(err)
	}
	return ItsId
}

func getAllWriters(order, priority string) (writers []Writer) {
	rows, err := DB.Query("select * from writers order by id desc")
	if err != nil {
		fmt.Println(err)

	}
	var writer Writer
	for rows.Next() {
		rows.Scan(&writer.Id, &writer.Name, &writer.Email, &writer.Password, &writer.Quote, &writer.Pic, &writer.Rate, &writer.RateCount, &writer.Permission, &writer.LastLog, &writer.CreatedOn, &writer.DeletedOn)
		writers = append(writers, writer)
	}
	switch order {
	case "rate":
		writers = sortWritersByRate(writers)
		if priority == "asc" {
			for i, j := 0, len(writers)-1; i < j; i, j = i+1, j-1 {
				writers[i], writers[j] = writers[j], writers[i]
			}
		}
		return writers
	case "date":
		if priority == "asc" {
			for i, j := 0, len(writers)-1; i < j; i, j = i+1, j-1 {
				writers[i], writers[j] = writers[j], writers[i]
			}
		}
		return writers
	default:
		return writers
	}
}

func deleteWriterById(WriterId int) int64 {
	if affect := updateDeletionDependencies(getWriterById(WriterId)); affect < 1 {
		fmt.Println("updateDeletionDependencies of story failed.")
	}
	stmt, err := DB.Prepare("update writers set name= $1, email= $2, pic= $3, permission= $4, password= $5, deletedOn= $6 where id= $7")
	if err != nil {
		fmt.Println(err)
	}
	res, err := stmt.Exec(WriterId, WriterId, "", false, WriterId, getNow(), WriterId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}
func deleteStoryPic(storyId int) int64{
	stmt, err := DB.Prepare("update stories set pic= $1 where id= $2")
	if err != nil {
		fmt.Println(err)
	}
	res, err := stmt.Exec("", storyId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}
func deleteWritersPic(WriterId int) int64 {
	stmt, err := DB.Prepare("update writers set pic= $1 where id= $2")
	if err != nil {
		fmt.Println(err)
	}
	res, err := stmt.Exec("", WriterId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}

func updateDeletionDependencies(writer Writer) int64 {
	stmt, err := DB.Prepare("update coms set by= $1 where by= $2")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(writer.Id, writer.Name)
	if err != nil {
		fmt.Println(err)
	}

	comAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	if comAffect < 1 {
		fmt.Println("updating coms by failed.")
	}
	stmt, err = DB.Prepare("update cats set by= $1 where by= $2")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(writer.Id, writer.Name)
	if err != nil {
		fmt.Println(err)
	}

	catAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	if catAffect < 1 {
		fmt.Println("updating cats by failed.")
	}
	stmt, err = DB.Prepare("update stories set by= $1 where by= $2")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(writer.Id, writer.Name)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}

func updateLastLog(writerId int) int64 {
	stmt, err := DB.Prepare("update writers set lastlog= $1 where id= $2")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(getNow(), writerId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}

func updatePass(WriterId int, password string) int64 {
	stmt, err := DB.Prepare("update writers set password= $1 where id= $2")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(password, WriterId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}

func updateWriter(upWriter Writer) int64 {
	// we should have previous information of the writer before updating
	// to be able to find its dependent tables and update them with new one,
	// e.g. if we use upWriter.Name to find its coms, it'll return nothing.
	// because there is no comment record with updated name
	notUpWriter := getWriterById(upWriter.Id)
	stmt, err := DB.Prepare("update writers set name= $1, email= $2, quote= $3, pic= $4, permission= $5 where id= $6")
	if err != nil {
		fmt.Println(err)
	}
	res, err := stmt.Exec(upWriter.Name, upWriter.Email, upWriter.Quote, upWriter.Pic, upWriter.Permission, upWriter.Id)
	if err != nil {
		fmt.Println(err)
	}
	WriterAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	stories := getWritersStories(notUpWriter.Id)
	for i, v := range stories {
		stmt, err = DB.Prepare("update stories set by= $1 where id= $2")
		if err != nil {
			fmt.Println(err)
		}
		res, err = stmt.Exec(upWriter.Name, v.Id)
		if err != nil {
			fmt.Println(err)
		}
		StoryAffect, err := res.RowsAffected()
		if err != nil {
			fmt.Println(err)
		}
		if StoryAffect < 1 {
			fmt.Println("StoryAffect less than one at ", i)
		}
	}

	coms := getWritersComs(notUpWriter.Name)
	for i, notUpWriter := range coms {
		stmt, err = DB.Prepare("update coms set by= $1 where id= $2")
		if err != nil {
			fmt.Println(err)
		}
		res, err = stmt.Exec(upWriter.Name, notUpWriter.Id)
		if err != nil {
			fmt.Println(err)
		}
		StoryAffect, err := res.RowsAffected()
		if err != nil {
			fmt.Println(err)
		}
		if StoryAffect < 1 {
			fmt.Println("ComAffect less than one at ", i)
		}
	}

	cats := getWritersCats(notUpWriter.Id)
	for i, notUpWriter := range cats {
		stmt, err = DB.Prepare("update cats set by= $1 where id= $2")
		if err != nil {
			fmt.Println(err)
		}
		res, err = stmt.Exec(upWriter.Name, notUpWriter.Id)
		if err != nil {
			fmt.Println(err)
		}
		StoryAffect, err := res.RowsAffected()
		if err != nil {
			fmt.Println(err)
		}
		if StoryAffect < 1 {
			fmt.Println("CatAffect less than one at ", i)
		}
	}
	return WriterAffect
}

func getWriterById(WriterId int) (writer Writer) {
	rows, err := DB.Query(
		"SELECT * FROM writers where id = $1", WriterId)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		rows.Scan(&writer.Id, &writer.Name, &writer.Email, &writer.Password, &writer.Quote, &writer.Pic, &writer.Rate, &writer.RateCount, &writer.Permission, &writer.LastLog, &writer.CreatedOn, &writer.DeletedOn)
	}
	return writer
}

func updateWriterRate(WriterName string) int64 {
	// first we get the id of all stories that this writer has written
	rows, err := DB.Query("SELECT id FROM stories WHERE by= $1", WriterName)
	if err != nil {
		fmt.Println(err)
	}
	var storyIds []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		storyIds = append(storyIds, id)
	}
	// then we range over each id to get its overall score
	var allOveralls []float64
	for _, storyId := range storyIds {
		rows, err = DB.Query("SELECT overAll FROM storyRates WHERE storyId= $1", storyId)
		if err != nil {
			fmt.Println(err)
		}
		for rows.Next() {
			var overAll float64
			rows.Scan(&overAll)
			allOveralls = append(allOveralls, overAll)
		}
	}
	// now it's time to get the average of all overall scores.
	overallAverage := getAverage(allOveralls)
	// and update the writer rate with this new rate that we've got.
	stmt, err := DB.Prepare("update writers set rate= $1 where name= $2")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(overallAverage, WriterName)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}

	return affect
}
