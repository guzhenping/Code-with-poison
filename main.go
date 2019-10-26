package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	initDB()
}
func main() {
	sql := "EXPLAIN SELECT * FROM trips WHERE duration > 100"
	explains, err := getExplains(sql)
	if err != nil {
		fmt.Println("getExplain err", explains)
	}
	for _, v := range explains {
		fmt.Println(v)
	}
}
