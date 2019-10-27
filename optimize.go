package main

import (
	"fmt"
	"strings"
)

func optimizeIndex(rawSQL string) {

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
				if !judgeIsIndexByColumnName(fatherOpera) {
					grandFather := childFatherMap[father]
					grandFatherOpera := explainsMap[grandFather]
					if judgeIsIndexByColumnName(grandFatherOpera) {
						colunumNameList := getColumnNameByFatherOperator(grandFatherOpera)
						dbAndTableName := getDBAndTableNameByOperator(grandFatherOpera)
						// TODO: colunumNameList > 1 暂时不处理
						if len(colunumNameList) <= 1 {
							index := getIdxFromColumnName(colunumNameList[0],dbAndTableName)
							splitStrs := strings.SplitAfter(rawSQL,dbAndTableName)
							var optimizeStrs []string
							optimizeStrs = append(optimizeStrs, splitStrs[0])
							optimizeStrs = append(optimizeStrs, fmt.Sprintf("use index(%s)",index))
							optimizeStrs = append(optimizeStrs, splitStrs[1])
							fmt.Println("[INFO] please use index:", strings.Join(optimizeStrs," "))
						}
					} else {
						fmt.Println("[INFO] need add index")
					}
				}
			}
		}
	} else {
		fmt.Println("[INFO] Good, using IndexScan")
	}
}

