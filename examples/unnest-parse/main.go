package main

import (
	"fmt"
	"log"

	"github.com/k0kubun/pp/v3"
	"github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/cloudspannerecosystem/memefish/token"
)

// analyzeUnnestQuery はUNNESTクエリの構造を詳しく分析します
func analyzeUnnestQuery(sql string) {
	fmt.Printf("=== SQL: %s ===\n", sql)
	
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
		analyzeQueryExpr(queryStmt.Query, 0)
	}
	
	fmt.Println("\n詳細なAST構造:")
	pp.Print(stmt)
}

// analyzeQueryExpr はQueryExprを再帰的に分析します
func analyzeQueryExpr(queryExpr ast.QueryExpr, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	fmt.Printf("%sQueryExpr型: %T\n", indent, queryExpr)
	
	switch q := queryExpr.(type) {
	case *ast.Select:
		fmt.Printf("%sSelect詳細:\n", indent)
		fmt.Printf("%s  Results: %d個\n", indent, len(q.Results))
		for i, result := range q.Results {
			fmt.Printf("%s    Result[%d]: %T\n", indent, i, result)
		}
		
		if q.From != nil {
			fmt.Printf("%s  From: %T\n", indent, q.From)
			if q.From.Source != nil {
				fmt.Printf("%s    Source: %T\n", indent, q.From.Source)
				analyzeTableExpr(q.From.Source, depth+2)
			}
		}
		
	case *ast.Query:
		fmt.Printf("%s  OrderBy: %v\n", indent, q.OrderBy != nil)
		fmt.Printf("%s  Limit: %v\n", indent, q.Limit != nil)
		analyzeQueryExpr(q.Query, depth+1)
	}
}

// analyzeTableExpr はTableExprを詳しく分析します
func analyzeTableExpr(tableExpr ast.TableExpr, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	fmt.Printf("%sTableExpr型: %T\n", indent, tableExpr)
	
	switch t := tableExpr.(type) {
	case *ast.Unnest:
		fmt.Printf("%sUnnest詳細:\n", indent)
		fmt.Printf("%s  Expr: %T\n", indent, t.Expr)
		analyzeExpr(t.Expr, depth+1)
		
		if t.As != nil {
			fmt.Printf("%s  As: %s\n", indent, t.As.Alias.Name)
		}
		if t.WithOffset != nil {
			if t.WithOffset.As != nil {
				fmt.Printf("%s  WithOffset: %s\n", indent, t.WithOffset.As.Alias.Name)
			} else {
				fmt.Printf("%s  WithOffset: (無名)\n", indent)
			}
		}
		
	default:
		fmt.Printf("%s  その他のTableExpr: %T\n", indent, t)
	}
}

// analyzeExpr は式を詳しく分析します
func analyzeExpr(expr ast.Expr, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	fmt.Printf("%sExpr型: %T\n", indent, expr)
	
	switch e := expr.(type) {
	case *ast.ArrayLiteral:
		fmt.Printf("%sArrayLiteral詳細:\n", indent)
		if e.Type != nil {
			fmt.Printf("%s  Type: %T\n", indent, e.Type)
			analyzeType(e.Type, depth+1)
		}
		fmt.Printf("%s  Values: %d個\n", indent, len(e.Values))
		for i, value := range e.Values {
			fmt.Printf("%s    Value[%d]: %T\n", indent, i, value)
			switch v := value.(type) {
			case *ast.TupleStructLiteral:
				fmt.Printf("%s      TupleStructLiteral.Values: %d個\n", indent, len(v.Values))
				for j, val := range v.Values {
					fmt.Printf("%s        Value[%d]: %T = %s\n", indent, j, val, val.SQL())
				}
			case *ast.TypedStructLiteral:
				fmt.Printf("%s      TypedStructLiteral.Values: %d個\n", indent, len(v.Values))
				for j, val := range v.Values {
					fmt.Printf("%s        Value[%d]: %T = %s\n", indent, j, val, val.SQL())
				}
			case *ast.TypelessStructLiteral:
				fmt.Printf("%s      TypelessStructLiteral.Values: %d個\n", indent, len(v.Values))
				for j, val := range v.Values {
					fmt.Printf("%s        Value[%d]: %T = %s\n", indent, j, val.SQL())
				}
			default:
				fmt.Printf("%s      Value内容: %s\n", indent, v.SQL())
			}
		}
		
	default:
		fmt.Printf("%s  式の内容: %s\n", indent, e.SQL())
	}
}

// analyzeType は型を詳しく分析します
func analyzeType(typ ast.Type, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	fmt.Printf("%sType型: %T\n", indent, typ)
	
	switch t := typ.(type) {
	case *ast.ArrayType:
		fmt.Printf("%sArrayType詳細:\n", indent)
		if t.Item != nil {
			fmt.Printf("%s  Item: %T\n", indent, t.Item)
			analyzeType(t.Item, depth+1)
		}
		
	case *ast.StructType:
		fmt.Printf("%sStructType詳細:\n", indent)
		fmt.Printf("%s  Fields: %d個\n", indent, len(t.Fields))
		for i, field := range t.Fields {
			fieldName := "<anonymous>"
			if field.Ident != nil {
				fieldName = field.Ident.Name
			}
			fmt.Printf("%s    Field[%d]: %s %T\n", indent, i, fieldName, field.Type)
		}
		
	case *ast.SimpleType:
		fmt.Printf("%sSimpleType: %s\n", indent, t.Name)
		
	default:
		fmt.Printf("%s  その他の型: %T\n", indent, t)
	}
}

func main() {
	// 分析対象のクエリ
	sql := "SELECT * FROM UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')])"
	
	analyzeUnnestQuery(sql)
	
	fmt.Println("\n=== 構造の理解 ===")
	fmt.Println("1. SELECT * FROM UNNEST(...) の基本構造:")
	fmt.Println("   QueryStatement -> Select -> From -> Unnest")
	fmt.Println("2. UNNEST の中身:")
	fmt.Println("   ArrayLiteral with typed array ARRAY<STRUCT<...>>[...]")
	fmt.Println("3. STRUCT型の定義:")
	fmt.Println("   StructType with fields: x INT64, y STRING")
	fmt.Println("4. 配列の値:")
	fmt.Println("   StructLiteral values: (1, 'foo'), (3, 'bar')")
}