# examples/path-table 実行結果まとめ

このディレクトリには、memefishのASTにおけるPathTableExprとTableNameの違いと使い分けを理解するためのデモンストレーションが含まれています。

## ファイル構成

- `main.go`: メインのデモプログラム（PathTableExprとTableNameの分析）
- `comparison.md`: PathTableExprとTableNameの違いの詳細説明

## 実行結果

### 重要な発見

1. **TableName vs PathTableExpr の自動判別**
   - **TableName**: 単一の識別子（ドットなし）→ `*ast.TableName`
   - **PathTableExpr**: 複数の識別子（ドット区切り）→ `*ast.PathTableExpr`

2. **BigQuery/Spanner特有の階層構造サポート**
   - プロジェクト.データセット.テーブルの3階層パス
   - バッククォート付き識別子でハイフンも対応

3. **共通機能と独自機能**
   - 両方ともHint、As、Sampleをサポート
   - PathTableExprのみWithOffsetフィールドあり（ただしUNNESTでは別型）

### テストケースと結果

| SQL例 | AST型 | Path構造 | 特徴 |
|-------|-------|----------|------|
| `SELECT * FROM users` | `*ast.TableName` | 単一識別子 | 最もシンプルなテーブル参照 |
| `SELECT * FROM users AS u` | `*ast.TableName` | 単一識別子 + エイリアス | エイリアス付き単一テーブル |
| `SELECT * FROM my_schema.users` | `*ast.PathTableExpr` | 2階層 ["my_schema", "users"] | スキーマ.テーブル |
| `SELECT * FROM my_database.my_schema.users` | `*ast.PathTableExpr` | 3階層 ["my_database", "my_schema", "users"] | データベース.スキーマ.テーブル |
| `SELECT * FROM \`my-database\`.\`my-schema\`.\`my-table\`` | `*ast.PathTableExpr` | 3階層 ["my-database", "my-schema", "my-table"] | バッククォート付き識別子 |
| `SELECT * FROM project_id.dataset_id.table_id AS t` | `*ast.PathTableExpr` | 3階層 + エイリアス | BigQuery典型パターン |

### 詳細な実行結果

#### Example 1: 単純なテーブル名
```sql
SELECT * FROM users
```
- **AST型**: `*ast.TableName`
- **特徴**: 単一の識別子、最もシンプルな形

#### Example 2: スキーマ付きテーブル名
```sql
SELECT * FROM my_schema.users
```
- **AST型**: `*ast.PathTableExpr`
- **Path**: 2個の識別子 ["my_schema", "users"]
- **特徴**: ドット区切りでPathTableExprに分類

#### Example 3: フルパス（3階層）
```sql
SELECT * FROM my_database.my_schema.users
```
- **AST型**: `*ast.PathTableExpr`
- **Path**: 3個の識別子 ["my_database", "my_schema", "users"]
- **特徴**: BigQuery/Spannerでよく使われる形

#### Example 4: AS句付き単一テーブル
```sql
SELECT * FROM users AS u
```
- **AST型**: `*ast.TableName`
- **特徴**: エイリアス付きでも単一識別子はTableName

#### Example 5: バッククォート付き識別子
```sql
SELECT * FROM `my-database`.`my-schema`.`my-table`
```
- **AST型**: `*ast.PathTableExpr`
- **Path**: 3個の識別子 ["my-database", "my-schema", "my-table"]
- **特徴**: ハイフンを含む識別子もサポート

#### Example 6: JOIN構文
```sql
SELECT * FROM users u JOIN orders o ON u.id = o.user_id
```
- **AST型**: `*ast.Join`
- **特徴**: Join型でLeft/Rightテーブルがそれぞれ別々に処理

#### Example 7: UNNEST構文
```sql
SELECT * FROM UNNEST([1, 2, 3]) AS t WITH OFFSET AS pos
```
- **AST型**: `*ast.Unnest`
- **特徴**: UNNESTは専用の型、WithOffsetもサポート

#### Example 8: テーブルヒント
```sql
SELECT * FROM users@{FORCE_INDEX=idx_name}
```
- **AST型**: `*ast.TableName`
- **特徴**: ヒント付きでも単一識別子はTableName

#### Example 9: テーブルサンプリング（パースエラー）
```sql
SELECT * FROM users TABLESAMPLE SYSTEM (10 PERCENT)
```
- **結果**: パースエラー（SYSTEM構文未サポート）

### AST構造の比較

#### ast.TableName
```go
type TableName struct {
    Table  *Ident       // 単一の識別子
    Hint   *Hint        // テーブルヒント
    As     *AsAlias     // エイリアス
    Sample *TableSample // テーブルサンプル
}
```

#### ast.PathTableExpr
```go
type PathTableExpr struct {
    Path       *Path         // 識別子のリスト
    Hint       *Hint         // テーブルヒント
    As         *AsAlias      // エイリアス
    WithOffset *WithOffset   // OFFSET句（通常は未使用）
    Sample     *TableSample  // テーブルサンプル
}
```

### 実装のポイント

1. **パーサーの自動判別**: ドット(`.`)の有無で自動的に型を決定
2. **Path構造**: `[]*Ident`のスライスで複数階層を表現
3. **共通インターフェース**: 両型ともTableExprインターフェースを実装
4. **BigQuery対応**: プロジェクト.データセット.テーブルの3階層構造をサポート
5. **特殊文字対応**: バッククォートでハイフン等の特殊文字を含む識別子をサポート

### 実行方法

```bash
# メインデモの実行
go run .
```

この実装により、SQLの様々なテーブル参照パターンを適切に解析し、構造化された形で扱うことができます。特に、BigQueryやCloud Spannerでよく使われる階層的なテーブル名を正しく処理できるのが特徴です。