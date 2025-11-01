package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// --- APIレスポンス用の構造体 ---
type SentimentReport struct {
	OverallSentiment     OverallSentiment       `json:"overall_sentiment"`
	SentimentComposition []SentimentComposition `json:"sentiment_composition"`
	Transcript           string                 `json:"transcript"`
	DiarizedTranscript   string                 `json:"diarized_transcript,omitempty"`
	TimedAnalysis        []TimedAnalysis        `json:"timed_analysis"`
}
type OverallSentiment struct {
	Sentiment string `json:"sentiment"`
	Summary   string `json:"summary"`
}
type SentimentComposition struct {
	Sentiment string `json:"sentiment"`
	Score     int    `json:"score"`
}
type TimedAnalysis struct {
	Timestamp string `json:"timestamp"`
	Utterance string `json:"utterance"`
	Sentiment string `json:"sentiment"`
	Keywords  string `json:"keywords"`
}

// buildPromptは話者分離オプションに基づいてプロンプトを生成します。
func buildPrompt(diarize bool) string {
	basePrompt := `
以下の音声データを分析し、指定された形式で感情分析レポートを生成してください。

出力形式はマークダウンのコードブロックを使わず、純粋なJSONオブジェクトのみとしてください。

感情ラベルについて：
- 以下は感情分析で使用できるラベルの例です。これらに限定されません。音声の内容に応じて、より適切な感情を自由に選択してください。
- 基本的な感情：喜び、悲しみ、怒り、恐怖、驚き、嫌悪、中立、期待
- ポジティブ感情：楽観的、愛情、感謝、満足、希望、安心、誇り、興奮
- ネガティブ感情：失望、後悔、不安、焦り、疲れ、イライラ、沈み込み、虚無感
- その他：迷い、困惑、同情、尊敬、興味、好奇心など
`

	baseSchema := `
JSONスキーマ：
{
  "overall_sentiment": {
    "sentiment": "ポジティブ | ネガティブ | ニュートラル（音声全体の総合的な感情分類）",
    "summary": "感情の理由の短い要約"
  },
  "sentiment_composition": [
    {"sentiment": "喜び", "score": 35},
    {"sentiment": "楽観的", "score": 30},
    {"sentiment": "期待", "score": 20},
    {"sentiment": "中立", "score": 15}
  ],
  "transcript": "音声の完全な文字起こしテキスト。",
  "timed_analysis": [
    {
      "timestamp": "00:00-00:03",
      "utterance": "こんにちは、今日はとても良い天気ですね。",
      "sentiment": "喜び",
      "keywords": "良い天気、明るい"
    }
  ]
}
`

	diarizeSchema := `
JSONスキーマ（話者分離有効時）：
{
  "overall_sentiment": {
    "sentiment": "ポジティブ | ネガティブ | ニュートラル（音声全体の総合的な感情分類）",
    "summary": "感情の理由の短い要約"
  },
  "sentiment_composition": [
    {"sentiment": "喜び", "score": 35},
    {"sentiment": "楽観的", "score": 30},
    {"sentiment": "期待", "score": 20},
    {"sentiment": "中立", "score": 15}
  ],
  "transcript": "音声の完全な文字起こしテキスト。",
  "diarized_transcript": "[話者A] こんにちは。今日はいい天気ですね。\n[話者B] そうですね。本当にいい天気です。\n[話者A] また明日も良い天気が続くといいですね。",
  "timed_analysis": [
    {
      "timestamp": "00:00-00:03",
      "utterance": "こんにちは、今日はとても良い天気ですね。",
      "sentiment": "喜び",
      "keywords": "良い天気、明るい"
    }
  ]
}
`

	notes := `
重要：
- sentiment_composition内の各感情のscoreは0-100の値を指定し、合計が100になるようにしてください。
- もしタイムスタンプの取得が不可能であれば、"timed_analysis" は空の配列 '[]' にしてください。
`

	if diarize {
		diarizeNotes := `
話者分離について：
- "diarized_transcript"フィールドに話者分離を行った文字起こしを含めてください。
- 各発言の先頭に [話者A]、[話者B]、[話者C] などのラベルを付けてください。
- 複数の話者が検出された場合は、自然な割り当てで対応してください。
- 改行で複数の発言を区切ってください。
`
		return basePrompt + diarizeSchema + diarizeNotes + notes
	}

	return basePrompt + baseSchema + notes
}

// getMimeTypeはファイルパスからMIMEタイプを判別します。
func getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	default:
		return ""
	}
}

