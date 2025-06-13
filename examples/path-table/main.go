package main

import (
	"fmt"
	"log"

	"github.com/k0kubun/pp/v3"
	"github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/cloudspannerecosystem/memefish/token"
)

// analyzePathTableExpr はPathTableExprを含むクエリを分析します
func analyzePathTableExpr(sql string) {
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

	// FROM句を探してTableExprを分析
	if queryStmt, ok := stmt.(*ast.QueryStatement); ok {
		analyzeQueryForTableExpr(queryStmt.Query, 0)
	}
	
	fmt.Println("\n詳細なAST構造:")
	pp.Print(stmt)
}

func analyzeQueryForTableExpr(queryExpr ast.QueryExpr, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	switch q := queryExpr.(type) {
	case *ast.Select:
		fmt.Printf("%sSelect found\n", indent)
		if q.From != nil {
			fmt.Printf("%sFrom.Source: %T\n", indent, q.From.Source)
			analyzeTableExpr(q.From.Source, depth+1)
		}
		
	case *ast.Query:
		fmt.Printf("%sQuery wrapper\n", indent)
		analyzeQueryForTableExpr(q.Query, depth+1)
	}
}

func analyzeTableExpr(tableExpr ast.TableExpr, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	fmt.Printf("%sTableExpr型: %T\n", indent, tableExpr)
	
	switch t := tableExpr.(type) {
	case *ast.PathTableExpr:
		fmt.Printf("%sPathTableExpr詳細:\n", indent)
		fmt.Printf("%s  Path: %T\n", indent, t.Path)
		if t.Path != nil {
			fmt.Printf("%s    Idents数: %d\n", indent, len(t.Path.Idents))
			for i, ident := range t.Path.Idents {
				fmt.Printf("%s      Ident[%d]: %s\n", indent, i, ident.Name)
			}
		}
		fmt.Printf("%s  Hint: %v\n", indent, t.Hint != nil)
		if t.As != nil {
			fmt.Printf("%s  As: %s\n", indent, t.As.Alias.Name)
		}
		fmt.Printf("%s  WithOffset: %v\n", indent, t.WithOffset != nil)
		fmt.Printf("%s  Sample: %v\n", indent, t.Sample != nil)
		
	case *ast.Join:
		fmt.Printf("%sJoin found\n", indent)
		fmt.Printf("%s  Op: %s\n", indent, t.Op)
		fmt.Printf("%s  Method: %s\n", indent, t.Method)
		fmt.Printf("%s  Left: %T\n", indent, t.Left)
		fmt.Printf("%s  Right: %T\n", indent, t.Right)
		analyzeTableExpr(t.Left, depth+1)
		analyzeTableExpr(t.Right, depth+1)
		
	default:
		fmt.Printf("%s  その他のTableExpr: %T\n", indent, t)
	}
}

func main() {
	examples := []string{
		// 単純なテーブル名
		"SELECT * FROM users",
		
		// スキーマ付きテーブル名
		"SELECT * FROM my_schema.users",
		
		// データベース.スキーマ.テーブル
		"SELECT * FROM my_database.my_schema.users",
		
		// AS句付き
		"SELECT * FROM users AS u",
		
		// バッククォート付きの識別子
		"SELECT * FROM `my-database`.`my-schema`.`my-table`",
		
		// JOIN
		"SELECT * FROM users u JOIN orders o ON u.id = o.user_id",
		
		// 複雑なPath
		"SELECT * FROM project_id.dataset_id.table_id AS t",
		
		// WITH OFFSET
		"SELECT * FROM UNNEST([1, 2, 3]) AS t WITH OFFSET AS pos",
		
		// Hint付き
		"SELECT * FROM users@{FORCE_INDEX=idx_name}",
		
		// Table sample
		"SELECT * FROM users TABLESAMPLE SYSTEM (10 PERCENT)",
	}

	for _, sql := range examples {
		analyzePathTableExpr(sql)
	}
	
	fmt.Println("\n=== PathTableExpr の理解 ===")
	fmt.Println("PathTableExpr は以下の場合に使用されます:")
	fmt.Println("1. 単純なテーブル名: users")
	fmt.Println("2. スキーマ付き: schema.table")  
	fmt.Println("3. フルパス: project.dataset.table")
	fmt.Println("4. バッククォート付き: `project-id`.`dataset-id`.`table-id`")
	fmt.Println()
	fmt.Println("PathTableExpr の構成要素:")
	fmt.Println("- Path: 識別子のリスト ([project, dataset, table])")
	fmt.Println("- Hint: テーブルヒント (@{FORCE_INDEX=...})")
	fmt.Println("- As: テーブルエイリアス (AS alias)")
	fmt.Println("- WithOffset: UNNEST用のOFFSET句")
	fmt.Println("- Sample: テーブルサンプリング句")
}