# emotion-analyzer

`emotion-analyzer` は、音声ファイルから話者の感情を分析し、レポートを生成するCLIツールです。OpenAIのリアルタイム音声APIを利用して、音声データを直接解析し、感情のスコアや関連する音響特徴を抽出します。

## 主な機能

- 音声ファイル（WAV, MP3などAPIがサポートする形式）を入力
- 感情の多次元的な分析（離散感情、VADモデル）
- Markdown形式での詳細なレポート出力
- 設定が容易なコマンドラインインターフェース

## インストール

Go言語の環境が設定されていることを前提とします。

1.  **リポジトリのクローン:**
    ```bash
    git clone <リポジトリのURL>
    ```

2.  **ツールディレクトリへの移動:**
    ```bash
    cd <リポジトリのパス>/emotion-analyzer
    ```

3.  **依存関係のインストール:**
    ```bash
    go mod tidy
    ```

4.  **ビルド:**
    ```bash
    go build -o emotion-analyzer .
    ```
    これにより、実行可能な `emotion-analyzer` ファイルが生成されます。

## 設定

このツールを使用するには、OpenAIのAPIキーが必要です。

1.  `emotion-analyzer` ディレクトリ内に `.env` という名前のファイルを作成します。
2.  ファイルに以下の内容を記述し、`YOUR_API_KEY_HERE` の部分をあなたの実際のAPIキーに置き換えてください。

    ```
    OPENAI_API_KEY="YOUR_API_KEY_HERE"
    ```

    このファイルは `.gitignore` に追加することが推奨されます。

## 使い方

ビルドして生成された `emotion-analyzer` を使って、音声ファイルを分析します。

### 基本的な使用法

分析したい音声ファイルを引数として渡します。

```bash
./emotion-analyzer <path/to/your/audio.wav>
```

実行が成功すると、入力ファイルと同じディレクトリに `audio.md` のような名前でレポートファイルが自動的に生成されます。

### 出力ファイルの指定

`-o` フラグを使って、レポートの出力先を自由に指定することもできます。

```bash
./emotion-analyzer -o <path/to/report.md> <path/to/your/audio.wav>
```

### ヘルプ

利用可能なオプションについては、`-h` または `--help` フラグで確認できます。

```bash
./emotion-analyzer -h
```
