package main

import (
	"fmt"
	_ "github.com/lib/pq"
)

func alreadyName(writerName string, writerId int) int {

	rows, err := DB.Query("select id from writers where name = $1", writerName)
	if err != nil {
		fmt.Println(err)
	}

	var id int
	for rows.Next() {
		rows.Scan(&id)
	}
	if id == 0 || writerId == id {
		return 0
	}
	return 1
}

func alreadyEmail(email string, writerId int) int64 {

	rows, err := DB.Query("select id from writers where email = $1", email)
	if err != nil {
		fmt.Println(err)
	}

	var id int
	for rows.Next() {
		rows.Scan(&id)
	}
	if id == 0 || writerId == id {
		return 0
	}
	return 1
}

func checkWriterName(uORe string) (ItsId int) {

	rows, err := DB.Query("select id from writers where name = $1 or email = $1", uORe)
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
