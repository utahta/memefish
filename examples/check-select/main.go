package main

import (
	"fmt"
	"log"

	"github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/cloudspannerecosystem/memefish/token"
)

// selectVisitor はSelectを検出してOrderBy/Limitをチェックするvisitor
type selectVisitor struct{}

func (v *selectVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.Select:
		fmt.Printf("Select found: %s\n", n.SQL())
		fmt.Printf("  -> Select自体にはOrderBy/Limitはありません\n")
		fmt.Printf("  -> 親のQueryまたはSubQueryでOrderBy/Limitをチェックする必要があります\n")
	case *ast.Query:
		fmt.Printf("Query found: %s\n", n.SQL())
		// Queryの中にSelectがあるかチェック
		if select_, ok := n.Query.(*ast.Select); ok {
			fmt.Printf("  -> 内部にSelectを含むQuery\n")
			fmt.Printf("  -> Select: %s\n", select_.SQL())
			if n.OrderBy != nil {
				fmt.Printf("  -> OrderBy detected: %s\n", n.OrderBy.SQL())
			}
			if n.Limit != nil {
				fmt.Printf("  -> Limit detected: %s\n", n.Limit.SQL())
			}
			if n.OrderBy == nil && n.Limit == nil {
				fmt.Printf("  -> No OrderBy or Limit\n")
			}
		}
	case *ast.SubQuery:
		fmt.Printf("SubQuery found: %s\n", n.SQL())
		// SubQueryの中のSelectをチェック
		checkSelectInQueryExpr(n.Query, "SubQuery")
	case *ast.SubQueryTableExpr:
		fmt.Printf("SubQueryTableExpr found: %s\n", n.SQL())
		// SubQueryTableExprの中のSelectをチェック
		checkSelectInQueryExpr(n.Query, "SubQueryTableExpr")
	case *ast.ScalarSubQuery:
		fmt.Printf("ScalarSubQuery found: %s\n", n.SQL())
		checkSelectInQueryExpr(n.Query, "ScalarSubQuery")
	case *ast.ArraySubQuery:
		fmt.Printf("ArraySubQuery found: %s\n", n.SQL())
		checkSelectInQueryExpr(n.Query, "ArraySubQuery")
	case *ast.ExistsSubQuery:
		fmt.Printf("ExistsSubQuery found: %s\n", n.SQL())
		checkSelectInQueryExpr(n.Query, "ExistsSubQuery")
	}
	return v
}

func (v *selectVisitor) VisitMany(nodes []ast.Node) ast.Visitor {
	return v
}

func (v *selectVisitor) Field(name string) ast.Visitor {
	return v
}

func (v *selectVisitor) Index(index int) ast.Visitor {
	return v
}

// checkSelectInQueryExpr はQueryExprの中のSelectにOrderBy/Limitがあるかをチェック
func checkSelectInQueryExpr(queryExpr ast.QueryExpr, context string) {
	switch q := queryExpr.(type) {
	case *ast.Query:
		if select_, ok := q.Query.(*ast.Select); ok {
			fmt.Printf("  -> %s内のSelect: %s\n", context, select_.SQL())
			if q.OrderBy != nil {
				fmt.Printf("     -> OrderBy detected: %s\n", q.OrderBy.SQL())
			}
			if q.Limit != nil {
				fmt.Printf("     -> Limit detected: %s\n", q.Limit.SQL())
			}
			if q.OrderBy == nil && q.Limit == nil {
				fmt.Printf("     -> No OrderBy or Limit\n")
			}
		}
	case *ast.Select:
		fmt.Printf("  -> %s内の直接Select: %s\n", context, q.SQL())
		fmt.Printf("     -> Select自体にはOrderBy/Limitなし（親のQueryで管理）\n")
	case *ast.SubQuery:
		fmt.Printf("  -> %s内のネストしたSubQuery\n", context)
		checkSelectInQueryExpr(q.Query, context+"->SubQuery")
	case *ast.CompoundQuery:
		fmt.Printf("  -> %s内のCompoundQuery\n", context)
		for i, query := range q.Queries {
			fmt.Printf("     -> CompoundQuery[%d]をチェック\n", i)
			checkSelectInQueryExpr(query, fmt.Sprintf("%s->CompoundQuery[%d]", context, i))
		}
	}
}

