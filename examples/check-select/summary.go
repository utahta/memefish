package main

import (
	"fmt"
	"log"

	"github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/token"
)

// demonstrateSelectOrderByLimitCheck はSelectでOrderBy/Limitチェックの使用例を示します
func demonstrateSelectOrderByLimitCheck() {
	fmt.Println("=== ast.SelectでOrderBy/Limitをチェックする方法 ===\n")

	examples := []struct {
		sql         string
		description string
	}{
		{"SELECT name FROM users", "OrderBy/Limitなし"},
		{"SELECT name FROM users ORDER BY name LIMIT 10", "OrderBy/Limitあり"},
		{"SELECT * FROM (SELECT name FROM users ORDER BY name) AS u", "SubQuery内でOrderBy"},
		{"SELECT (SELECT COUNT(*) FROM users LIMIT 1) AS count", "ScalarSubQuery内でLimit"},
		{"SELECT name FROM users UNION ALL SELECT name FROM customers ORDER BY name", "UNION with OrderBy"},
	}

	for i, example := range examples {
		fmt.Printf("%d. %s\n", i+1, example.description)
		fmt.Printf("   SQL: %s\n", example.sql)

		// パース
		file := &token.File{Buffer: example.sql}
		parser := &memefish.Parser{
			Lexer: &memefish.Lexer{File: file},
		}

		stmt, err := parser.ParseStatement()
		if err != nil {
			log.Printf("   エラー: %v\n", err)
			continue
		}

		// チェック
		hasOrderByOrLimit := HasAnySelectWithOrderByOrLimit(stmt)
		selectsWithOrderByOrLimit := FindSelectsWithOrderByOrLimit(stmt)

		fmt.Printf("   結果: OrderBy/Limitを持つSelect = %t\n", hasOrderByOrLimit)
		
		if len(selectsWithOrderByOrLimit) > 0 {
			fmt.Printf("   詳細:\n")
			for j, info := range selectsWithOrderByOrLimit {
				fmt.Printf("     Select %d [%s]: %s\n", j+1, info.ContextType, info.Select.SQL())
				if info.HasOrderBy {
					fmt.Printf("       -> OrderBy: %s\n", info.OrderBySQL)
				}
				if info.HasLimit {
					fmt.Printf("       -> Limit: %s\n", info.LimitSQL)
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("重要なポイント:")
	fmt.Println("1. ast.Select自体にはOrderBy/Limitフィールドがありません")
	fmt.Println("2. OrderBy/Limitは親のast.Queryによって管理されます")
	fmt.Println("3. SelectとQueryの関係を正しく追跡する必要があります")
	fmt.Println("4. SubQuery、ScalarSubQuery、ArraySubQuery、ExistsSubQueryなどのコンテキストを考慮する必要があります")
}

func main() {
	demonstrateSelectOrderByLimitCheck()
}