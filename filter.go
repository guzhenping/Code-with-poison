package main

import (
	"fmt"
	"strings"
)

func dataSetCheck(rawSQL string) {
	sqlStr := fmt.Sprintf("explain %s", rawSQL)
	explains, err := getExplains(sqlStr)
	if err != nil {
		return
	}
	explainsMap := make(map[string]string)

	var isIndexExist bool
	var isTableScanExist bool
	childFatherMap := getChildFatherMap(rawSQL)
	for _, v := range explains {
		node := getPrefixPath(v.Id)
		explainsMap[node] = v.Operator
		// 包含 isIndexExist
		if strings.Contains(v.Id, "IndexScan") {
			isIndexExist = true
		}
		// 包含 TableScan
		if strings.Contains(v.Id, "TableScan") {
			isTableScanExist = true
		}
	}
	if isIndexExist == false && isTableScanExist == true {

	}
}
func indexCheck(rawSQL string) {
	sqlStr := fmt.Sprintf("explain %s", rawSQL)
	explains, err := getExplains(sqlStr)
	if err != nil {
		return
	}
	explainsMap := make(map[string]string)

	var isIndexExist bool
	var isTableScanExist bool
	childFatherMap := getChildFatherMap(rawSQL)
	for _, v := range explains {
		node := getPrefixPath(v.Id)
		explainsMap[node] = v.Operator
		// 包含 isIndexExist
		if strings.Contains(v.Id, "IndexScan") {
			isIndexExist = true
		}
		// 包含 TableScan
		if strings.Contains(v.Id, "TableScan") {
			isTableScanExist = true
		}
	}

	if isIndexExist == false && isTableScanExist == true {
		for _, v := range explains {
			if strings.Contains(v.Id, "TableScan") {
				node := getPrefixPath(v.Id)
				father := childFatherMap[node]
				fatherOpera := explainsMap[father]
				if !judgeIsIndexByColumnIds(fatherOpera) {
					fmt.Println("[INFO] need add index")
				} else {
					fmt.Println("[INFO] please use index")
				}
			}
		}
	} else {
		fmt.Println("[INFO] IndexScan is ok")
	}
}
