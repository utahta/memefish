package main

import (
	"fmt"
	"log"

	"github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/cloudspannerecosystem/memefish/token"
)

// quickAnalyze はクイック分析用の関数です
func quickAnalyze(sql string) {
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

	if queryStmt, ok := stmt.(*ast.QueryStatement); ok {
		if selectStmt, ok := queryStmt.Query.(*ast.Select); ok {
			if selectStmt.From != nil {
				if unnest, ok := selectStmt.From.Source.(*ast.Unnest); ok {
					fmt.Printf("UNNEST.Expr型: %T\n", unnest.Expr)
					fmt.Printf("UNNEST.As: %v\n", unnest.As != nil)
					fmt.Printf("UNNEST.WithOffset: %v\n", unnest.WithOffset != nil)
					
					if arrayLit, ok := unnest.Expr.(*ast.ArrayLiteral); ok {
						fmt.Printf("ArrayLiteral.Type: %T\n", arrayLit.Type)
						fmt.Printf("ArrayLiteral.Values数: %d\n", len(arrayLit.Values))
						if len(arrayLit.Values) > 0 {
							fmt.Printf("最初のValue型: %T\n", arrayLit.Values[0])
						}
					}
				}
			}
		}
	}
}

func main() {
	variations := []string{
		// 基本形
		"SELECT * FROM UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')])",
		
		// AS句付き
		"SELECT * FROM UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')]) AS t",
		
		// WITH OFFSET付き
		"SELECT * FROM UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')]) WITH OFFSET",
		
		// AS句とWITH OFFSET付き
		"SELECT * FROM UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')]) AS t WITH OFFSET AS pos",
		
		// 単純なARRAY
		"SELECT * FROM UNNEST([1, 2, 3])",
		
		// 型指定あり単純ARRAY
		"SELECT * FROM UNNEST(ARRAY<INT64>[1, 2, 3])",
		
		// STRING ARRAY
		"SELECT * FROM UNNEST(['hello', 'world'])",
		
		// ネストしたSTRUCT
		"SELECT * FROM UNNEST(ARRAY<STRUCT<a STRUCT<x INT64>, b STRING>>[(STRUCT(1), 'test')])",
	}

	for _, sql := range variations {
		quickAnalyze(sql)
	}
	
	fmt.Println("\n=== UNNEST パターンのまとめ ===")
	fmt.Println("1. 基本構造: SELECT * FROM UNNEST(expr)")
	fmt.Println("2. expr は通常 ArrayLiteral")  
	fmt.Println("3. ArrayLiteral は Type と Values を持つ")
	fmt.Println("4. オプション:")
	fmt.Println("   - AS alias: テーブルエイリアス")
	fmt.Println("   - WITH OFFSET [AS alias]: 行番号カラム")
	fmt.Println("5. STRUCT型の値:")
	fmt.Println("   - TupleStructLiteral: (val1, val2, ...)")
	fmt.Println("   - TypedStructLiteral: STRUCT<Type>(val1, val2, ...)")
	fmt.Println("   - TypelessStructLiteral: STRUCT(name1 val1, name2 val2, ...)")
}