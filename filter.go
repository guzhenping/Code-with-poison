package main

import (
	"fmt"
	"strings"
)

//| └─TableReader_21       | 583666.00 | root      | data:Selection_20                                                     |
//|   └─Selection_20       | 583666.00 | cop[tikv] | gt(Column#5, 1)                                                       |
//|     └─TableScan_19     | 583666.00 | cop[tikv] | table:trips, range:[-inf,+inf], keep order:false

func dataSetCheck(rawSQL string, args float32) {
	sqlStr := fmt.Sprintf("explain %s", rawSQL)
	explains, err := getExplains(sqlStr)
	if err != nil {
		return
	}
	explainsMap := make(map[string]string)
	explainsMapWithCount := make(map[string]float32)
	var isTableScanExist bool
	childFatherMap := getChildFatherMap(rawSQL)
	for _, v := range explains {
		node := getPrefixPath(v.Id)
		explainsMap[node] = v.Operator
		explainsMapWithCount[node] = v.Count
		// 包含 isIndexExist
		// 包含 TableScan
		if strings.Contains(v.Id, "TableScan") {
			isTableScanExist = true
		}
	}
	if isTableScanExist == true {
		for _, v := range explains {
			if strings.Contains(v.Id, "TableScan") {

				// deal with TableScan
				node := getPrefixPath(v.Id)
				if v.Count > args {
					// deal with Selection
					father := childFatherMap[node]
					fatherCount := explainsMapWithCount[father]
					if fatherCount > args {
						// deal with TableReader
						grandFather := childFatherMap[father]
						grandFatherCount := explainsMapWithCount[grandFather]
						if grandFatherCount > args {
							fmt.Println(fmt.Sprintf("[INFO] %s scan data too much", getPrefixPath(v.Id)))
						}
					}
				}
				if v.Count < args {
					fmt.Println(fmt.Sprintf("[INFO] %s scan data correct", getPrefixPath(v.Id)))
				}
			}
		}
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
				if !judgeIsIndexByColumnName(fatherOpera) {
					grandFather := childFatherMap[father]
					grandFatherOpera := explainsMap[grandFather]
					if judgeIsIndexByColumnName(grandFatherOpera) {
						fmt.Println("[INFO] please use index1")
					} else {
						fmt.Println("[INFO] need add index")
					}
				}
				//TODO:存在多个indexTable 和 TableScan 并存 ，仅 TableScan 存在的两种情况
				// else {
				//	fmt.Println("[INFO] please use index2")
				//}
			}
		}
	} else {
		fmt.Println("[INFO] Good, using IndexScan")
	}

}

