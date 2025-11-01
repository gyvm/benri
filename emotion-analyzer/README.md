# emotion-analyzer

`emotion-analyzer` は、音声ファイルから話者の感情を分析し、レポートを生成するCLIツールです。OpenAIの gpt-4o-audio-preview モデルを利用して、音声データを直接解析し、感情のスコアや関連する音響特徴を抽出します。

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

## 技術仕様

このツールは OpenAI の **gpt-4o-audio-preview** モデルを使用して感情分析を行います。

- **エンドポイント**: `https://api.openai.com/v1/messages`
- **入力形式**: Base64 エンコードされた音声データ
- **出力形式**: JSON Schema による構造化レスポンス
- **対応フォーマット**: WAV, MP3, M4A, FLAC, OGG

### 分析結果に含まれる項目

- **version**: スキーマバージョン
- **model**: 使用したモデル名
- **language**: 推定される話者の言語
- **emotions**: 感情スコア
  - joy (喜び): 0.0～1.0
  - sadness (悲しみ): 0.0～1.0
  - anger (怒り): 0.0～1.0
  - fear (恐れ): 0.0～1.0
  - surprise (驚き): 0.0～1.0
  - disgust (嫌悪): 0.0～1.0
  - neutral (中立): 0.0～1.0
- **VAD モデル** (連続感情モデル)
  - valence (快-不快): -1.0～+1.0
  - arousal (覚醒-睡眠): 0.0～1.0
  - dominance (優位-劣位): 0.0～1.0
- **notes**: 分析の根拠に関する短い注記（日本語）
- **音響分析** (オプション)
  - speaking_rate: 話速（単語/分）
  - avg_pitch_hz: 平均ピッチ（Hz）
  - energy_proxy: エネルギープロキシ
  - silence_ratio: 無音比率
