# Diarizer

`diarizer`は、テキストファイル内の会話を解析し、話者分離を行うGo製のCLIツールです。OpenAI APIを利用して、各発言が誰によるものかを推測し、`[話者A]`のようなラベルを付与します。

## 主な機能

- テキストファイルを入力として話者分離を実行
- 結果をMarkdownファイルに出力
- 使用するOpenAIモデルを選択可能 (e.g., `gpt-4o`, `gpt-4-turbo`)

## インストール

Goの実行環境がセットアップされていることを確認してください。

```bash
go install github.com/user/diarizer
```

## 使い方

### 1. APIキーの設定

このツールを使用するには、OpenAI APIキーが必要です。プロジェクトのルートディレクトリに`.env`ファイルを作成し、以下のようにキーを設定してください。

```
OPENAI_API_KEY="sk-..."
```

または、環境変数として`OPENAI_API_KEY`を設定することも可能です。

### 2. コマンドの実行

話者分離を行いたいテキストファイルを引数として指定します。

```bash
diarizer <入力ファイル.txt>
```

#### オプション

- `-o <出力ファイル.md>`: 出力先のファイルパスを指定します。指定しない場合、入力ファイル名に`.diarized.md`が付与されたファイルが作成されます。
- `-m <モデル名>`: 使用するモデルを指定します。デフォルトは`gpt-4o`です。対応モデル: `gpt-4o`, `gpt-4-turbo`, `gpt-3.5-turbo`。

**実行例:**

```bash
# `meeting.txt`を話者分離し、`meeting.diarized.md`に出力する
diarizer meeting.txt

# gpt-4-turboモデルを使い、結果を`result.md`に出力する
diarizer -m gpt-4-turbo -o result.md meeting.txt
```
