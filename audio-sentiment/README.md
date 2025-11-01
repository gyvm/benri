# 音声感情分析ツール

これは、音声ファイルから感情を分析し、結果をMarkdown形式のレポートとして出力するCLIツールです。

## 使い方

```bash
# 基本的な使い方
./audio-sentiment [オプション] <音声ファイルパス>

# 例1: 基本的な感情分析
./audio-sentiment -o report.md audio/my_voice.mp3

# 例2: 話者分離を有効にして分析
./audio-sentiment -d -o report.md audio/my_voice.mp3

# 例3: 特定のモデルを使用して話者分離で分析
./audio-sentiment -m models/gemini-pro-latest -d audio/my_voice.mp3
```

**重要:** オプションフラグ (`-o`, `-m`) は、必ず音声ファイルパスの前に指定してください。

### オプション

- `-o <出力ファイルパス>`: 出力するMarkdownファイルの名前を指定します。指定しない場合は、入力ファイル名に基づいて自動的に決まります。
- `-m <モデル名>`: 使用するAIモデルを指定します。指定しない場合は、デフォルトで `models/gemini-2.5-pro` が使用されます。
- `-d`: 話者分離を有効にします。このオプションを使用すると、文字起こし結果に `[話者A]`、`[話者B]` などのラベルが付与されます。

## セットアップ

1.  リポジトリをクローンします。
2.  `.env.example` を参考に `.env` ファイルを作成し、お使いのAPIキーを設定してください。

    ```
    GEMINI_API_KEY=YOUR_API_KEY
    ```
3.  依存関係をインストールします。
    ```bash
    go mod tidy
    ```
4.  ビルドします。
    ```bash
    go build
    ```
