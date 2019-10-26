package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

var DB *sql.DB

const (
	USERNAME = "root"
	PASSWORD = ""
	NETWORK  = "tcp"
	SERVER   = "39.100.84.79"
	PORT     = 4000
	DATABASE = "bikeshare"
)

//user表结构体定义
type Explain struct {
	Id       string  `json:"id" form:"id"`
	Count    float32 `json:"count" form:"count"`
	Task     string  `json:"task" form:"task"`
	Operator string  `json:"operator" form:"operator"`
}

func initDB() {
	conn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println("connection to tidb failed:", err)
		os.Exit(2)
		return
	}

	db.SetConnMaxLifetime(100 * time.Second) //最大连接周期，超时的连接就close
	db.SetMaxOpenConns(100)

	err = db.Ping()
	if err != nil {
		fmt.Println("ping db err ", err)
	}
	DB = db
}

func getDB() sql.DB {
	return *DB
}

func getExplains(sql string) ([]Explain, error) {
	DB := getDB()
	var explains []Explain
	rows, err := DB.Query(sql)

	defer func() {
		if rows != nil {
			_ = rows.Close() //关闭掉未scan的sql连接
		}
	}()

	if err != nil {
		return explains, err
	}

	for rows.Next() {
		explain := new(Explain)
		err = rows.Scan(&explain.Id, &explain.Count, &explain.Task, &explain.Operator) //不scan会导致连接不释放
		if err != nil {
			return explains, err
		}
		explains = append(explains, *explain)
	}
	return explains, nil
}
