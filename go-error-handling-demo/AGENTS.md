# Go エラーハンドリング デモプロジェクト - AGENTS.md

このドキュメントは、go-error-handling-demo プロジェクトで作業するAIアシスタント向けの説明書です。

## プロジェクト概要

このプロジェクトは、Go における 5 つの異なるエラー処理戦略を比較・検証するデモです。各戦略の実装、動作確認、および学習リソースの充実化が継続的に行われています。

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
├── go.mod
├── go.sum
├── README.md
└── AGENTS.md                    # このファイル
```

## アーキテクチャ

このプロジェクトはクリーンアーキテクチャに基づいており、以下の層構造を採用しています：

### Usecase Layer (`internal/usecase/user_usecase.go`)
- ビジネスロジック実装
- 5 つのエラーハンドリング戦略の実装
- リポジトリレイヤーへの依存

### Server Layer (`internal/server/server.go`)
- API ハンドラー実装
- Usecase レイヤーの呼び出し
- エラーレスポンス処理

### Repository Layer (`internal/repository/user_repository.go`)
- データアクセス実装
- 最下位層

### Utilities (`internal/errorsdemo/print.go`)
- エラー出力フォーマッティング
- デバッグ情報の整形

## 実行方法

```bash
go run main.go
```

## 依存関係

このプロジェクトが使用する主なライブラリ：

- `github.com/cockroachdb/errors` - 高機能エラーハンドリング
- `github.com/pkg/errors` - スタックトレース付きエラー
- `go.uber.org/multierr` - 複数エラー管理

すべての依存関係は `go.mod` で管理されています。

## 開発ガイドライン

### コード追加時の規約

1. **ファイル配置**: 新しいコンポーネントは `internal/` 配下の適切なディレクトリに配置します
   - ビジネスロジック → `usecase/`
   - API ハンドラー → `server/`
   - データアクセス → `repository/`
   - ユーティリティ → `errorsdemo/`

2. **クリーンアーキテクチャの原則**:
   - 各レイヤーは疎結合に保つ
   - 外部ライブラリへの依存は最小化する
   - 依存性注入により、テスト容易性を確保

3. **エラーハンドリング**:
   - 各ユースケースメソッドで異なる戦略を検証
   - エラーの種類に応じて適切なハンドリング手法を選択

### テスト

テストケースが必要な場合は、各実装ファイルと同じディレクトリに `*_test.go` ファイルを作成します。

例：
- `internal/usecase/user_usecase_test.go`
- `internal/repository/user_repository_test.go`

## 依存関係の管理

新しいライブラリを追加する場合：

```bash
go get github.com/owner/library-name
go mod tidy
```

`go mod tidy` を実行して、`go.mod` ファイルを最新の状態に更新してください。

## ドキュメント

- `README.md` - ユーザー向けのドキュメント
- `AGENTS.md` - このファイル（AI開発者向け）
- `main.go` - コード内コメントで戦略の説明あり

## その他のノート

- Go バージョン: 1.24.3 以上
- このプロジェクトは学習目的での参考実装です
- 本番環境での使用時は、プロジェクトの要件に応じて適切なエラーハンドリング戦略を選択してください
