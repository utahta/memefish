package main

import (
	"fmt"
	"log"

	"github.com/k0kubun/pp/v3"
	"github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/cloudspannerecosystem/memefish/token"
)

// analyzeQueryStructure はクエリの構造を詳しく分析します
func analyzeQueryStructure(sql string) {
	fmt.Printf("\n=== SQL: %s ===\n", sql)
	
	file := &token.File{Buffer: sql}
	parser := &memefish.Parser{
		Lexer: &memefish.Lexer{File: file},
	}

	stmt, err := parser.ParseStatement()
	if err != nil {
		log.Printf("Parse error: %v", err)
		return
	}

	// トップレベルの構造を表示
	fmt.Printf("トップレベルのStatement型: %T\n", stmt)
	
	if queryStmt, ok := stmt.(*ast.QueryStatement); ok {
		fmt.Printf("QueryStatement.Query型: %T\n", queryStmt.Query)
		analyzeQueryExpr(queryStmt.Query, "QueryStatement.Query", 0)
	}
	
	// より詳細な構造を表示
	fmt.Println("\n詳細なAST構造:")
	pp.Print(stmt)
}

// analyzeQueryExpr はQueryExprを再帰的に分析します
func analyzeQueryExpr(queryExpr ast.QueryExpr, context string, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	fmt.Printf("%s%s: %T\n", indent, context, queryExpr)
	
	switch q := queryExpr.(type) {
	case *ast.Query:
		fmt.Printf("%s  With: %v\n", indent, q.With != nil)
		fmt.Printf("%s  OrderBy: %v\n", indent, q.OrderBy != nil)
		fmt.Printf("%s  Limit: %v\n", indent, q.Limit != nil)
		fmt.Printf("%s  ForUpdate: %v\n", indent, q.ForUpdate != nil)
		fmt.Printf("%s  PipeOperators: %d個\n", indent, len(q.PipeOperators))
		
		analyzeQueryExpr(q.Query, "Query.Query", depth+1)
		
	case *ast.Select:
		fmt.Printf("%s  Results: %d個\n", indent, len(q.Results))
		fmt.Printf("%s  From: %v\n", indent, q.From != nil)
		fmt.Printf("%s  Where: %v\n", indent, q.Where != nil)
		fmt.Printf("%s  GroupBy: %v\n", indent, q.GroupBy != nil)
		fmt.Printf("%s  Having: %v\n", indent, q.Having != nil)
		
	case *ast.CompoundQuery:
		fmt.Printf("%s  Op: %s\n", indent, q.Op)
		fmt.Printf("%s  AllOrDistinct: %s\n", indent, q.AllOrDistinct)
		fmt.Printf("%s  Queries: %d個\n", indent, len(q.Queries))
		
		for i, query := range q.Queries {
			analyzeQueryExpr(query, fmt.Sprintf("CompoundQuery[%d]", i), depth+1)
		}
		
	case *ast.SubQuery:
		fmt.Printf("%s  SubQueryの中身:\n", indent)
		analyzeQueryExpr(q.Query, "SubQuery.Query", depth+1)
		
	case *ast.FromQuery:
		fmt.Printf("%s  FromQueryの中身:\n", indent)
		if q.From != nil {
			fmt.Printf("%s  From: %T\n", indent, q.From)
		}
		
	default:
		fmt.Printf("%s  その他の型: %T\n", indent, q)
	}
}

func main() {
	examples := []string{
		// 1. 単純なSelect
		"SELECT name FROM users",
		
		// 2. SelectをQueryでラップ（OrderBy付き）
		"SELECT name FROM users ORDER BY name",
		
		// 3. SubQuery
		"SELECT * FROM (SELECT name FROM users) AS u",
		
		// 4. SubQueryにOrderBy
		"SELECT * FROM (SELECT name FROM users ORDER BY name) AS u",
		
		// 5. CompoundQuery（UNION）
		"SELECT name FROM users UNION ALL SELECT name FROM customers",
		
		// 6. CompoundQueryにOrderBy
		"SELECT name FROM users UNION ALL SELECT name FROM customers ORDER BY name",
		
		// 7. WITH句付き
		"WITH user_names AS (SELECT name FROM users) SELECT * FROM user_names",
		
		// 8. 複雑なネスト
		"SELECT * FROM (SELECT name FROM (SELECT * FROM users ORDER BY id) ORDER BY name LIMIT 5) AS nested",
	}

	for _, sql := range examples {
		analyzeQueryStructure(sql)
	}
	
	fmt.Println("\n=== 構造の理解 ===")
	fmt.Println("1. ルートレベルでは常に ast.QueryStatement が来ます")
	fmt.Println("2. QueryStatement.Query は QueryExpr インターフェース型です")
	fmt.Println("3. QueryExpr の実装型:")
	fmt.Println("   - ast.Query: OrderBy/Limit/WITH/PipeOperatorsを持つクエリラッパー")
	fmt.Println("   - ast.Select: 基本的なSELECT文")
	fmt.Println("   - ast.SubQuery: 括弧付きクエリ")
	fmt.Println("   - ast.CompoundQuery: UNION/INTERSECT/EXCEPT")
	fmt.Println("   - ast.FromQuery: FROM句のクエリ")
	fmt.Println("4. OrderBy/Limitは ast.Query でのみ管理されます")
	fmt.Println("5. ast.Query は再帰的にネストできます（Query.Query が別の Query の場合）")
}