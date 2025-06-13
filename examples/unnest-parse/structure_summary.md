# UNNEST クエリのパース結果構造

## クエリ
```sql
SELECT * FROM UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')])
```

## AST構造の詳細

### 全体の階層構造
```
ast.QueryStatement
└── ast.Select
    ├── Results: [ast.Star]  (SELECT *)
    └── From: ast.From
        └── Source: ast.Unnest
            └── Expr: ast.ArrayLiteral
                ├── Type: ast.StructType
                │   └── Fields: [x INT64, y STRING]
                └── Values: [TupleStructLiteral, TupleStructLiteral]
                    ├── (1, 'foo')
                    └── (3, 'bar')
```

### 各ノードの詳細

#### 1. ast.Unnest
```go
type Unnest struct {
    Unnest     token.Pos      // "UNNEST"の位置
    Rparen     token.Pos      // ")"の位置 
    Expr       Expr           // ArrayLiteral
    Hint       *Hint          // nil
    As         *AsAlias       // nil (AS句なし)
    WithOffset *WithOffset    // nil (WITH OFFSET句なし)
    Sample     *TableSample   // nil
}
```

#### 2. ast.ArrayLiteral  
```go
type ArrayLiteral struct {
    Array     token.Pos    // "ARRAY"キーワードの位置
    Lbrack    token.Pos    // "["の位置
    Rbrack    token.Pos    // "]"の位置
    Type      Type         // ast.StructType<x INT64, y STRING>
    Values    []Expr       // [TupleStructLiteral, TupleStructLiteral]
}
```

#### 3. ast.StructType
```go
type StructType struct {
    Struct   token.Pos        // "STRUCT"の位置
    Gt       token.Pos        // ">"の位置  
    Fields   []*StructField   // [x INT64, y STRING]
}

// 各フィールド:
StructField {
    Ident: &Ident{Name: "x"}
    Type:  &SimpleType{Name: "INT64"}
}
StructField {
    Ident: &Ident{Name: "y"}  
    Type:  &SimpleType{Name: "STRING"}
}
```

#### 4. ast.TupleStructLiteral (2個)
```go
// 1個目: (1, 'foo')
TupleStructLiteral {
    Lparen: token.Pos        // "("の位置
    Rparen: token.Pos        // ")"の位置
    Values: []Expr {
        &IntLiteral{Value: "1"},
        &StringLiteral{Value: "foo"}
    }
}

// 2個目: (3, 'bar')  
TupleStructLiteral {
    Lparen: token.Pos
    Rparen: token.Pos
    Values: []Expr {
        &IntLiteral{Value: "3"},
        &StringLiteral{Value: "bar"}
    }
}
```

## 重要なポイント

### 1. **ArrayLiteralの型指定**
- `ARRAY<STRUCT<x INT64, y STRING>>` の部分は `ArrayLiteral.Type` として `ast.StructType` で表現
- 型指定は配列の要素型を定義

### 2. **StructLiteralの種類**
- `(1, 'foo')` のようなタプル形式は `ast.TupleStructLiteral`
- 他にも `ast.TypedStructLiteral`, `ast.TypelessStructLiteral` が存在

### 3. **UNNESTの表現**
- `UNNEST()` は `ast.Unnest` として `TableExpr` インターフェースを実装
- `FROM` 句の `Source` として配置される

### 4. **位置情報の保持**
- 全てのトークンの位置情報が保持される（エラー報告や IDE連携に有用）

### 5. **型安全性**
- STRUCT の フィールド定義(`x INT64, y STRING`)と実際の値(`(1, 'foo')`)は別々に管理
- パース時点では型の一致性は検証されない（セマンティック解析が必要）

## SQL生成
各ノードは `SQL()` メソッドを持ち、元のSQLを再構築可能：
```go
stmt.SQL() // "SELECT * FROM UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')])"
```