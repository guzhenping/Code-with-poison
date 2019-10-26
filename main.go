package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

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

func main() {
	conn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	DB, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println("connection to mysql failed:", err)
		return
	}

	DB.SetConnMaxLifetime(100 * time.Second) //最大连接周期，超时的连接就close
	DB.SetMaxOpenConns(100)

	err = DB.Ping()
	if err != nil {
		fmt.Println("ping err ", err)
	}
	sql := "EXPLAIN SELECT * FROM trips WHERE duration > 100"
	rows, err := DB.Query(sql)

	defer func() {
		if rows != nil {
			_ = rows.Close() //关闭掉未scan的sql连接
		}
	}()

	if err != nil {
		fmt.Printf("Query failed,err:%v\n", err)
		return
	}

	explain := new(Explain) //用new()函数初始化一个结构体对象
	for rows.Next() {

		err = rows.Scan(&explain.Id, &explain.Count, &explain.Task, &explain.Operator) //不scan会导致连接不释放
		if err != nil {
			fmt.Printf("Scan failed,err:%v\n", err)
			return
		}
		fmt.Println("output: ", *explain)
	}

}
