package main

import (
"fmt"

"github.com/pingcap/parser"
"github.com/pingcap/parser/ast"
_ "github.com/pingcap/tidb/types/parser_driver"
)

type MyVisitor struct{}

func (m MyVisitor) Enter(n ast.Node) (node ast.Node, skipChildren bool) {
	//fmt.Printf("%s, %T\n", n.Text(), n)
	switch v := n.(type) {
	case *ast.ColumnName:
		fmt.Println(v)
		fmt.Printf("ColumnName: %s:%s\n", v.Table, v.Name)
		//case *ast.TableName:
		//	fmt.Printf("TableName: %s\n", v.Name)
	}
	return n, false
}

func (m MyVisitor) Leave(n ast.Node) (node ast.Node, ok bool) {
	return n, true
}

// This example show how to parse a text sql into ast.
func main() {

	// 0. make sure import parser_driver implemented by TiDB(user also can implement own driver by self).
	// and add `import _ "github.com/pingcap/tidb/types/parser_driver"` in the head of file.

	// 1. Create a parser. The parser is NOT goroutine safe and should
	// not be shared among multiple goroutines. However, parser is also
	// heavy, so each goroutine should reuse its own local instance if
	// possible.
	p := parser.New()

	// 2. Parse a text SQL into AST([]ast.StmtNode).
	sql := "select  t1.a, t2.b, t3.c  from t1 join t2 on t1.xx =t2.xz join t3 on t1.xx =t3.xx where exists (select d from t4 where t4.xx =t1.xx)"
	//sql := "SELECT emp_no, first_name, last_name " +
	//	"FROM employees USE INDEX (last_name) " +
	//	"where last_name='Aamodt' and gender='F' and birth_date > '1960-01-01'"

	stmtNodes, _, err := p.Parse(sql, "", "")

	for _, stmtNode := range stmtNodes {
		stmtNode.Accept(MyVisitor{})
	}

	// 3. Use AST to do cool things.
	fmt.Println(stmtNodes[0], err)
}
