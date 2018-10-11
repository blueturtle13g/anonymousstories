package main

import (
	"fmt"
	"sort"

	_ "github.com/lib/pq"
)

type Cat struct {
	Id, RateCount                  int
	Rate                           float64
	Name, By, CreatedOn, DeletedOn string
}
type CatRelations struct {
	CatId, WriterId, StoryId int
}

func getCatsByIds(ids []int) (cats []Cat) {
	for _, id := range ids {
		rows, err := DB.Query("select * from cats where id = $1", id)
		if err != nil {
			fmt.Println(err)
		}
		var cat Cat
		for rows.Next() {
			rows.Scan(&cat.Id, &cat.Name, &cat.By, &cat.Rate, &cat.RateCount, &cat.CreatedOn, &cat.DeletedOn)
		}
		cats = append(cats, cat)
	}

	return cats
}

func getWritersStories(WriterId int) (stories []Story) {
	rows, err := DB.Query("select storyId from catRel where writerId = $1", WriterId)
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
	return getStoriesByIds(ids)
}

func getWritersCats(WriterId int) (cats []Cat) {
	rows, err := DB.Query("select catId from catRel where writerId = $1", WriterId)
	if err != nil {
		fmt.Println(err)
	}
	var ids []int
	for rows.Next() {
		var catId int
		rows.Scan(&catId)
		ids = append(ids, catId)
	}
	ids = getUniqueInt(ids)
	return getCatsByIds(ids)
}

func getCatByName(CatName string) (cat Cat) {

	rows, err := DB.Query("select * from cats where name= $1", CatName)
	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		rows.Scan(&cat.Id, &cat.Name, &cat.By, &cat.Rate, &cat.RateCount, &cat.CreatedOn, &cat.DeletedOn)
	}
	return cat
}

func getCatById(CatId int) (cat Cat) {
	rows, err := DB.Query("select * from cats where id = $1", CatId)
	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		rows.Scan(&cat.Id, &cat.Name, &cat.By, &cat.Rate, &cat.RateCount, &cat.CreatedOn, &cat.DeletedOn)
	}
	return cat
}

func getAllCats(order, priority string) (cats []Cat) {
	rows, err := DB.Query("select * from cats order by id desc")

	if err != nil {
		fmt.Println(err)

	}
	var cat Cat
	for rows.Next() {
		rows.Scan(&cat.Id, &cat.Name, &cat.By, &cat.Rate, &cat.RateCount, &cat.CreatedOn, &cat.DeletedOn)
		cats = append(cats, cat)
	}
	switch order {
	case "rate":
		cats = sortCatsByRate(cats)
		if priority == "asc" {
			for i, j := 0, len(cats)-1; i < j; i, j = i+1, j-1 {
				cats[i], cats[j] = cats[j], cats[i]
			}
		}
		return cats
	case "date":
		if priority == "asc" {
			for i, j := 0, len(cats)-1; i < j; i, j = i+1, j-1 {
				cats[i], cats[j] = cats[j], cats[i]
			}
		}
		return cats
	default:
		return cats
	}
}

func insertCat(catName, by string) (catId int) {
	err := DB.QueryRow(
		"insert into cats(name, by, createdOn) VALUES($1,$2,$3) returning id;",
		catName, by, getNow()).Scan(&catId)
	if err != nil {
		fmt.Println(err)

	}

	return catId
}

func searchCats(word string) (cats []Cat) {
	word = "%" + word + "%"

	rows, err := DB.Query("SELECT * FROM cats WHERE name LIKE $1", word)
	if err != nil {
		fmt.Println(err)

	}
	cat := Cat{}
	for rows.Next() {
		rows.Scan(&cat.Id, &cat.Name, &cat.By, &cat.Rate, &cat.RateCount, &cat.CreatedOn, &cat.DeletedOn)
		cats = append(cats, cat)
	}

	return cats
}

func getStoriesByCatId(CatId int) (stories []Story) {
	rows, err := DB.Query("select storyId from catRel where catId = $1", CatId)
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
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))
	stories = getStoriesByIds(ids)
	return stories
}

func getCatByStoryId(storyId int) (ItsCat Cat) {
	rows, err := DB.Query("SELECT * FROM catRel where storyId = $1", storyId)
	if err != nil {
		fmt.Println(err)
	}
	var catRel CatRelations
	for rows.Next() {
		err = rows.Scan(&catRel.CatId, &catRel.StoryId, &catRel.WriterId)
		if err != nil {
			fmt.Println(err)
		}
	}
	ItsCat = getCatById(catRel.CatId)
	return ItsCat
}

func alreadyCat(catName string) (ItsId int) {
	rows, err := DB.Query("select id from cats where name = $1", catName)
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

func getWritersByCatId(CatId int) (writers []Writer) {
	rows, err := DB.Query("select writerId from catRel where catId = $1", CatId)
	if err != nil {
		fmt.Println(err)
	}

	var ids []int

	for rows.Next() {
		var writerId int
		rows.Scan(&writerId)
		ids = append(ids, writerId)
	}

	ids = getUniqueInt(ids)
	writers = getWritersByIds(ids)
	return writers
}

func insertCatRel(catId, writerId, storyId int) (ItsId int) {
	err := DB.QueryRow(
		"insert into catRel(catId, storyId, writerId) values($1, $2, $3) returning catId;",
		catId, storyId, writerId).Scan(&ItsId)
	if err != nil {
		fmt.Println(err)
	}

	return ItsId
}

func updateCatRate(catId int) int64 {
	// first we get the id of all stories that this belongs to this category
	rows, err := DB.Query("SELECT storyId FROM catrel WHERE catId= $1", catId)
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
	var allOverAlls []float64
	for _, storyId := range storyIds {
		rows, err = DB.Query("SELECT overAll FROM storyRates WHERE storyId= $1", storyId)
		if err != nil {
			fmt.Println(err)
		}
		for rows.Next() {
			var overAll float64
			rows.Scan(&overAll)
			allOverAlls = append(allOverAlls, overAll)
		}
	}
	// now it's time to get the average of all overall scores.
	overAllAverage := getAverage(allOverAlls)
	// and update the writer rate with this new rate that we've got.
	stmt, err := DB.Prepare("update cats set rate= $1 where id= $2")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(overAllAverage, catId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}

	return affect
}
