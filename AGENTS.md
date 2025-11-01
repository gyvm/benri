# AGENTS.md

このドキュメントは、このリポジトリで作業するAIアシスタント向けの説明書です。

## 開発フロー

新しいツールを追加する際は、ツール名のディレクトリを新たに作成してください。ソースコード、`README.md`、`AGENTS.md`など、ツールに関連するすべてのファイルは、このディレクトリ内に配置してください。

## バージョン管理

- **ブランチ名**: ブランチ名は英語を使用してください。
- **コミットメッセージ**: コミットメッセージは日本語で記述してください。

## ドキュメント

- `README.md`、`AGENTS.md`、およびGitHubのPull Requestは日本語で記述してください。

## プログラミング言語

このリポジトリで使用する主な言語は、TypeScriptとGoです。

## Goの依存関係管理

Goプロジェクトに新しいライブラリ（依存関係）を追加する際は、以下の手順に従ってください。これにより、`go.mod`ファイル内の依存関係が正しく管理され、`// indirect`の記述が適切に扱われます。

## Go CLIツール開発の規約

GoでCLIツールを開発する際は、以下の規約に従ってください。

- **入力引数**: ファイルパスやリポジトリ名などの主要な入力は、 positional argument（位置引数）として受け取ります。
  ```bash
  # 良い例
  my-tool <input-file.txt>
  ```
- **出力パス**: 出力ファイルのパスを指定するために、 `-o` フラグを設けてください。このフラグは任意（optional）とします。
- **デフォルトの出力ファイル名**: `-o` フラグが指定されなかった場合、出力ファイル名は入力ファイル名から自動的に決定されるべきです。例えば、 `input.mp3` を処理した結果は `input.md` として保存します。

1.  **新しいライブラリの追加**:
    `go get`コマンドを使用して、必要なライブラリをプロジェクトに追加します。

    ```bash
    # 例: a-new-library を追加する場合
    go get github.com/owner/a-new-library
    ```

2.  **`go.mod`ファイルの整理**:
    ライブラリを追加または削除した後は、必ず`go mod tidy`コマンドを実行してください。このコマンドは、ソースコードを静的解析し、`go.mod`ファイルを最新の状態に更新します。
    -   ソースコードから直接`import`されているライブラリは、直接的な依存関係としてマークされます（`// indirect`が外れます）。
    -   不要になった依存関係は`go.mod`から削除されます。

    ```bash
    go mod tidy
    ```

## Secrets Management

For tools requiring secrets like API tokens, use a `.env` file to load them as environment variables. The `github.com/joho/godotenv` Go package is suitable for this.
