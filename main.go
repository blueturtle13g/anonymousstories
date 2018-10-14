package main

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"log"
)

func main(){
	defer DB.Close()
	Routes()
}

func getAverage(xs []float64) float64 {
	total := 0.0
	for _, v := range xs {
		total += v
	}
	return toFixed(total/float64(len(xs)), 2)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func sortWritersByRate(writers []Writer) (sortedWriters []Writer) {
	sort.Slice(writers, func(i, j int) bool {
		return writers[i].Rate < writers[j].Rate
	})
	for i := 1; i <= len(writers); i++ {
		sortedWriters = append(sortedWriters, writers[len(writers)-i])
	}
	return sortedWriters
}
func sortCatsByRate(cats []Cat) (sortedCats []Cat) {
	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Rate < cats[j].Rate
	})
	for i := 1; i <= len(cats); i++ {
		sortedCats = append(sortedCats, cats[len(cats)-i])
	}
	return sortedCats
}
func sortStoriesByDate(stories []Story) (sortedStories []Story) {
	sort.Slice(stories, func(i, j int) bool {
		return stories[i].Id < stories[j].Id
	})
	for i := 1; i <= len(stories); i++ {
		sortedStories = append(sortedStories, stories[len(stories)-i])
	}
	return sortedStories
}
func sortStoriesByRate(stories []Story) (sortedStories []Story) {
	sort.Slice(stories, func(i, j int) bool {
		return stories[i].Rate < stories[j].Rate
	})
	for i := 1; i <= len(stories); i++ {
		sortedStories = append(sortedStories, stories[len(stories)-i])
	}
	return sortedStories
}

func dbConn() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
	return db
}

func tagFinder(body string) (tags []string) {
	split := strings.Split(body, "#")
	for i, sentence := range split {
		if !strings.HasPrefix(body, "#") {
			if i == 0 {
				continue
			}
		}
		words := strings.Split(sentence, " ")
		word := words[0]
		if !(strings.Contains(word, "(") && strings.Contains(word, ")")) {
			if strings.HasSuffix(word, ".") || strings.HasSuffix(word, "!") || strings.HasSuffix(word, ")") || strings.HasSuffix(word, "(") || strings.HasSuffix(word, ":") {
				word = word[0 : len(word)-1]
			}
		}
		tags = append(tags, word)
	}
	return tags
}

func processTag(w http.ResponseWriter, writerId, storyId int, tags []string) {
	for _, v := range tags {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		var tagId int
		// we check to figure out if this tag_name already exists with
		// getting its tag_id, if tag_id is not bigger than one, so
		// there is no such tag_name
		if tagId = alreadyTag(v); tagId < 1 {
			tagId = insertTag(v)
		}
		_ = insertTagRel(tagId, writerId, storyId)
	}
}

func updateDependentRates(cStory Story, storyCat Cat, situation string) {
	// we want to check if this rate is just updating,
	// we don't increment the rateCount field.
	if situation != "up" {
		if writerAffect, storyAffect, catAffect := incrementRateCounts(cStory.By, cStory.Id, storyCat.Id); writerAffect < 1 || storyAffect < 1 || catAffect < 1 {
			fmt.Println("increment didn't word")
		}
	}
	if affect := updateStoryRate(cStory.Id); affect < 1 {
		fmt.Println("story rate didn't get updated.")
	}

	if affect := updateWriterRate(cStory.By); affect < 1 {
		fmt.Println("writer rate didn't get updated.")
	}
	if affect := updateCatRate(storyCat.Id); affect < 1 {
		fmt.Println("cat rate didn't get updated.")
	}

}

func processCat(cat, by string, writerId, storyId int) {
	var catId int
	if catId = alreadyCat(cat); catId < 1 {
		catId = insertCat(cat, by)
	}
	_ = insertCatRel(catId, writerId, storyId)
}

func getNow() string {
	splitTime := strings.Split(time.Now().String(), " ")
	now := splitTime[0]
	return now
}

func picSha(pic multipart.File) string {
	h := sha1.New()
	io.Copy(h, pic)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func processStoryPic(file multipart.File, story Story) (picName string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	previousPic := filepath.Join(wd, "static", "pic", "stories", strconv.Itoa(story.Id), story.Pic)
	if err := os.Remove(previousPic); err != nil {
		fmt.Println(err)
	}
	// create sha for file name
	picName = picSha(file) + ".jpg"
	// create new file
	if err := os.Mkdir(wd+string(filepath.Separator)+"static"+string(filepath.Separator)+
		"pic"+string(filepath.Separator)+"stories"+string(filepath.Separator)+strconv.Itoa(story.Id), 0777); err != nil {
		fmt.Println(err)
	}
	path := filepath.Join(wd, "static", "pic", "stories", strconv.Itoa(story.Id), picName)
	nf, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}
	defer nf.Close()
	// copy
	file.Seek(0, 0)
	io.Copy(nf, file)
	return picName
}

func processProPic(file multipart.File, writer Writer) (picName string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	previousPic := filepath.Join(wd, "static", "pic", "pros", strconv.Itoa(writer.Id), writer.Pic)
	if err := os.Remove(previousPic); err != nil {
		fmt.Println(err)
	}
	// create sha for file name
	picName = picSha(file) + ".jpg"
	// create new file
	if err := os.Mkdir(wd+string(filepath.Separator)+"static"+string(filepath.Separator)+
		"pic"+string(filepath.Separator)+"pros"+string(filepath.Separator)+strconv.Itoa(writer.Id), 0777); err != nil {
		fmt.Println(err)
	}
	path := filepath.Join(wd, "static", "pic", "pros", strconv.Itoa(writer.Id), picName)
	nf, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}
	defer nf.Close()
	// copy
	file.Seek(0, 0)
	io.Copy(nf, file)
	return picName
}

func detectFileType(file multipart.File) (valid bool) {
	buff := make([]byte, 512) // why 512 bytes ? see http://golang.org/pkg/net/http/#DetectContentType
	var err error
	_, err = file.Read(buff)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filetype := http.DetectContentType(buff)

	switch filetype {
	case "image/jpeg", "image/jpg":
		return true

	case "image/gif":
		return true

	case "image/png":
		return true
	default:
		return false
	}
}

func getUniqueInt(intSlice []int) []int {
	keys := make(map[int]bool)
	var list []int
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func getUniqueTag(TagSlice []Tag) []Tag {
	keys := make(map[Tag]bool)
	var list []Tag
	for _, entry := range TagSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
