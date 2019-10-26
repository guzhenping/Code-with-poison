package main

import (
	_ "github.com/go-sql-driver/mysql"
)
var tableName string

func init() {
	initDB()
	tableName = "trips"
}

func main() {

	var test = true
	if !test {

	}
	if test {
		// 过滤
		// [INFO] please use index
		indexCheck("SELECT * FROM trips WHERE member_type = 123;")
		indexCheck("SELECT * FROM trips WHERE duration = 123;")
		// [INFO] Good, using IndexScan
		indexCheck("SELECT * FROM trips WHERE member_type = 'test';")
		// [INFO] need add index
		indexCheck("SELECT * FROM trips WHERE start_station_number = 123;")
		// [INFO] TableScan_16 scan data correct
		dataSetCheck("SELECT * FROM trips WHERE member_type = 'heh' union all SELECT * FROM trips WHERE start_station_number = 1",300000)
		// [INFO] TableScan_16 scan data correct
		// [INFO] TableScan_19 scan data too much
		dataSetCheck("SELECT * FROM trips WHERE member_type = 'heh' union all SELECT * FROM trips WHERE start_station_number > 1",300000)

		// 优化
		// [INFO] Good, using IndexScan
		optimizeIndex("SELECT count(*) FROM trips use index(end_station_number_idx)  WHERE cast(end_station_number as char) > '123';")
		// [INFO] please use index: SELECT count(*) FROM trips use index(end_station_number_idx)  WHERE cast(end_station_number as char) > '123';
		optimizeIndex("SELECT count(*) FROM trips WHERE cast(end_station_number as char) > '123';")
	}

	// 优化

}