// generateMarkdownはSentimentReportからMarkdown文字列を生成します。
func generateMarkdown(report SentimentReport, audioFilePath, modelName, apiResponseBody string) string {
	var md strings.Builder

	md.WriteString("# 感情分析レポート\n\n")
	md.WriteString(fmt.Sprintf("- **ファイル名:** `%s`\n", filepath.Base(audioFilePath)))
	md.WriteString(fmt.Sprintf("- **分析日時:** `%s`\n", time.Now().Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("- **使用モデル:** `%s`\n\n", modelName))

	md.WriteString("## 総合的な感情\n\n")
	md.WriteString(fmt.Sprintf("この音声は全体的に **%s** な印象です。\n\n", report.OverallSentiment.Sentiment))
	md.WriteString(fmt.Sprintf("> %s\n\n", report.OverallSentiment.Summary))

	md.WriteString("## 感情の構成比\n\n")
	md.WriteString("| 感情 | 割合（%） |\n")
	md.WriteString("| :--- | :---: |\n")
	for _, s := range report.SentimentComposition {
		md.WriteString(fmt.Sprintf("| %s | %d |\n", s.Sentiment, s.Score))
	}
	md.WriteString("\n")

	if len(report.TimedAnalysis) > 0 {
		md.WriteString("## 発言ごとの感情分析 (時系列)\n\n")
		md.WriteString("| 時間 (秒) | 発言内容 | 感情 | 補足・キーワード |\n")
		md.WriteString("|:---|:---|:---|:---|\n")
		for _, t := range report.TimedAnalysis {
			md.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s |\n", t.Timestamp, t.Utterance, t.Sentiment, t.Keywords))
		}
		md.WriteString("\n")
	}

	md.WriteString("---\n\n")
	md.WriteString("## 音声の文字起こし\n\n")
	md.WriteString(fmt.Sprintf("```text\n%s\n```\n\n", report.Transcript))

	if report.DiarizedTranscript != "" {
		md.WriteString("---\n\n")
		md.WriteString("## 話者分離済み文字起こし\n\n")
		md.WriteString(fmt.Sprintf("```text\n%s\n```\n\n", report.DiarizedTranscript))
	}

	md.WriteString("---\n\n")
	md.WriteString("## APIレスポンス\n\n")
	md.WriteString("```json\n")
	md.WriteString(apiResponseBody)
	md.WriteString("\n```\n")

	return md.String()
}

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %v", err)
	}

	outputFlag := flag.String("o", "", "出力するMarkdownファイルのパス")
	modelFlag := flag.String("m", "models/gemini-2.5-pro", "使用するAIモデル名")
	diarizeFlag := flag.Bool("d", false, "話者分離を有効にするフラグ")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("エラー: 音声ファイルのパスが指定されていません。")
		os.Exit(1)
	}
	audioFilePath := flag.Arg(0)

	outputFilePath := *outputFlag
	if outputFilePath == "" {
		base := filepath.Base(audioFilePath)
		ext := filepath.Ext(base)
		outputFilePath = strings.TrimSuffix(base, ext) + ".md"
	}

	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("エラー: 環境変数 GEMINI_API_KEY が設定されていません。")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("APIクライアントの作成に失敗しました: %v", err)
	}
	defer client.Close()

	audioData, err := os.ReadFile(audioFilePath)
	if err != nil {
		log.Fatalf("音声ファイルの読み込みに失敗しました: %v", err)
	}

	mimeType := getMimeType(audioFilePath)
	if mimeType == "" {
		log.Fatalf("対応していないファイル形式です: %s", audioFilePath)
	}

	model := client.GenerativeModel(*modelFlag)

	// 話者分離オプションに応じてプロンプトを生成
	prompt := buildPrompt(*diarizeFlag)

	fmt.Println("APIにリクエストを送信しています...")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt), genai.Blob{MIMEType: mimeType, Data: audioData})
	if err != nil {
		log.Fatalf("APIリクエストに失敗しました: %v", err)
	}

	// レスポンスのテキスト部分を取得
	var apiResponseText string
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				apiResponseText = string(txt)
				break
			}
		}
		if apiResponseText != "" {
			break
		}
	}

	// JSONをパース
	var report SentimentReport
	// APIからのレスポンスに含まれるコードブロックの```jsonと```を削除
	apiResponseText = strings.TrimPrefix(apiResponseText, "```json")
	apiResponseText = strings.TrimSuffix(apiResponseText, "```")
	if err := json.Unmarshal([]byte(apiResponseText), &report); err != nil {
		log.Fatalf("APIレスポンスのJSONパースに失敗しました: %v\nレスポンス内容:\n%s", err, apiResponseText)
	}

	// Markdownを生成
	markdownContent := generateMarkdown(report, audioFilePath, *modelFlag, apiResponseText)

	// ファイルに書き込み
	if err := os.WriteFile(outputFilePath, []byte(markdownContent), 0644); err != nil {
		log.Fatalf("Markdownファイルの書き込みに失敗しました: %v", err)
	}

	fmt.Printf("処理が完了しました。レポートが %s に保存されました。\n", outputFilePath)
}
