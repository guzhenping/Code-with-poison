package main

import (
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	initDB()
}
func main() {

	// [INFO] please use index
	indexCheck("SELECT * FROM trips WHERE member_type = 123;")
	// [INFO] IndexScan is ok
	indexCheck("SELECT * FROM trips WHERE member_type = 'test';")
	// [INFO] need add index
	indexCheck("SELECT * FROM trips WHERE start_station_number = 123;")
	//
	//fmt.Println(getColumnName("explain SELECT * FROM trips WHERE member_type = 'sdsa' and hehe = 'asa';"))
	//indexs ,err := getIndex("trips")
	//if err!=nil {
	//	fmt.Println("getIndex get err",err)
	//}
	//for _,v := range indexs{
	//
	//}
}
