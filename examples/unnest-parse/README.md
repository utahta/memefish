# examples/unnest-parse 実行結果まとめ

このディレクトリには、memefishのASTにおけるUNNEST構文の詳細なパース結果を理解するためのデモンストレーションが含まれています。

## ファイル構成

- `main.go`: メインのデモプログラム（詳細なUNNEST構造分析）
- `variations.go`: 様々なUNNESTパターンのクイック分析
- `structure_summary.md`: UNNEST構造の詳細説明

## 実行結果

### 重要な発見

1. **UNNEST構文の基本構造**
   - `SELECT * FROM UNNEST(expr)` → `QueryStatement → Select → From → Unnest`
   - UNNESTは `ast.Unnest` として `TableExpr` インターフェースを実装

2. **ArrayLiteralの複雑な型システム**
   - 型指定付きARRAY: `ARRAY<STRUCT<...>>[...]`
   - ArrayLiteralは `Type` と `Values` を分離して管理
   - StructTypeは詳細なフィールド定義を保持

3. **StructLiteralの種類**
   - `TupleStructLiteral`: `(val1, val2, ...)` 形式
   - `TypedStructLiteral`: `STRUCT<Type>(val1, val2, ...)` 形式
   - `TypelessStructLiteral`: `STRUCT(name1 val1, name2 val2, ...)` 形式

### テストケースと結果

| SQL例 | ArrayLiteral.Type | Values.Type | As | WithOffset | 特徴 |
|-------|------------------|-------------|----|-----------| -----|
| `UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')])` | `*ast.StructType` | `*ast.TupleStructLiteral` | ❌ | ❌ | 型指定STRUCT配列 |
| `UNNEST(...) AS t` | `*ast.StructType` | `*ast.TupleStructLiteral` | ✅ | ❌ | テーブルエイリアス付き |
| `UNNEST(...) WITH OFFSET` | `*ast.StructType` | `*ast.TupleStructLiteral` | ❌ | ✅ | 行番号カラム付き |
| `UNNEST(...) AS t WITH OFFSET AS pos` | `*ast.StructType` | `*ast.TupleStructLiteral` | ✅ | ✅ | 両方のオプション |
| `UNNEST([1, 2, 3])` | `nil` | `*ast.IntLiteral` | ❌ | ❌ | 型推論による単純配列 |
| `UNNEST(ARRAY<INT64>[1, 2, 3])` | `*ast.SimpleType` | `*ast.IntLiteral` | ❌ | ❌ | 型指定単純配列 |
| `UNNEST(['hello', 'world'])` | `nil` | `*ast.StringLiteral` | ❌ | ❌ | 文字列配列 |
| ネストしたSTRUCT | `*ast.StructType` | `*ast.TupleStructLiteral` | ❌ | ❌ | 複雑なネスト構造 |

### 詳細な実行結果

#### メインケース: 複雑なSTRUCT配列
```sql
SELECT * FROM UNNEST(ARRAY<STRUCT<x INT64, y STRING>>[(1, 'foo'), (3, 'bar')])
```

**AST構造:**
```
QueryStatement
└── Select
    ├── Results: [Star] (SELECT *)
    └── From: From
        └── Source: Unnest
            └── Expr: ArrayLiteral
                ├── Type: StructType
                │   └── Fields: [x INT64, y STRING]
                └── Values: [TupleStructLiteral, TupleStructLiteral]
                    ├── (1, 'foo')
                    └── (3, 'bar')
```

**各コンポーネントの詳細:**

1. **ast.Unnest**
   - `Expr`: ArrayLiteral（配列式）
   - `As`: nil（エイリアスなし）
   - `WithOffset`: nil（OFFSETなし）
   - `Hint`, `Sample`: nil

2. **ast.ArrayLiteral**
   - `Type`: StructType（`STRUCT<x INT64, y STRING>`）
   - `Values`: 2個のTupleStructLiteral

3. **ast.StructType**
   - `Fields`: 2個のStructField
     - Field[0]: `x *ast.SimpleType(INT64)`
     - Field[1]: `y *ast.SimpleType(STRING)`

4. **ast.TupleStructLiteral** (2個)
   - 1個目: `(1, 'foo')` → Values: [IntLiteral("1"), StringLiteral("foo")]
   - 2個目: `(3, 'bar')` → Values: [IntLiteral("3"), StringLiteral("bar")]

#### パターン別結果

##### 1. AS句付き
```sql
UNNEST(...) AS t
```
- `Unnest.As`: AsAlias構造体でエイリアス情報を保持

##### 2. WITH OFFSET付き
```sql
UNNEST(...) WITH OFFSET
```
- `Unnest.WithOffset`: WithOffset構造体で行番号カラム情報を保持

##### 3. 単純な配列
```sql
UNNEST([1, 2, 3])
```
- `ArrayLiteral.Type`: nil（型推論）
- `Values`: 直接IntLiteralの配列

##### 4. 型指定付き単純配列
```sql
UNNEST(ARRAY<INT64>[1, 2, 3])
```
- `ArrayLiteral.Type`: SimpleType("INT64")
- `Values`: IntLiteralの配列

### AST構造の設計原則

1. **型情報の分離**
   - 型定義（`ARRAY<STRUCT<...>>`）と実際の値（`[(1, 'foo'), ...]`）を分離
   - 型チェックは後続のセマンティック解析で実施

2. **位置情報の完全保持**
   - 全てのトークンの位置情報を保持
   - エラー報告やIDE連携に必要

3. **柔軟な型システム**
   - 明示的型指定と型推論の両方をサポート
   - ネストした複雑な型も適切に表現

4. **SQL再構築可能性**
   - 全てのノードが`SQL()`メソッドを実装
   - パース→変換→再出力のワークフローをサポート

### 実装のポイント

1. **TableExprインターフェース**: UNNESTは他のテーブル参照と同等に扱える
2. **型階層の適切な表現**: SimpleType、StructType、ArrayTypeの階層
3. **値表現の多様性**: 異なるStructLiteral形式への対応
4. **オプション機能の管理**: As句、WithOffset句の適切な表現

### 実行方法

```bash
# 詳細分析の実行
go run main.go

# 様々なパターンの分析
go run variations.go
```

この実装により、BigQueryの複雑なUNNEST構文を完全に解析し、構造化された形で扱うことができます。特に、型システムの完全なサポートと柔軟な値表現により、様々なUNNESTパターンに対応できるのが特徴です。