package main

import (
	"database/sql"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"os"
	"strings"
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

//explain表结构体定义
type Explain struct {
	Id       string  `json:"id" form:"id"`
	Count    float32 `json:"count" form:"count"`
	Task     string  `json:"task" form:"task"`
	Operator string  `json:"operator" form:"operator"`
}

type ExplainWithDot struct {
	DotContent string `json:"dot_content"`
}

// index 表结构定义

type Describe struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   string
}
type Index struct {
	Table        string         `json:"table"`
	NonUnique    string         `json:"non_unique"`
	KeyName      string         `json:"key_name"`
	SeqInIndex   string         `json:"seq_in_index"`
	ColumnName   string         `json:"column_name"`
	Collation    string         `json:"collation"`
	Cardinality  string         `json:"cardinality"`
	SubPart      sql.NullString `json:"sub_part"`
	Packed       sql.NullString `json:"packed"`
	Null         sql.NullString `json:"null"`
	IndexType    sql.NullString `json:"index_type"`
	Comment      sql.NullString `json:"comment"`
	IndexComment sql.NullString `json:"index_comment"`
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
	err := DB.Ping()
	if err != nil {
		initDB()
	}
	return *DB
}

func getExplains(sql string) ([]Explain, error) {
	initDB()
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

func getExplainsWithDot(sql string) (ExplainWithDot, error) {
	DB := getDB()
	var explainWithDot ExplainWithDot
	rows, err := DB.Query(sql)

	defer func() {
		if rows != nil {
			_ = rows.Close() //关闭掉未scan的sql连接
		}
	}()

	if err != nil {
		return explainWithDot, err
	}

	for rows.Next() {
		ewt := new(ExplainWithDot)
		err = rows.Scan(&ewt.DotContent) //不scan会导致连接不释放
		if err != nil {
			return explainWithDot, err
		}
		explainWithDot = *ewt
	}
	return explainWithDot, nil
}

func getIndex(table string) ([]Index, error) {
	DB := getDB()
	var indesx []Index
	sqlStr := fmt.Sprintf(" SHOW INDEX FROM %s;", table)
	rows, err := DB.Query(sqlStr)

	defer func() {
		if rows != nil {
			_ = rows.Close() //关闭掉未scan的sql连接
		}
	}()

	if err != nil {
		return indesx, err
	}

	for rows.Next() {
		index := new(Index)
		err = rows.Scan(
			&index.Table,
			&index.NonUnique,
			&index.KeyName,
			&index.SeqInIndex,
			&index.ColumnName,
			&index.Collation,
			&index.Cardinality,
			&index.SubPart,
			&index.Packed,
			&index.Null,
			&index.IndexType,
			&index.Comment,
			&index.IndexComment,
		) //不scan会导致连接不释放
		if err != nil {
			return indesx, err
		}
		indesx = append(indesx, *index)
	}
	return indesx, nil

}

func getChildFatherMap(rawSQL string) map[string]string {
	childFatherMap := make(map[string]string)
	sqlStr := fmt.Sprintf("explain format = \"dot\" %s", rawSQL)
	explainsWithDot, err := getExplainsWithDot(sqlStr)
	if err != nil {
		return childFatherMap
	}
	dotStr := explainsWithDot.DotContent

	graphAst, _ := gographviz.ParseString(dotStr)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}
	for _, v := range graph.Edges.Edges {
		dst := strings.Trim(v.Dst, "\"")
		src := strings.Trim(v.Src, "\"")
		childFatherMap[dst] = src
	}
	return childFatherMap
}

func getDesr(table string) ([]Describe, error) {
	initDB()
	var describes []Describe
	sqlStr := fmt.Sprintf(" DESCRIBE %s;", table)
	rows, err := DB.Query(sqlStr)

	defer func() {
		if rows != nil {
			_ = rows.Close() //关闭掉未scan的sql连接
		}
	}()

	if err != nil {
		return describes, err
	}

	for rows.Next() {
		describe := new(Describe)
		err = rows.Scan(
			&describe.Field,
			&describe.Type,
			&describe.Null,
			&describe.Key,
			&describe.Default,
			&describe.Extra,
		) //不scan会导致连接不释放
		if err != nil {
			return describes, err
		}
		describes = append(describes, *describe)
	}
	return describes, nil

}

// 传递 explain 中的 operator
func judgeIsIndexByColumnName(operation string) bool {
	isIndex := false
	dbAndTableName := getDBAndTableNameByOperator(operation)
	if dbAndTableName == "" {
		return false
	}
	describes, err := getDesr(dbAndTableName)
	if err != nil {
		fmt.Println("getDesr err", err)
	}
	// member_type
	list := getColumnNameByFatherOperator(operation)
	for _, v := range list {
		for _,d := range describes {
			if d.Field == v && d.Key != ""{
				isIndex = true
			}
		}
	}
	return isIndex
}

func getIdxFromColumnName(name string,dbAndTable string) string {
	indexs,err := getIndex(dbAndTable)
	if err!=nil {
		fmt.Println("getIndex",err)
	}
	for _,v := range indexs{
		if v.ColumnName == name {
			return v.KeyName
		}
	}
	return ""
}
