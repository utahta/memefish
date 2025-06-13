package main

import (
	"fmt"
	"log"

	"github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/cloudspannerecosystem/memefish/token"
)

// subQueryVisitor はSubQueryを検出してOrderBy/Limitをチェックするvisitor
type subQueryVisitor struct{}

func (v *subQueryVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.SubQuery:
		fmt.Printf("SubQuery found: %s\n", n.SQL())
		hasOrderByOrLimit := checkQueryExprForOrderByLimit(n.Query)
		if hasOrderByOrLimit {
			fmt.Printf("  -> OrderBy or Limit detected!\n")
		} else {
			fmt.Printf("  -> No OrderBy or Limit\n")
		}
	case *ast.SubQueryTableExpr:
		fmt.Printf("SubQueryTableExpr found: %s\n", n.SQL())
		hasOrderByOrLimit := checkQueryExprForOrderByLimit(n.Query)
		if hasOrderByOrLimit {
			fmt.Printf("  -> OrderBy or Limit detected!\n")
		} else {
			fmt.Printf("  -> No OrderBy or Limit\n")
		}
	case *ast.ScalarSubQuery:
		fmt.Printf("ScalarSubQuery found: %s\n", n.SQL())
		hasOrderByOrLimit := checkQueryExprForOrderByLimit(n.Query)
		if hasOrderByOrLimit {
			fmt.Printf("  -> OrderBy or Limit detected!\n")
		} else {
			fmt.Printf("  -> No OrderBy or Limit\n")
		}
	case *ast.ArraySubQuery:
		fmt.Printf("ArraySubQuery found: %s\n", n.SQL())
		hasOrderByOrLimit := checkQueryExprForOrderByLimit(n.Query)
		if hasOrderByOrLimit {
			fmt.Printf("  -> OrderBy or Limit detected!\n")
		} else {
			fmt.Printf("  -> No OrderBy or Limit\n")
		}
	case *ast.ExistsSubQuery:
		fmt.Printf("ExistsSubQuery found: %s\n", n.SQL())
		hasOrderByOrLimit := checkQueryExprForOrderByLimit(n.Query)
		if hasOrderByOrLimit {
			fmt.Printf("  -> OrderBy or Limit detected!\n")
		} else {
			fmt.Printf("  -> No OrderBy or Limit\n")
		}
	}
	return v
}

func (v *subQueryVisitor) VisitMany(nodes []ast.Node) ast.Visitor {
	return v
}

func (v *subQueryVisitor) Field(name string) ast.Visitor {
	return v
}

func (v *subQueryVisitor) Index(index int) ast.Visitor {
	return v
}

// checkSubQueryOrderByLimit はSubQueryでOrderByやLimitが使われているかをチェックします
func checkSubQueryOrderByLimit(node ast.Node) {
	visitor := &subQueryVisitor{}
	ast.Walk(node, visitor)
}

// checkQueryExprForOrderByLimit はQueryExprの中にOrderByやLimitがあるかをチェックします
func checkQueryExprForOrderByLimit(queryExpr ast.QueryExpr) bool {
	switch q := queryExpr.(type) {
	case *ast.Query:
		// QueryはOrderByやLimitフィールドを持っています
		return q.OrderBy != nil || q.Limit != nil
	case *ast.SubQuery:
		// 再帰的にチェック
		return checkQueryExprForOrderByLimit(q.Query)
	case *ast.CompoundQuery:
		// CompoundQueryの各Queryをチェック
		for _, query := range q.Queries {
			if checkQueryExprForOrderByLimit(query) {
				return true
			}
		}
		return false
	case *ast.Select, *ast.FromQuery:
		// SelectやFromQueryは直接OrderBy/Limitを持たない
		return false
	default:
		return false
	}
}

func main() {
	examples := []string{
		// OrderByありのSubQuery
		"SELECT * FROM (SELECT name FROM users ORDER BY name) AS u",
		// LimitありのSubQuery
		"SELECT * FROM (SELECT name FROM users LIMIT 10) AS u",
		// OrderByとLimitありのSubQuery
		"SELECT * FROM (SELECT name FROM users ORDER BY name LIMIT 10) AS u",
		// OrderByもLimitもないSubQuery
		"SELECT * FROM (SELECT name FROM users) AS u",
		// ScalarSubQuery
		"SELECT (SELECT COUNT(*) FROM users ORDER BY id LIMIT 1) AS count",
		// ExistsSubQuery
		"SELECT * FROM products WHERE EXISTS (SELECT 1 FROM orders WHERE orders.product_id = products.id ORDER BY created_at)",
		// ArraySubQuery
		"SELECT ARRAY(SELECT name FROM users ORDER BY name) AS names",
		// ネストしたSubQuery
		"SELECT * FROM (SELECT * FROM (SELECT id FROM users ORDER BY id) ORDER BY id LIMIT 5) AS nested",
	}

	fmt.Println("=== 詳細な分析 ===")
	for i, sql := range examples {
		fmt.Printf("\n=== Example %d ===\n", i+1)
		fmt.Printf("SQL: %s\n", sql)

		file := &token.File{Buffer: sql}
		parser := &memefish.Parser{
			Lexer: &memefish.Lexer{File: file},
		}

		stmt, err := parser.ParseStatement()
		if err != nil {
			log.Printf("Parse error: %v", err)
			continue
		}

		checkSubQueryOrderByLimit(stmt)
	}

	fmt.Println("\n\n=== ユーティリティ関数のテスト ===")
	for i, sql := range examples {
		fmt.Printf("\nExample %d: %s\n", i+1, sql)

		file := &token.File{Buffer: sql}
		parser := &memefish.Parser{
			Lexer: &memefish.Lexer{File: file},
		}

		stmt, err := parser.ParseStatement()
		if err != nil {
			log.Printf("Parse error: %v", err)
			continue
		}

		// 単純にチェック
		hasOrderByOrLimit := HasSubQueryWithOrderByOrLimit(stmt)
		fmt.Printf("  HasSubQueryWithOrderByOrLimit: %t\n", hasOrderByOrLimit)

		// OrderByやLimitを持つSubQueryを収集
		subQueries := GetSubQueriesWithOrderByOrLimit(stmt)
		fmt.Printf("  OrderBy/Limitを持つSubQuery数: %d\n", len(subQueries))
		for j, sq := range subQueries {
			fmt.Printf("    %d: %s\n", j+1, sq.SQL())
		}
	}
}