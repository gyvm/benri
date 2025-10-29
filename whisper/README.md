# Whisper CLI Tool

OpenAIのWhisper APIを利用して、音声ファイルの文字起こしを行うGo製のコマンドラインツールです。

## インストール

```bash
go install github.com/user/whisper
```

## 事前準備

このツールを使用するには、OpenAIのAPIキーが必要です。
ツールの実行ディレクトリに `.env` という名前のファイルを作成し、以下のようにAPIキーを記述してください。

```
OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

## 使い方

### 基本的な使い方

文字起こししたい音声ファイルを引数に指定して実行します。
文字起こし結果は、音声ファイルと同じ階層に `（音声ファイル名）.md` という名前で保存されます。

```bash
whisper audio.mp3
```

### オプション

#### `-o`: 出力ファイル名の指定

`-o` フラグを使って、出力されるMarkdownファイルの名前を任意に指定できます。

```bash
whisper -o transcript.md audio.mp3
```

#### `-lang`: 言語の指定

`-lang` フラグを使って、音声の言語を指定できます。指定しない場合は、Whisperが自動で言語を検出します。
現在サポートしている言語は `ja`（日本語）と `en`（英語）です。

```bash
whisper -lang ja audio_japanese.mp3
```
