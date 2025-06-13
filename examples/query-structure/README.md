# examples/query-structure 実行結果まとめ

このディレクトリには、memefishのASTにおけるクエリ構造（QueryExprの階層）の理解を深めるためのデモンストレーションが含まれています。

## ファイル構成

- `main.go`: メインのデモプログラム（QueryExprの階層構造分析）
- `summary.md`: ast.Queryの構造についての詳細説明

## 実行結果

### 重要な発見

1. **ルートレベルでは常に `ast.QueryStatement` が来る**
   - 全てのSQLクエリは `ast.QueryStatement` でラップされる
   - `QueryStatement.Query` は `QueryExpr` インターフェース型

2. **`ast.Query` は必須ではない装飾的ラッパー**
   - OrderBy/Limit/WITH/PipeOperatorsが必要な場合にのみ使用される
   - 単純なクエリでは `ast.Select` や `ast.CompoundQuery` が直接使用される

3. **階層構造の動的な決定**
   - パーサーがSQL構文に応じて適切な型を自動選択
   - OrderBy/Limit等の存在により構造が変化

### テストケースと結果

| SQL例 | QueryStatement.Query型 | 特徴 |
|-------|----------------------|------|
| `SELECT name FROM users` | `*ast.Select` | 単純なSelect（装飾なし） |
| `SELECT name FROM users ORDER BY name` | `*ast.Query` | OrderByありでQueryラッパー |
| `SELECT * FROM (SELECT name FROM users) AS u` | `*ast.Select` | SubQuery内もOrderByなしなら直接Select |
| `SELECT * FROM (SELECT name FROM users ORDER BY name) AS u` | `*ast.Select` | SubQuery内のOrderByは内部Query構造で処理 |
| `SELECT ... UNION ...` | `*ast.CompoundQuery` | 複合クエリ（装飾なし） |
| `SELECT ... UNION ... ORDER BY ...` | `*ast.Query` | 複合クエリ + OrderByでQueryラッパー |
| `WITH ... SELECT ...` | `*ast.Query` | WITH句ありでQueryラッパー |
| 複雑なネスト | `*ast.Select` | 多層ネストでも各レベルで適切に処理 |

### 詳細な実行結果

#### Example 1: 単純なSelect
```sql
SELECT name FROM users
```
- **構造**: `QueryStatement → Select`
- **特徴**: 最もシンプルな形、装飾的ラッパーなし

#### Example 2: OrderBy付きSelect
```sql
SELECT name FROM users ORDER BY name
```
- **構造**: `QueryStatement → Query → Select`
- **特徴**: OrderByのためQueryラッパーが追加される

#### Example 3: 単純なSubQuery
```sql
SELECT * FROM (SELECT name FROM users) AS u
```
- **構造**: `QueryStatement → Select → From → SubQueryTableExpr → Select`
- **特徴**: SubQuery内にOrderByがないため直接Selectが使用される

#### Example 4: OrderBy付きSubQuery
```sql
SELECT * FROM (SELECT name FROM users ORDER BY name) AS u
```
- **構造**: `QueryStatement → Select → From → SubQueryTableExpr → Query → Select`
- **特徴**: SubQuery内のOrderByのためQuery構造が挿入される

#### Example 5: 単純なUNION
```sql
SELECT name FROM users UNION ALL SELECT name FROM customers
```
- **構造**: `QueryStatement → CompoundQuery`
- **特徴**: 複合クエリでも装飾がなければ直接CompoundQuery

#### Example 6: OrderBy付きUNION
```sql
SELECT name FROM users UNION ALL SELECT name FROM customers ORDER BY name
```
- **構造**: `QueryStatement → Query → CompoundQuery`
- **特徴**: 全体のOrderByのためQueryラッパーが追加される

#### Example 7: WITH句
```sql
WITH user_names AS (SELECT name FROM users) SELECT * FROM user_names
```
- **構造**: `QueryStatement → Query → Select`
- **特徴**: WITH句のためQueryラッパーが必要

#### Example 8: 複雑なネスト
```sql
SELECT * FROM (SELECT name FROM (SELECT * FROM users ORDER BY id) ORDER BY name LIMIT 5) AS nested
```
- **構造**: 多層のネストした構造
- **特徴**: 各レベルでOrderBy/Limitに応じてQuery構造が適切に挿入される

### AST構造の詳細

#### ast.QueryStatement（ルート）
```go
type QueryStatement struct {
    Hint  *Hint     // クエリヒント（オプション）
    Query QueryExpr  // 実際のクエリ本体
}
```

#### ast.Query（装飾的ラッパー）
```go
type Query struct {
    With          *With          // WITH句
    Query         QueryExpr      // 実際のクエリ本体
    OrderBy       *OrderBy       // ORDER BY句
    Limit         *Limit         // LIMIT句
    ForUpdate     *ForUpdate     // FOR UPDATE句
    PipeOperators []PipeOperator // パイプ演算子
}
```

#### QueryExprインターフェースの実装型
- `*ast.Query`: 装飾的機能を持つクエリラッパー
- `*ast.Select`: 基本的なSELECT文
- `*ast.CompoundQuery`: UNION/INTERSECT/EXCEPT
- `*ast.SubQuery`: 括弧付きクエリ
- `*ast.FromQuery`: FROM句のクエリ

### 構造決定のルール

1. **基本原則**: 必要最小限の構造を使用
2. **装飾的機能の追加**: OrderBy/Limit/WITH/PipeOperatorsがあるときのみQueryラッパーを使用
3. **再帰的適用**: SubQuery内でも同じルールを適用
4. **型の自動選択**: パーサーがSQL構文に応じて最適な型を選択

### 実装のポイント

1. **動的構造**: SQL構文に応じて柔軟に構造が変化
2. **効率的設計**: 不要なラッパーを避けるミニマルな設計
3. **再帰的処理**: ネストしたクエリも適切に処理
4. **型安全性**: インターフェースによる型安全な設計

### 実行方法

```bash
# メインデモの実行
go run .
```

この実装により、SQLクエリの構造を効率的かつ柔軟に表現できます。特に、必要な場合にのみ装飾的機能を追加する設計により、シンプルなクエリでは軽量な構造を保ちながら、複雑なクエリでも適切に対応できるのが特徴です。