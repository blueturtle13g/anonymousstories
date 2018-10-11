package main

import (
	"fmt"
	_ "github.com/lib/pq"
)

type Rate struct {
	Id, StoryId, RaterId, Premise, Presentation, Structure, Characters, Theme, Style, Commercial int
	Overall                                                                                      float64
	CreatedOn, UpdatedOn                                                                         string
}
type StrRate struct {
	StoryId, RaterId, Premise, Presentation, Structure, Characters, Theme, Style, Commercial, Overall float64
	UpdatedOn                                                                                         string
}

func insertRates(raterId, storyId int, newRate Rate) (rateId int64) {
	err := DB.QueryRow("INSERT INTO rates(raterId, storyId, overall, premise,"+
		" presentation, structure, characters, theme, style, commercial, createdOn)"+
		"VALUES($1,$2,$3,$4,$5, $6, $7, $8, $9, $10, $11) returning id;",
		raterId, storyId, newRate.Overall, newRate.Premise, newRate.Presentation,
		newRate.Structure, newRate.Characters, newRate.Theme, newRate.Style,
		newRate.Commercial, getNow()).Scan(&rateId)
	if err != nil {
		fmt.Println(err)
	}
	return rateId
}
func updateRates(raterId, storyId int, newRate Rate) int64 {
	stmt, err := DB.Prepare("update rates set overall= $1, premise= $2, presentation= $3," +
		" structure= $4, characters= $5, theme= $6, style= $7, commercial= $8," +
		" updatedOn= $9 where raterId= $10 and storyId= $11")
	if err != nil {
		fmt.Println(err)
	}
	res, err := stmt.Exec(newRate.Overall, newRate.Premise, newRate.Presentation,
		newRate.Structure, newRate.Characters, newRate.Theme, newRate.Style,
		newRate.Commercial, getNow(), raterId, storyId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}

func alreadyRate(raterId, storyId int) int {
	rows, err := DB.Query("select id from rates where raterId = $1 and storyId = $2", raterId, storyId)
	if err != nil {
		fmt.Println(err)
	}

	var id int
	for rows.Next() {
		rows.Scan(&id)
	}
	return id
}

func updateStoryRate(StoryId int) int64 {
	rows, err := DB.Query("SELECT * FROM Rates WHERE storyId= $1", StoryId)
	if err != nil {
		fmt.Println(err)
	}
	var allRates []Rate
	for rows.Next() {
		var rate Rate
		rows.Scan(&rate.Id, &rate.RaterId, &rate.StoryId, &rate.Overall, &rate.Premise, &rate.Presentation, &rate.Structure, &rate.Characters, &rate.Theme, &rate.Style, &rate.Commercial, &rate.CreatedOn, &rate.UpdatedOn)
		allRates = append(allRates, rate)
	}
	var overalls, premises, presentations, structures, characters, themes, styles, commercials []float64
	for _, v := range allRates {
		overalls = append(overalls, v.Overall)
		premises = append(premises, float64(v.Premise))
		presentations = append(presentations, float64(v.Presentation))
		structures = append(structures, float64(v.Structure))
		characters = append(characters, float64(v.Characters))
		themes = append(themes, float64(v.Theme))
		styles = append(styles, float64(v.Style))
		commercials = append(commercials, float64(v.Commercial))
	}

	overallAverage := getAverage(overalls)
	premisesAverage := getAverage(premises)
	presentationAverage := getAverage(presentations)
	structureAverage := getAverage(structures)
	characterAverage := getAverage(characters)
	themeAverage := getAverage(themes)
	styleAverage := getAverage(styles)
	commercialAverage := getAverage(commercials)

	stmt, err := DB.Prepare("update storyRates set overall= $1, premise= $2, presentation= $3, structure= $4, characters= $5, theme= $6, style= $7, commercial= $8, updatedOn= $9 where storyId= $10")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(overallAverage, premisesAverage, presentationAverage, structureAverage, characterAverage, themeAverage, styleAverage, commercialAverage, getNow(), StoryId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}

	return affect
}

func getStoryRates(storyId int) (storyRates StrRate) {
	rows, err := DB.Query("select * from storyRates where storyId= $1", storyId)
	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		rows.Scan(&storyRates.StoryId, &storyRates.Overall, &storyRates.Premise,
			&storyRates.Presentation, &storyRates.Structure, &storyRates.Characters,
			&storyRates.Theme, &storyRates.Style, &storyRates.Commercial, &storyRates.UpdatedOn)
	}
	return storyRates
}

func incrementViewCount(storyId int) int64 {
	stmt, err := DB.Prepare("update stories SET viewCount = viewCount + 1 WHERE id= $1")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(storyId)
	if err != nil {
		fmt.Println(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return affect
}

func incrementRateCounts(writerName string, storyId, catId int) (int64, int64, int64) {
	stmt, err := DB.Prepare("update writers SET rateCount = rateCount + 1 WHERE name= $1")
	if err != nil {
		fmt.Println(err)
	}

	res, err := stmt.Exec(writerName)
	if err != nil {
		fmt.Println(err)
	}

	WriterAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	stmt, err = DB.Prepare("update stories SET rateCount = rateCount + 1 WHERE id= $1")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(storyId)
	if err != nil {
		fmt.Println(err)
	}

	StoryAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}

	stmt, err = DB.Prepare("update cats SET rateCount = rateCount + 1 WHERE id= $1")
	if err != nil {
		fmt.Println(err)
	}

	res, err = stmt.Exec(catId)
	if err != nil {
		fmt.Println(err)
	}

	CatAffect, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	return WriterAffect, StoryAffect, CatAffect
}
