# examples/check-select 実行結果まとめ

このディレクトリには、memefishのASTを使ってSQL文内のSELECT文のOrderBy/Limit節をチェックする機能のデモンストレーションが含まれています。

## ファイル構成

- `main.go`: メインのデモプログラム（詳細なAST解析）
- `summary.go`: シンプルなデモンストレーション
- `utils.go`: 詳細なSelect情報を取得するユーティリティ関数群
- `simple_utils.go`: シンプルなSelect情報を取得するユーティリティ関数群

## 実行結果

### 重要な発見

1. **ast.Select自体にはOrderBy/Limitフィールドがない**
   - OrderBy/Limitは親の`ast.Query`によって管理される
   - `ast.Select`を直接調べてもOrderBy/Limitの情報は得られない

2. **複数のコンテキストでSelectが使用される**
   - TopLevel: メインクエリのSelect
   - SubQuery: サブクエリ内のSelect
   - ScalarSubQuery: スカラーサブクエリ内のSelect
   - SubQueryTableExpr: FROM句のサブクエリ内のSelect
   - CompoundQuery: UNION等の複合クエリ内のSelect

### テストケースと結果

| 例 | SQL | OrderBy/Limit検出 | 詳細 |
|----|-----|------------------|------|
| 1 | `SELECT name FROM users` | ❌ | OrderBy/Limitなし |
| 2 | `SELECT name FROM users ORDER BY name` | ✅ | TopLevelでOrderBy検出 |
| 3 | `SELECT name FROM users LIMIT 10` | ✅ | TopLevelでLimit検出 |
| 4 | `SELECT name FROM users ORDER BY name LIMIT 10` | ✅ | TopLevelでOrderBy+Limit検出 |
| 5 | `SELECT * FROM (SELECT name FROM users ORDER BY name) AS u` | ✅ | SubQueryTableExpr内でOrderBy検出 |
| 6 | `SELECT (SELECT COUNT(*) FROM users ORDER BY id LIMIT 1) AS count` | ✅ | ScalarSubQuery内でOrderBy+Limit検出 |
| 7 | `SELECT name FROM users UNION ALL SELECT name FROM customers ORDER BY name` | ✅ | CompoundQuery全体でOrderBy検出 |
| 8 | 複雑なネスト構造 | ✅ | 複数レベルでOrderBy/Limit検出 |

### 提供されるユーティリティ関数

#### 詳細版（utils.go）
- `HasSelectWithOrderByOrLimit(node ast.Node) bool`: OrderBy/Limitを持つSelectがあるかチェック
- `GetSelectsWithOrderByLimitInfo(node ast.Node) []SelectWithOrderByLimitInfo`: 全Select情報を取得
- `GetSelectsWithOrderByOrLimit(node ast.Node) []SelectWithOrderByLimitInfo`: OrderBy/Limitを持つSelectのみ取得

#### シンプル版（simple_utils.go）
- `HasAnySelectWithOrderByOrLimit(node ast.Node) bool`: OrderBy/Limitを持つSelectがあるかチェック
- `FindSelectsWithOrderByOrLimit(node ast.Node) []SimpleSelectInfo`: OrderBy/Limitを持つSelectを検索
- `FindAllSelects(node ast.Node) []SimpleSelectInfo`: 全Selectを検索

### 実装のポイント

1. **AST走査**: `ast.Walk`を使用してAST全体を走査
2. **コンテキスト追跡**: Selectがどのコンテキスト（TopLevel, SubQuery等）にあるかを記録
3. **Query-Select関係**: `ast.Query`と内部の`ast.Select`の関係を正しく処理
4. **CompoundQuery対応**: UNION等の複合クエリも適切に処理

### 実行方法

```bash
# メインデモの実行
go run main.go utils.go simple_utils.go

# シンプルデモの実行
go run summary.go simple_utils.go
```

この実装により、複雑なSQL文中のSELECT文でOrderBy/Limitが使用されているかを確実に検出できます。