// checkSelectOrderByLimit はSelectでOrderByやLimitが使われているかをチェックします
func checkSelectOrderByLimit(node ast.Node) {
	visitor := &selectVisitor{}
	ast.Walk(node, visitor)
}

func main() {
	examples := []string{
		// 単純なSelect（OrderBy/Limitなし）
		"SELECT name FROM users",
		// SelectをQueryでラップしてOrderBy追加
		"SELECT name FROM users ORDER BY name",
		// SelectをQueryでラップしてLimit追加
		"SELECT name FROM users LIMIT 10",
		// SelectをQueryでラップしてOrderByとLimit追加
		"SELECT name FROM users ORDER BY name LIMIT 10",
		// SubQuery内のSelect
		"SELECT * FROM (SELECT name FROM users ORDER BY name) AS u",
		// ScalarSubQuery内のSelect
		"SELECT (SELECT COUNT(*) FROM users ORDER BY id LIMIT 1) AS count",
		// CompoundQuery（UNION）
		"SELECT name FROM users UNION ALL SELECT name FROM customers ORDER BY name",
		// 複雑なネスト
		"SELECT * FROM (SELECT name FROM (SELECT * FROM users ORDER BY id) ORDER BY name LIMIT 5) AS nested",
	}

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

		checkSelectOrderByLimit(stmt)
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
		hasOrderByOrLimit := HasSelectWithOrderByOrLimit(stmt)
		fmt.Printf("  HasSelectWithOrderByOrLimit: %t\n", hasOrderByOrLimit)

		// 詳細なSelect情報を取得
		selectInfos := GetDetailedSelectsWithOrderByLimitInfo(stmt)
		fmt.Printf("  Select文の数: %d\n", len(selectInfos))
		
		for j, info := range selectInfos {
			fmt.Printf("    Select %d [%s]: %s\n", j+1, info.Context, info.Select.SQL())
			if info.OrderBy != nil {
				fmt.Printf("      -> OrderBy: %s\n", info.OrderBy.SQL())
			}
			if info.Limit != nil {
				fmt.Printf("      -> Limit: %s\n", info.Limit.SQL())
			}
			if info.OrderBy == nil && info.Limit == nil {
				fmt.Printf("      -> No OrderBy or Limit\n")
			}
		}

		// OrderByまたはLimitを持つSelectのみを取得
		selectsWithOrderByOrLimit := GetSelectsWithOrderByOrLimit(stmt)
		if len(selectsWithOrderByOrLimit) > 0 {
			fmt.Printf("  OrderBy/Limitを持つSelect数: %d\n", len(selectsWithOrderByOrLimit))
			for j, info := range selectsWithOrderByOrLimit {
				fmt.Printf("    OrderBy/Limit Select %d [%s]: %s\n", j+1, info.Context, info.Select.SQL())
			}
		}
	}

	fmt.Println("\n\n=== シンプルなユーティリティ関数のテスト ===")
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

		// シンプルなチェック
		hasAnyOrderByOrLimit := HasAnySelectWithOrderByOrLimit(stmt)
		fmt.Printf("  HasAnySelectWithOrderByOrLimit: %t\n", hasAnyOrderByOrLimit)

		// OrderByまたはLimitを持つSelectを取得
		selectsWithOrderByOrLimit := FindSelectsWithOrderByOrLimit(stmt)
		fmt.Printf("  OrderBy/Limitを持つSelect数: %d\n", len(selectsWithOrderByOrLimit))
		
		for j, info := range selectsWithOrderByOrLimit {
			fmt.Printf("    Select %d [%s]: %s\n", j+1, info.ContextType, info.Select.SQL())
			if info.HasOrderBy {
				fmt.Printf("      -> OrderBy: %s\n", info.OrderBySQL)
			}
			if info.HasLimit {
				fmt.Printf("      -> Limit: %s\n", info.LimitSQL)
			}
		}

		// 全てのSelectの概要
		allSelects := FindAllSelects(stmt)
		fmt.Printf("  全Select数: %d (OrderBy/Limitあり: %d, なし: %d)\n", 
			len(allSelects), 
			len(selectsWithOrderByOrLimit), 
			len(allSelects)-len(selectsWithOrderByOrLimit))
	}
}