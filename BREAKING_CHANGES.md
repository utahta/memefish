# Breaking Changes

このドキュメントは、commit `0a1141e8eb65d172230cc00ab072b252f1674a8a` と `ffe92eac2e61da887872db1fb96c498e5c06ef45` 間の破壊的変更をリストします。

## 1. INTERVAL TO mandatory (#312) - commit ffe92ea

**変更内容**: `INTERVAL`リテラルで`TO`キーワードが必須になりました。

**影響**: 
- **変更前**: `INTERVAL "1" YEAR`
- **変更後**: `INTERVAL "1" YEAR TO MONTH`

**移行方法**: 以前`TO`句を省略していた`INTERVAL`リテラルにすべて`TO`句を追加してください。

## 2. QueryExpr and pipe syntax support (#201) - commit 4c78bec

### ast.Select構造の変更

**破壊的変更**:
- `OrderBy`と`Limit`フィールドが`ast.Select`から削除され、`ast.Query`に移動
- `Distinct bool`フィールドが`AllOrDistinct AllOrDistinct`に変更
- `QueryStatement`の`With`フィールドが削除され、`ast.Query`に移動

**新しいノード型**:
- `ast.Query`: オプションのCTE、ORDER BY、LIMIT、pipe operatorを持つクエリ表現
- `ast.FromQuery`: FROM句のみのクエリ表現
- `ast.PipeOperator`: パイプ演算子インターフェース
- `ast.PipeSelect`, `ast.PipeWhere`: パイプ演算子の実装

**移行方法**:
```go
// 変更前
select.OrderBy  // → 親のQueryノードからアクセス
select.Limit    // → 親のQueryノードからアクセス
select.Distinct // → select.AllOrDistinctを使用

// QueryStatement
stmt.With       // → stmt.Query.Withを使用（Queryがast.Query型の場合）
```

## 3. Path関連の変更

### 3.1 Enable path keys in hints (#250) - commit f71f971

**破壊的変更**: `HintRecord.Key`の型が`*Ident`から`*Path`に変更されました。

**移行方法**:
```go
// 変更前
hint.Key.Name

// 変更後
hint.Key.Idents[0].Name  // 単純な識別子の場合
// "spanner.optimizer_version"のようなパス表現の場合
// hint.Key.Idents[0].Name + "." + hint.Key.Idents[1].Name
```

### 3.2 Support path function calls (#232) - commit 11a7b9c

**破壊的変更**: `CallExpr.Func`の型が`*Ident`から`*Path`に変更されました。

**移行方法**:
```go
// 変更前
call.Func.Name

// 変更後
call.Func.Idents[0].Name  // 単純な関数名の場合
// "SAFE.NET.HOST"のようなパス関数呼び出しの場合
// call.Func.Idents[0].Name, call.Func.Idents[1].Name などを使用
```

### 3.3 Introduce path table expression (#159) - commit a61894d

**破壊的変更**:
- FROM句のパス表現用に新しい`PathTableExpr`ノード型が追加
- `Unnest.Implicit`フィールドが削除

**移行方法**:
```go
// 変更前
unnest.Implicit  // このフィールドは存在しません

// 変更後
// named schemaまたはimplicit UNNESTの表現にはPathTableExprを使用
// パーサーではこれらを区別できません - 後の段階で処理されます
```

## 4. StructLiteral分離とNEW Constructor (#183) - commit db40951

### StructLiteralの破壊的変更

**破壊的変更**: `StructLiteral`型が3つの独立した型に分離されました：

1. **`TupleStructLiteral`**: タプル構文 `(value1, value2)`
2. **`TypedStructLiteral`**: 型付き構文 `STRUCT<field1 type1, field2 type2>(value1, value2)`
3. **`TypelessStructLiteral`**: 型なし構文 `STRUCT(value1, value2)`

### 新しいインターフェース

- **`TypelessStructLiteralArg`**: typeless STRUCT literalの引数のインターフェース
  - 実装者: `ExprArg`, `Alias`
- **`NewConstructorArg`**: NEW constructorの引数のインターフェース
  - 実装者: `ExprArg`, `Alias`

### NEW Constructor型

**新しい式の型**:
- `NewConstructor`: `NEW Type(args...)`
- `BracedNewConstructor`: `NEW Type {...}`
- `BracedConstructor`: `{field1: value1, field2: value2}`
- `BracedConstructorField`
- `BracedConstructorFieldValue` (インターフェース)
- `BracedConstructorFieldValueExpr`

### AsAliasの変更

**破壊的変更**: `AsAlias.As`フィールがオプションになりました。

**移行方法**:
```go
// 変更前
// ASキーワードは常に存在していました

// 変更後
if !asAlias.As.Invalid() {
    // ASキーワードが存在
} else {
    // ASキーワードが省略されています
}
```

### Aliasの使用範囲拡張

**変更**: `Alias`ノードがSELECT結果列だけでなく、typeless STRUCT literalsやNEW constructorsでも使用可能になりました。

## 移行の概要

ASTを使用するコードを移行するには：

1. **INTERVALリテラル**: 不足している`TO`句を追加
2. **ast.Select**: 親の`Query`ノードから`OrderBy`/`Limit`にアクセス、`Distinct`の代わりに`AllOrDistinct`を使用
3. **ヒントキー**: `Ident.Name`の代わりに`Path.Idents[0].Name`を使用
4. **関数呼び出し**: `Ident.Name`の代わりに`Path.Idents[0].Name`を使用
5. **StructLiteral**: 1つではなく3つの独立したstruct literal型を処理
6. **AsAlias**: 省略されたASキーワードを検出するために`As.Invalid()`をチェック

## バージョン互換性

これらの変更はAST構造に影響し、後方互換性はありません。古いAST構造に依存するコードは、新しい構造で動作するように更新する必要があります。