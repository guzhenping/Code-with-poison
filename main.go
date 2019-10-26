package main

import (
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	initDB()
}
func main() {
	var test = false
	if !test {

	}
	if test {
		// 过滤
		// [INFO] please use index
		indexCheck("SELECT * FROM trips WHERE member_type = 123;")
		// [INFO] Good, using IndexScan
		indexCheck("SELECT * FROM trips WHERE member_type = 'test';")
		// [INFO] need add index
		indexCheck("SELECT * FROM trips WHERE start_station_number = 123;")
		// [INFO] TableScan_16 scan data correct
		dataSetCheck("SELECT * FROM trips WHERE member_type = 'heh' union all SELECT * FROM trips WHERE start_station_number = 1",300000)
		// [INFO] TableScan_16 scan data correct
		// [INFO] TableScan_19 scan data too much
		dataSetCheck("SELECT * FROM trips WHERE member_type = 'heh' union all SELECT * FROM trips WHERE start_station_number > 1",300000)

		getColumnNameByFatherOperator("eq(ying99_pomodel.b.type, \"LONG_WIN\"),eq(cast(bikeshare.trips.member_type), 123)")
	}

	// 优化

}
