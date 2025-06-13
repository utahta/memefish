package main

import (
	"github.com/cloudspannerecosystem/memefish/ast"
)

// SimpleSelectInfo はSelectとその親のQueryによるOrderBy/Limit情報を保持する
type SimpleSelectInfo struct {
	Select       *ast.Select
	HasOrderBy   bool
	HasLimit     bool
	OrderBySQL   string
	LimitSQL     string
	ContextType  string // "TopLevel", "SubQuery", "ScalarSubQuery", etc.
}

// FindSelectsWithOrderByOrLimit はOrderByまたはLimitが適用されているSelectを見つけます
func FindSelectsWithOrderByOrLimit(node ast.Node) []SimpleSelectInfo {
	finder := &simpleSelectFinder{}
	ast.Walk(node, finder)
	
	var result []SimpleSelectInfo
	for _, info := range finder.selects {
		if info.HasOrderBy || info.HasLimit {
			result = append(result, info)
		}
	}
	return result
}

// FindAllSelects はAST内の全てのSelectとその状態を返します
func FindAllSelects(node ast.Node) []SimpleSelectInfo {
	finder := &simpleSelectFinder{}
	ast.Walk(node, finder)
	return finder.selects
}

// HasAnySelectWithOrderByOrLimit はOrderByまたはLimitが適用されているSelectがあるかを返します
func HasAnySelectWithOrderByOrLimit(node ast.Node) bool {
	selects := FindSelectsWithOrderByOrLimit(node)
	return len(selects) > 0
}

type simpleSelectFinder struct {
	selects []SimpleSelectInfo
}

func (f *simpleSelectFinder) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.QueryStatement:
		f.processQuery(n.Query, "TopLevel")
	case *ast.SubQueryTableExpr:
		f.processQuery(n.Query, "SubQueryTableExpr")
	case *ast.ScalarSubQuery:
		f.processQuery(n.Query, "ScalarSubQuery")
	case *ast.ArraySubQuery:
		f.processQuery(n.Query, "ArraySubQuery")
	case *ast.ExistsSubQuery:
		f.processQuery(n.Query, "ExistsSubQuery")
	case *ast.SubQuery:
		f.processQuery(n.Query, "SubQuery")
	}
	return f
}

func (f *simpleSelectFinder) VisitMany(nodes []ast.Node) ast.Visitor {
	return f
}

func (f *simpleSelectFinder) Field(name string) ast.Visitor {
	return f
}

func (f *simpleSelectFinder) Index(index int) ast.Visitor {
	return f
}

func (f *simpleSelectFinder) processQuery(queryExpr ast.QueryExpr, contextType string) {
	switch q := queryExpr.(type) {
	case *ast.Query:
		if selectStmt, ok := q.Query.(*ast.Select); ok {
			info := SimpleSelectInfo{
				Select:      selectStmt,
				HasOrderBy:  q.OrderBy != nil,
				HasLimit:    q.Limit != nil,
				ContextType: contextType,
			}
			
			if q.OrderBy != nil {
				info.OrderBySQL = q.OrderBy.SQL()
			}
			if q.Limit != nil {
				info.LimitSQL = q.Limit.SQL()
			}
			
			f.selects = append(f.selects, info)
		} else if compoundQuery, ok := q.Query.(*ast.CompoundQuery); ok {
			// CompoundQueryの各クエリを処理
			for _, query := range compoundQuery.Queries {
				f.processQuery(query, contextType+"_CompoundQuery")
			}
			
			// CompoundQuery全体のOrderBy/Limitがある場合
			if q.OrderBy != nil || q.Limit != nil {
				// 最後のSelectに適用されると考える
				if len(compoundQuery.Queries) > 0 {
					if lastSelect, ok := compoundQuery.Queries[len(compoundQuery.Queries)-1].(*ast.Select); ok {
						info := SimpleSelectInfo{
							Select:      lastSelect,
							HasOrderBy:  q.OrderBy != nil,
							HasLimit:    q.Limit != nil,
							ContextType: contextType + "_CompoundQueryFinal",
						}
						
						if q.OrderBy != nil {
							info.OrderBySQL = q.OrderBy.SQL()
						}
						if q.Limit != nil {
							info.LimitSQL = q.Limit.SQL()
						}
						
						f.selects = append(f.selects, info)
					}
				}
			}
		}
	case *ast.Select:
		// 直接Selectの場合（OrderBy/Limitは親のQueryで管理される）
		info := SimpleSelectInfo{
			Select:      q,
			HasOrderBy:  false,
			HasLimit:    false,
			ContextType: contextType + "_DirectSelect",
		}
		f.selects = append(f.selects, info)
	case *ast.SubQuery:
		f.processQuery(q.Query, contextType+"_NestedSubQuery")
	case *ast.CompoundQuery:
		for _, query := range q.Queries {
			f.processQuery(query, contextType+"_CompoundQuery")
		}
	}
}