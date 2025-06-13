package main

import (
	"github.com/cloudspannerecosystem/memefish/ast"
)

// SelectWithOrderByLimitInfo はSelectとそのOrderBy/Limit情報を保持する構造体
type SelectWithOrderByLimitInfo struct {
	Select  *ast.Select
	OrderBy *ast.OrderBy // nilの場合はOrderByなし
	Limit   *ast.Limit   // nilの場合はLimitなし
	Context string       // "TopLevel", "SubQuery", "ScalarSubQuery"など
}

// HasSelectWithOrderByOrLimit は与えられたAST内のSelectでOrderByやLimitが使われているかを返します
func HasSelectWithOrderByOrLimit(node ast.Node) bool {
	infos := GetSelectsWithOrderByLimitInfo(node)
	for _, info := range infos {
		if info.OrderBy != nil || info.Limit != nil {
			return true
		}
	}
	return false
}

// GetSelectsWithOrderByLimitInfo は与えられたAST内の全てのSelectとそのOrderBy/Limit情報を返します
func GetSelectsWithOrderByLimitInfo(node ast.Node) []SelectWithOrderByLimitInfo {
	collector := &selectInfoCollector{}
	ast.Walk(node, collector)
	return collector.selects
}

// GetSelectsWithOrderByOrLimit はOrderByまたはLimitを持つSelectの情報のみを返します
func GetSelectsWithOrderByOrLimit(node ast.Node) []SelectWithOrderByLimitInfo {
	allSelects := GetSelectsWithOrderByLimitInfo(node)
	var result []SelectWithOrderByLimitInfo
	
	for _, info := range allSelects {
		if info.OrderBy != nil || info.Limit != nil {
			result = append(result, info)
		}
	}
	
	return result
}

// selectInfoCollector はSelectとその親のOrderBy/Limit情報を収集するvisitor
type selectInfoCollector struct {
	selects []SelectWithOrderByLimitInfo
}

func (v *selectInfoCollector) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.Query:
		// QueryがSelectを含む場合
		if select_, ok := n.Query.(*ast.Select); ok {
			info := SelectWithOrderByLimitInfo{
				Select:  select_,
				OrderBy: n.OrderBy,
				Limit:   n.Limit,
				Context: v.determineContext(n),
			}
			v.selects = append(v.selects, info)
		}
	}
	return v
}

func (v *selectInfoCollector) VisitMany(nodes []ast.Node) ast.Visitor {
	return v
}

func (v *selectInfoCollector) Field(name string) ast.Visitor {
	return v
}

func (v *selectInfoCollector) Index(index int) ast.Visitor {
	return v
}

// determineContext はQueryのコンテキストを判定します
func (v *selectInfoCollector) determineContext(query *ast.Query) string {
	// 簡単な実装：実際のコンテキストを正確に判定するにはより複雑な解析が必要
	return "Query"
}

// より具体的なコンテキスト判定を行うcollector
type detailedSelectInfoCollector struct {
	selects []SelectWithOrderByLimitInfo
	context string
}

func (v *detailedSelectInfoCollector) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.QueryStatement:
		// トップレベルのQuery
		childCollector := &detailedSelectInfoCollector{context: "TopLevel"}
		ast.Walk(n.Query, childCollector)
		v.selects = append(v.selects, childCollector.selects...)
		return nil // 子ノードの走査は既に完了
		
	case *ast.Query:
		if select_, ok := n.Query.(*ast.Select); ok {
			info := SelectWithOrderByLimitInfo{
				Select:  select_,
				OrderBy: n.OrderBy,
				Limit:   n.Limit,
				Context: v.context,
			}
			v.selects = append(v.selects, info)
		} else if compoundQuery, ok := n.Query.(*ast.CompoundQuery); ok {
			// CompoundQueryの場合、各Queryを個別に処理
			for _, query := range compoundQuery.Queries {
				childCollector := &detailedSelectInfoCollector{context: v.context + "_CompoundQuery"}
				ast.Walk(query, childCollector)
				v.selects = append(v.selects, childCollector.selects...)
			}
			// CompoundQuery全体のOrderBy/Limitは最後に追加（該当するSelectがある場合）
			if (n.OrderBy != nil || n.Limit != nil) && len(compoundQuery.Queries) > 0 {
				// 最後のQueryがSelectの場合、そのSelectにOrderBy/Limitを適用
				if lastQuery, ok := compoundQuery.Queries[len(compoundQuery.Queries)-1].(*ast.Select); ok {
					info := SelectWithOrderByLimitInfo{
						Select:  lastQuery,
						OrderBy: n.OrderBy,
						Limit:   n.Limit,
						Context: v.context + "_CompoundQueryOrderBy",
					}
					v.selects = append(v.selects, info)
				}
			}
		}
		
	case *ast.Select:
		// 直接Selectが来た場合（Query内でない）
		info := SelectWithOrderByLimitInfo{
			Select:  n,
			OrderBy: nil, // SelectはOrderBy/Limitを直接持たない
			Limit:   nil,
			Context: v.context + "_DirectSelect",
		}
		v.selects = append(v.selects, info)
		
	case *ast.SubQuery:
		childCollector := &detailedSelectInfoCollector{context: "SubQuery"}
		ast.Walk(n.Query, childCollector)
		v.selects = append(v.selects, childCollector.selects...)
		return nil
		
	case *ast.SubQueryTableExpr:
		childCollector := &detailedSelectInfoCollector{context: "SubQueryTableExpr"}
		ast.Walk(n.Query, childCollector)
		v.selects = append(v.selects, childCollector.selects...)
		return nil
		
	case *ast.ScalarSubQuery:
		childCollector := &detailedSelectInfoCollector{context: "ScalarSubQuery"}
		ast.Walk(n.Query, childCollector)
		v.selects = append(v.selects, childCollector.selects...)
		return nil
		
	case *ast.ArraySubQuery:
		childCollector := &detailedSelectInfoCollector{context: "ArraySubQuery"}
		ast.Walk(n.Query, childCollector)
		v.selects = append(v.selects, childCollector.selects...)
		return nil
		
	case *ast.ExistsSubQuery:
		childCollector := &detailedSelectInfoCollector{context: "ExistsSubQuery"}
		ast.Walk(n.Query, childCollector)
		v.selects = append(v.selects, childCollector.selects...)
		return nil
	}
	return v
}

func (v *detailedSelectInfoCollector) VisitMany(nodes []ast.Node) ast.Visitor {
	return v
}

func (v *detailedSelectInfoCollector) Field(name string) ast.Visitor {
	return v
}

func (v *detailedSelectInfoCollector) Index(index int) ast.Visitor {
	return v
}

// GetDetailedSelectsWithOrderByLimitInfo はより詳細なコンテキスト情報を含むSelectの情報を返します
func GetDetailedSelectsWithOrderByLimitInfo(node ast.Node) []SelectWithOrderByLimitInfo {
	collector := &detailedSelectInfoCollector{context: "Unknown"}
	ast.Walk(node, collector)
	return collector.selects
}