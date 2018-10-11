package main

import (
	"fmt"
	_ "github.com/lib/pq"
)

type Message struct {
	Id                           int
	Name, Email, Text, CreatedOn string
}

func insertMsg(msg Message) (ItsId int) {
	err := DB.QueryRow(
		"INSERT INTO msg(name,email,text,createdOn) VALUES($1,$2,$3,$4) returning id;",
		msg.Name, msg.Email, msg.Text, getNow()).Scan(&ItsId)
	if err != nil {
		fmt.Println(err)
	}
	return ItsId
}
