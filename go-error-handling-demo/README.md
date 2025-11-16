# Go エラーハンドリング デモプロジェクト

このプロジェクトは、Go における 5 つの異なるエラー処理戦略を比較・検証するための包括的なデモです。クリーンアーキテクチャに基づいた設計で、各戦略の違いを明確に示します。

## 概要

Go のエラーハンドリングには複数のアプローチがあります。このプロジェクトでは、以下の 5 つの戦略を実装し、スタックトレース、エラーチェーン、および `errors.Is`/`As` での挙動の違いを実演します：

1. **Plain Error** (`errors.New`)
2. **fmt.Errorf with %w**
3. **github.com/pkg/errors**
4. **github.com/cockroachdb/errors**
5. **go.uber.org/multierr**

## プロジェクト構成

```
go-error-handling-demo/
├── main.go                      # エントリーポイント
├── internal/
│   ├── errorsdemo/
│   │   └── print.go            # エラー出力ユーティリティ
│   ├── repository/
│   │   └── user_repository.go  # リポジトリレイヤー
│   ├── server/
│   │   └── server.go           # ハンドラレイヤー
│   └── usecase/
│       └── user_usecase.go     # ユースケースレイヤー
├── go.mod                        # 依存関係管理
└── README.md                     # このファイル
```

## アーキテクチャ

このプロジェクトはクリーンアーキテクチャの原則に従っており、以下のレイヤー構造を採用しています：

- **Server Layer**: HTTP リクエストハンドラー、API エンドポイント
- **Usecase Layer**: ビジネスロジック、エラーハンドリング戦略の実装
- **Repository Layer**: データアクセス、永続化層

各レイヤーは依存性注入を通じて疎結合に保たれています。

## 実行方法

### 前提条件

- Go 1.24.3 以上

### セットアップ

```bash
cd go-error-handling-demo
go mod download
```

### 実行

```bash
go run main.go
```

## 各エラーハンドリング戦略の特徴

### 1. Plain Error (`errors.New`)

```go
err := errors.New("user not found")
```

- **特徴**: Go の最も基本的なエラー形式
- **スタックトレース**: 保持しない
- **用途**: 最も単純なエラー通知で十分な場合
- **デメリット**: デバッグに必要なコンテキストが不足しがちで、エラー発生源の特定が困難

### 2. fmt.Errorf with %w

```go
if err != nil {
    return fmt.Errorf("failed to fetch user: %w", err)
}
```

- **特徴**: Go 1.13 で導入された標準的なエラーラッピング機能
- **スタックトレース**: 保持しない
- **エラー構造**: エラーチェーンを保持し、`errors.Is`/`errors.As` と互換性あり
- **利点**: 標準ライブラリのみで完結でき、追加のコンテキストを付与しながら元のエラー情報も保持

### 3. github.com/pkg/errors

```go
if err != nil {
    return errors.Wrap(err, "failed to fetch user")
}
```

- **特徴**: Goコミュニティで長年デファクトスタンダードとされてきたライブラリ
- **スタックトレース**: `%+v` で詳細な情報を表示
- **利点**: スタックトレース自動付与により、エラー発生箇所の特定が容易
- **用途**: 詳細なデバッグ情報が不可欠なプロダクション環境

### 4. github.com/cockroachdb/errors

```go
if err != nil {
    return errors.Wrapf(err, "failed to fetch user")
}
```

- **特徴**: `pkg/errors` 機能に加え、分散システム対応の高機能ライブラリ
- **スタックトレース**: 自動でキャプチャ、非常に詳細な情報を提供
- **エラー構造**: ドメイン情報やヒントを追加可能なリッチな構造
- **用途**: マイクロサービス、分散システムなどの複雑なエラーハンドリング

### 5. go.uber.org/multierr

```go
var multiErr error
multiErr = multierr.Append(multiErr, err1)
multiErr = multierr.Append(multiErr, err2)
```

- **特徴**: 複数のエラーを単一のエラーオブジェクトに集約
- **スタックトレース**: 個々のエラーがスタックトレースを持っていれば、それを保持
- **エラー構造**: `multierr.Error` で複数エラーを一度に管理
- **用途**: バッチ処理、バリデーション、複数 goroutine の終了処理など

## 実行例

```bash
$ go run main.go

=== 1. Plain Error Case ===
Error: user not found
...

=== 2. fmt.Errorf with %w Case ===
Error: failed to fetch user: user not found
...

=== 3. pkg/errors Case ===
Error: failed to fetch user
user not found
...

=== 4. cockroachdb/errors Case ===
Error: failed to fetch user
user not found
...

=== 5. MultiError Case ===
Error 1: validation failed
Error 2: fetch failed
...
```

## 学習のポイント

このデモを通じて、以下のポイントを理解できます：

1. **スタックトレースの重要性**: エラー発生時のコールスタックは、本番環境でのデバッグにおいて非常に重要
2. **エラーチェーン**: 複数のレイヤーを通じたエラーの伝播と、元のエラー情報の保持
3. **標準ライブラリ vs 外部ライブラリ**: Go 1.13 以降の標準機能で十分なケースと、高度な機能が必要なケース
4. **複数エラーの処理**: 単一のエラーだけでなく、複数エラーの効率的な管理方法

## 参考資料

- [Go Errors](https://golang.org/pkg/errors/)
- [pkg/errors - GitHub](https://github.com/pkg/errors)
- [cockroachdb/errors - GitHub](https://github.com/cockroachdb/errors)
- [multierr - GitHub](https://github.com/uber-go/multierr)
- [Effective Go - Error handling](https://golang.org/doc/effective_go#error_handling)
