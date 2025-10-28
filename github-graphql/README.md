# GitHub Pull Request Analyzer

GitHubリポジトリのプルリクエスト情報を分析し、サイクルタイムなどをMarkdown形式で出力するCLIツールです。

## 機能

-   指定したリポジトリの直近のプルリクエスト情報を取得します。
-   以下の情報をMarkdown形式でレポートします。
    -   PRの作成者、作成日時、マージ日時
    -   リードタイム (PR作成からマージまでの時間)
    -   最初のコミットからマージまでの時間
    -   PR作成から最初のレビューまでの時間
    -   レビュー担当者とレビュー内容
    -   コメント

## 必要なもの

-   Go (1.18以上)
-   GitHub Personal Access Token

## セットアップ

1.  **リポジトリをクローンします。**

2.  **Personal Access Token を設定します。**

    リポジトリのルートに `.env` ファイルを作成し、GitHub Personal Access Tokenを記述します。`repo` スコープを持つトークンが必要です。

    ```
    GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
    ```

3.  **依存関係をインストールします。**

    ```bash
    go mod tidy
    ```

## 使い方

以下のコマンドを実行します。`<owner>/<repository>` には対象のリポジトリ名を、`-n` オプションで取得したいPRの件数を指定します（オプションを省略した場合、デフォルトで10件取得します）。

```bash
go run main.go [-n <件数>] <owner>/<repository>
```

### 実行例

```bash
go run main.go -n 20 octocat/Hello-World
```

### 出力例

```markdown
# Pull Request Analysis Report for octocat/Hello-World

## PR #123: Feature: Add new login button
- **Author:** user-a
- **Created at:** 2023-10-27T10:00:00Z
- **Merged at:** 2023-10-28T15:30:00Z
- **Lead Time:** 29h30m0s
- **Commit to Merge Time:** 29h25m0s
- **Time to First Review:** 28h0m0s

### Reviews
- **user-b:** APPROVED - 2023-10-28T14:00:00Z

### Comments
- No comments

---
```
