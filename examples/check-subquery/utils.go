package main

import (
	"github.com/cloudspannerecosystem/memefish/ast"
)

// HasSubQueryWithOrderByOrLimit は与えられたAST内のSubQueryにOrderByやLimitが含まれているかを返します
func HasSubQueryWithOrderByOrLimit(node ast.Node) bool {
	checker := &subQueryChecker{}
	ast.Walk(node, checker)
	return checker.found
}

// GetSubQueriesWithOrderByOrLimit は与えられたAST内のOrderByやLimitを含むSubQueryを全て返します
func GetSubQueriesWithOrderByOrLimit(node ast.Node) []ast.Node {
	collector := &subQueryCollector{}
	ast.Walk(node, collector)
	return collector.subQueries
}

// subQueryChecker はSubQueryにOrderByやLimitがあるかを単純にチェックするvisitor
type subQueryChecker struct {
	found bool
}

func (v *subQueryChecker) Visit(node ast.Node) ast.Visitor {
	if v.found {
		return nil // 既に見つかっている場合は探索を中止
	}

	var queryExpr ast.QueryExpr
	switch n := node.(type) {
	case *ast.SubQuery:
		queryExpr = n.Query
	case *ast.SubQueryTableExpr:
		queryExpr = n.Query
	case *ast.ScalarSubQuery:
		queryExpr = n.Query
	case *ast.ArraySubQuery:
		queryExpr = n.Query
	case *ast.ExistsSubQuery:
		queryExpr = n.Query
	default:
		return v
	}

	if checkQueryExprForOrderByLimit(queryExpr) {
		v.found = true
		return nil
	}

	return v
}

func (v *subQueryChecker) VisitMany(nodes []ast.Node) ast.Visitor {
	return v
}

func (v *subQueryChecker) Field(name string) ast.Visitor {
	return v
}

func (v *subQueryChecker) Index(index int) ast.Visitor {
	return v
}

// subQueryCollector はOrderByやLimitを含むSubQueryを収集するvisitor
type subQueryCollector struct {
	subQueries []ast.Node
}

func (v *subQueryCollector) Visit(node ast.Node) ast.Visitor {
	var queryExpr ast.QueryExpr
	var subQueryNode ast.Node

	switch n := node.(type) {
	case *ast.SubQuery:
		queryExpr = n.Query
		subQueryNode = n
	case *ast.SubQueryTableExpr:
		queryExpr = n.Query
		subQueryNode = n
	case *ast.ScalarSubQuery:
		queryExpr = n.Query
		subQueryNode = n
	case *ast.ArraySubQuery:
		queryExpr = n.Query
		subQueryNode = n
	case *ast.ExistsSubQuery:
		queryExpr = n.Query
		subQueryNode = n
	default:
		return v
	}

	if checkQueryExprForOrderByLimit(queryExpr) {
		v.subQueries = append(v.subQueries, subQueryNode)
	}

	return v
}

func (v *subQueryCollector) VisitMany(nodes []ast.Node) ast.Visitor {
	return v
}

func (v *subQueryCollector) Field(name string) ast.Visitor {
	return v
}

func (v *subQueryCollector) Index(index int) ast.Visitor {
	return v
}