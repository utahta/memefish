# PathTableExpr vs TableName の違い

## 基本的な使い分け

### TableName
- **単一の識別子** (テーブル名のみ)
- `users`, `orders` など

### PathTableExpr  
- **複数の識別子からなるパス** (ドット区切り)
- `schema.table`, `project.dataset.table` など

## 実際の例

| SQL | AST型 | 説明 |
|-----|-------|------|
| `SELECT * FROM users` | `*ast.TableName` | 単一テーブル名 |
| `SELECT * FROM users AS u` | `*ast.TableName` | 単一テーブル名 + エイリアス |
| `SELECT * FROM my_schema.users` | `*ast.PathTableExpr` | スキーマ.テーブル |
| `SELECT * FROM project.dataset.table` | `*ast.PathTableExpr` | プロジェクト.データセット.テーブル |
| `SELECT * FROM \`my-database\`.\`my-table\`` | `*ast.PathTableExpr` | バッククォート付きパス |

## 構造の比較

### ast.TableName
```go
type TableName struct {
    Table  *Ident       // 単一の識別子
    Hint   *Hint        
    As     *AsAlias     
    Sample *TableSample 
}
```

### ast.PathTableExpr
```go
type PathTableExpr struct {
    Path       *Path         // 識別子のスライス ([]*Ident)
    Hint       *Hint         
    As         *AsAlias      
    WithOffset *WithOffset   // UNNESTでは使用されないが構造上存在
    Sample     *TableSample  
}
```

### ast.Path
```go
type Path struct {
    Idents []*Ident      // ["project", "dataset", "table"]
}
```

## 重要なポイント

1. **判別基準**: ドット(`.`)の有無
   - ドットなし → `TableName`
   - ドットあり → `PathTableExpr`

2. **PathTableExprのPath**: 
   - 最低1個の識別子を含む
   - 通常は2個以上（ドット区切りの場合）

3. **共通フィールド**:
   - 両方とも `Hint`, `As`, `Sample` をサポート
   - `PathTableExpr` のみ `WithOffset` あり（ただしUNNESTでは使われない）

4. **BigQuery/Spanner特有**:
   - プロジェクト.データセット.テーブル の3階層パス
   - バッククォートでハイフン含み識別子をサポート

## パース時の決定

パーサーは以下の流れで決定：
1. 識別子を読む
2. 次のトークンが `.` かチェック
3. `.` があれば続きの識別子も読んでPathTableExpr作成
4. `.` がなければTableName作成

これにより、適切な型が自動選択される。