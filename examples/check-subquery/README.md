# examples/check-subquery 実行結果まとめ

このディレクトリには、memefishのASTを使ってSQL文内のサブクエリ（SubQuery）でOrderBy/Limit節が使用されているかをチェックする機能のデモンストレーションが含まれています。

## ファイル構成

- `main.go`: メインのデモプログラム（詳細なサブクエリ解析）
- `utils.go`: サブクエリ内のOrderBy/Limitをチェックするユーティリティ関数群

## 実行結果

### 重要な発見

1. **多様なSubQueryタイプをサポート**
   - `SubQueryTableExpr`: FROM句のサブクエリ `SELECT * FROM (subquery) AS alias`
   - `ScalarSubQuery`: スカラーサブクエリ `SELECT (subquery) AS value`
   - `ExistsSubQuery`: EXISTS句のサブクエリ `WHERE EXISTS (subquery)`
   - `ArraySubQuery`: ARRAY関数のサブクエリ `SELECT ARRAY(subquery)`
   - `SubQuery`: 一般的なサブクエリ

2. **ネストしたサブクエリも適切に処理**
   - 複数レベルのネストしたサブクエリを再帰的に解析
   - 各レベルでOrderBy/Limitを個別に検出

3. **QueryExprの階層構造を理解**
   - `ast.Query`がOrderBy/Limitフィールドを保持
   - サブクエリ内のQueryExprを正しく辿る必要がある

### テストケースと結果

| 例 | SQL概要 | OrderBy/Limit検出 | SubQuery種類 | 詳細 |
|----|---------|------------------|--------------|------|
| 1 | FROM句サブクエリ（ORDER BY） | ✅ | SubQueryTableExpr | ORDER BY name検出 |
| 2 | FROM句サブクエリ（LIMIT） | ✅ | SubQueryTableExpr | LIMIT 10検出 |
| 3 | FROM句サブクエリ（ORDER BY + LIMIT） | ✅ | SubQueryTableExpr | ORDER BY + LIMIT検出 |
| 4 | FROM句サブクエリ（なし） | ❌ | SubQueryTableExpr | OrderBy/Limitなし |
| 5 | スカラーサブクエリ | ✅ | ScalarSubQuery | ORDER BY + LIMIT検出 |
| 6 | EXISTSサブクエリ | ✅ | ExistsSubQuery | ORDER BY created_at検出 |
| 7 | ARRAYサブクエリ | ✅ | ArraySubQuery | ORDER BY name検出 |
| 8 | ネストしたサブクエリ | ✅ | SubQueryTableExpr (×2) | 複数レベルでOrderBy/Limit検出 |

### 詳細な実行結果

#### Example 1: FROM句サブクエリ（ORDER BY）
```sql
SELECT * FROM (SELECT name FROM users ORDER BY name) AS u
```
- **検出**: ✅ SubQueryTableExpr found with OrderBy

#### Example 2: FROM句サブクエリ（LIMIT）
```sql
SELECT * FROM (SELECT name FROM users LIMIT 10) AS u
```
- **検出**: ✅ SubQueryTableExpr found with Limit

#### Example 3: FROM句サブクエリ（ORDER BY + LIMIT）
```sql
SELECT * FROM (SELECT name FROM users ORDER BY name LIMIT 10) AS u
```
- **検出**: ✅ SubQueryTableExpr found with OrderBy and Limit

#### Example 4: FROM句サブクエリ（なし）
```sql
SELECT * FROM (SELECT name FROM users) AS u
```
- **検出**: ❌ No OrderBy or Limit

#### Example 5: スカラーサブクエリ
```sql
SELECT (SELECT COUNT(*) FROM users ORDER BY id LIMIT 1) AS count
```
- **検出**: ✅ ScalarSubQuery found with OrderBy and Limit

#### Example 6: EXISTSサブクエリ
```sql
SELECT * FROM products WHERE EXISTS (SELECT 1 FROM orders WHERE orders.product_id = products.id ORDER BY created_at)
```
- **検出**: ✅ ExistsSubQuery found with OrderBy

#### Example 7: ARRAYサブクエリ
```sql
SELECT ARRAY(SELECT name FROM users ORDER BY name) AS names
```
- **検出**: ✅ ArraySubQuery found with OrderBy

#### Example 8: ネストしたサブクエリ
```sql
SELECT * FROM (SELECT * FROM (SELECT id FROM users ORDER BY id) ORDER BY id LIMIT 5) AS nested
```
- **検出**: ✅ 2つのSubQueryTableExprで両方ともOrderBy/Limit検出

### 提供されるユーティリティ関数

#### 基本関数（utils.go）
- `HasSubQueryWithOrderByOrLimit(node ast.Node) bool`: OrderBy/Limitを持つサブクエリがあるかチェック
- `GetSubQueriesWithOrderByOrLimit(node ast.Node) []ast.Node`: OrderBy/Limitを持つサブクエリを全て取得

#### 内部関数
- `checkQueryExprForOrderByLimit(queryExpr ast.QueryExpr) bool`: QueryExpr内のOrderBy/Limitをチェック
- `subQueryChecker`: OrderBy/Limitを持つサブクエリの存在を確認するvisitor
- `subQueryCollector`: OrderBy/Limitを持つサブクエリを収集するvisitor

### 実装のポイント

1. **AST走査**: `ast.Walk`を使用してAST全体を走査
2. **SubQueryタイプ判定**: 各SubQueryタイプ（SubQueryTableExpr、ScalarSubQuery等）を適切に識別
3. **QueryExpr解析**: サブクエリ内のQueryExprを再帰的に解析してOrderBy/Limitを検出
4. **効率的な探索**: 早期終了による効率的な探索（HasSubQueryWithOrderByOrLimit）
5. **ネスト対応**: 複数レベルのネストしたサブクエリを適切に処理

### 実行方法

```bash
# メインデモの実行
go run .
```

この実装により、複雑なSQL文中の様々なタイプのサブクエリでOrderBy/Limitが使用されているかを確実に検出できます。特に、ネストしたサブクエリや異なるコンテキストでのサブクエリも適切に処理できるのが特徴です。