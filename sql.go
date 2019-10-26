package main

import (
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"regexp"
	"strings"
)

type MyVisitor struct{}

var columnName []string

func (m MyVisitor) Enter(n ast.Node) (node ast.Node, skipChildren bool) {
	//fmt.Printf("%s, %T\n", n.Text(), n)
	switch v := n.(type) {
	case *ast.ColumnName:
		//fmt.Println(v)
		//fmt.Printf("ColumnName: %s:%s\n", v.Table, v.Name)
		if v.Name.String() != "" {
			columnName = append(columnName, v.Name.String())
		}
		//case *ast.TableName:
		//	fmt.Printf("TableName: %s\n", v.columnName)
	}
	return n, false
}

func (m MyVisitor) Leave(n ast.Node) (node ast.Node, ok bool) {
	return n, true
}

// This example show how to parse a text sql into ast.
func getColumnName(sqlStr string) []string {

	// 0. make sure import parser_driver implemented by TiDB(user also can implement own driver by self).
	// and add `import _ "github.com/pingcap/tidb/types/parser_driver"` in the head of file.

	// 1. Create a parser. The parser is NOT goroutine safe and should
	// not be shared among multiple goroutines. However, parser is also
	// heavy, so each goroutine should reuse its own local instance if
	// possible.
	p := parser.New()

	// 2. Parse a text SQL into AST([]ast.StmtNode).
	//sql := "explain SELECT * FROM trips WHERE member_type = 'sdsa';"
	//sql := "select  t1.a, t2.b, t3.c  from t1 join t2 on t1.xx =t2.xz join t3 on t1.xx =t3.xx where exists (select d from t4 where t4.xx =t1.xx)"
	//sql := "SELECT emp_no, first_name, last_name " +
	//	"FROM employees USE INDEX (last_name) " +
	//	"where last_name='Aamodt' and gender='F' and birth_date > '1960-01-01'"

	stmtNodes, _, _ := p.Parse(sqlStr, "", "")
	for _, stmtNode := range stmtNodes {
		stmtNode.Accept(MyVisitor{})
	}

	// 3. Use AST to do cool things.
	//fmt.Println(stmtNodes[0], err)
	return columnName
}

func getPrefixPath(str string) string {
	rex := regexp.MustCompile("(?s)(\\w+)")
	params := rex.FindStringSubmatch(str)
	return params[0]
}

func getColumnIds(str string) []string {
	var res []string
	rex := regexp.MustCompile("Column#(\\d+)")
	params := rex.FindAllStringSubmatch(str, -1)
	for _, v := range params {
		for _, n := range v {
			if !strings.Contains(n, "Column") {
				res = append(res, n)
			}
		}
	}
	return res
}
