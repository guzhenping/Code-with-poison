package main

import (
	"bufio"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"strings"
)

func init() {
	initDB()
}

func main() {
	// 过滤
	// [INFO] please use index
	//indexCheck("SELECT * FROM bikeshare.trips WHERE member_type = 123;")
	explain(indexCheck, "SELECT * FROM bikeshare.trips WHERE member_type = 123;", "case1 : 本身存在索引 但是未使用")
	// [INFO] need add index
	explain(indexCheck, "SELECT * FROM bikeshare.trips WHERE duration > 123;", "case2 : 该列不存在索引 可以加索引")
	// [INFO] Good, using IndexScan
	explain(indexCheck, "SELECT * FROM bikeshare.trips WHERE member_type = 'test';", "case3: 正常 正在使用索引")
	// [INFO] need add index
	explain(indexCheck, "SELECT * FROM bikeshare.trips WHERE start_station_number > 123;", "case4: 该列不存在索引 可以加索引")
	// [INFO] TableScan_16 scan data correct
	explainDataSetCheck(dataSetCheck, "SELECT * FROM bikeshare.trips WHERE member_type = 'test' union all SELECT * FROM trips WHERE start_station_number = 1", 300000, "case5:tableScan 数据量正常")
	// [INFO] TableScan_16 scan data correct
	// [INFO] TableScan_19 scan data too much
	explainDataSetCheck(dataSetCheck, "SELECT * FROM bikeshare.trips WHERE member_type = 'test' union all SELECT * FROM trips WHERE start_station_number > 1", 300000, "case6:局部 tableScan 数据量不正常")

	// 优化

	// [INFO] please use index: SELECT count(*) FROM trips use index(end_station_number_idx)  WHERE cast(end_station_number as char) > '123';
	explain(optimizeIndex, "SELECT count(*) FROM bikeshare.trips WHERE cast(end_station_number as char) > '123';", "case7:未使用索引 使用优化策略 输出优化后SQL")

	// [INFO] Good, using IndexScan
	explain(optimizeIndex, "SELECT count(*) FROM bikeshare.trips use index(end_station_number_idx)  WHERE cast(end_station_number as char) > '123';", "case8:索引使用正常")
}

func explain(f func(string), sql string, chinese string) {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)

	fmt.Println(chinese)
	fmt.Println()
	fmt.Println("目标SQL:")
	fmt.Println(sql)
	fmt.Println()
	fmt.Println("explain 信息:")
	printfExplain(sql)
	fmt.Println()
	fmt.Println("过滤与优化结果:")
	f(sql)
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println()
	fmt.Println()

}

func explainDataSetCheck(f func(string, float32), sql string, number float32, chinese string) {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	fmt.Println(chinese)
	fmt.Println()
	fmt.Println("目标SQL:")
	fmt.Println(sql)
	fmt.Println()
	fmt.Println("explain 信息:")
	printfExplain(sql)
	fmt.Println()
	fmt.Println("模拟阈值", number)
	fmt.Println("过滤与优化结果:")
	f(sql, number)
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println()
	fmt.Println()
}
