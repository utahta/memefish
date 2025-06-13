# ast.Query の構造について

## 階層構造の理解

### 基本的な階層
```
ast.QueryStatement (トップレベル)
├── Hint (optional)
└── Query (ast.QueryExpr interface)
    ├── *ast.Query ← OrderBy/Limit/WITH/PipeOperatorsを持つラッパー
    ├── *ast.Select ← 基本的なSELECT文
    ├── *ast.CompoundQuery ← UNION/INTERSECT/EXCEPT
    ├── *ast.SubQuery ← 括弧付きクエリ
    └── *ast.FromQuery ← FROM句のクエリ
```

### 重要な発見

#### 1. **ルートレベルでは常に `ast.QueryStatement` が来ます**
- 全てのSQLクエリは `ast.QueryStatement` でラップされます
- `QueryStatement.Query` は `QueryExpr` インターフェース型です

#### 2. **`ast.Query` は必ずしもルートに来るわけではありません**
以下のパターンがあります：

| SQL例 | QueryStatement.Query型 | 説明 |
|-------|---------------------|------|
| `SELECT name FROM users` | `*ast.Select` | 単純なSelect（OrderBy/Limitなし） |
| `SELECT name FROM users ORDER BY name` | `*ast.Query` | OrderBy/Limitがある場合はQueryでラップ |
| `SELECT ... UNION ...` | `*ast.CompoundQuery` | 複合クエリ |
| `SELECT ... UNION ... ORDER BY ...` | `*ast.Query` | 複合クエリにOrderBy/Limitがある場合 |
| `WITH ... SELECT ...` | `*ast.Query` | WITH句がある場合 |

#### 3. **OrderBy/Limit/WITH/PipeOperatorsは `ast.Query` でのみ管理されます**
```go
type Query struct {
    With          *With
    Query         QueryExpr      // 実際のクエリ本体
    OrderBy       *OrderBy       // optional
    Limit         *Limit         // optional
    ForUpdate     *ForUpdate     // optional
    PipeOperators []PipeOperator
}
```

#### 4. **`ast.Query` は再帰的にネストできます**
- `Query.Query` が別の `*ast.Query` の場合もあります
- SubQuery内でもOrderBy/Limitがある場合は `ast.Query` でラップされます

### 実際の例

#### 例1: 単純なSelect
```sql
SELECT name FROM users
```
```
QueryStatement
└── Select (直接)
```

#### 例2: OrderBy付きSelect
```sql
SELECT name FROM users ORDER BY name
```
```
QueryStatement
└── Query
    ├── OrderBy: "ORDER BY name"
    └── Query: Select
```

#### 例3: SubQuery内でOrderBy
```sql
SELECT * FROM (SELECT name FROM users ORDER BY name) AS u
```
```
QueryStatement
└── Select
    └── From
        └── SubQueryTableExpr
            └── Query (SubQuery内のOrderByのため)
                ├── OrderBy: "ORDER BY name"
                └── Query: Select
```

#### 例4: UNION with OrderBy
```sql
SELECT name FROM users UNION ALL SELECT name FROM customers ORDER BY name
```
```
QueryStatement
└── Query (全体のOrderByのため)
    ├── OrderBy: "ORDER BY name"
    └── Query: CompoundQuery
        ├── Op: "UNION"
        ├── AllOrDistinct: "ALL"
        └── Queries: [Select, Select]
```

### 結論

**あなたの理解は部分的に正しいですが、完全ではありません：**

✅ **正しい部分:**
- OrderBy/Limit/WITHは `ast.Query` で管理される
- `ast.Query` は重要なラッパー役割を果たす

❌ **修正が必要な部分:**
- ルートレベルで常に `ast.Query` が来るわけではない
- 単純なSelectやCompoundQueryは直接 `QueryStatement.Query` に来ることがある
- `ast.Query` は「必要な場合にのみ」追加される装飾的なラッパー

### 正しい理解
**`ast.Query` は "OrderBy/Limit/WITH/PipeOperatorsが必要な場合にのみ使用される装飾的なラッパー" です。**

単純なクエリでは `ast.Select` や `ast.CompoundQuery` が直接使用され、OrderBy/Limit等の機能が必要になったときに `ast.Query` でラップされる設計になっています。