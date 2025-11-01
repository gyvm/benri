package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
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

// buildPrompt はプロンプトを生成します。
func buildPrompt() string {
	return `
以下の音声データを分析し、指定された形式で感情分析レポートを生成してください。

出力形式はマークダウンのコードブロックを使わず、純粋なJSONオブジェクトのみとしてください。

感情ラベルについて：
- 以下は感情分析で使用できるラベルの例です。これらに限定されません。音声の内容に応じて、より適切な感情を自由に選択してください。
- 基本的な感情：喜び、悲しみ、怒り、恐怖、驚き、嫌悪、中立、期待
- ポジティブ感情：楽観的、愛情、感謝、満足、希望、安心、誇り、興奮
- ネガティブ感情：失望、後悔、不安、焦り、疲れ、イライラ、沈み込み、虚無感
- その他：迷い、困惑、同情、尊敬、興味、好奇心など

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

重要：
- sentiment_composition内の各感情のscoreは0-100の値を指定し、合計が100になるようにしてください。
- もしタイムスタンプの取得が不可能であれば、"timed_analysis" は空の配列 '[]' にしてください。
- transcriptフィールドには音声の完全な文字起こしを含めてください。
`
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

// getMimeTypeはファイルパスからMIMEタイプを判別します。
func getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".mp3":
		return "mp3"
	case ".wav":
		return "wav"
	default:
		return "wav"
	}
}

func main() {
	runAnalysis()
}

func runAnalysis() {
	// コマンドライン引数を定義
	outFile := flag.String("o", "", "出力ファイルパス (例: report.md)")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "エラー: 入力ファイルが指定されていません。")
		os.Exit(1)
	}
	inFile := flag.Arg(0)

	if *outFile == "" {
		base := strings.TrimSuffix(inFile, filepath.Ext(inFile))
		*outFile = base + ".md"
	}

	_ = godotenv.Load()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "エラー: 環境変数 OPENAI_API_KEY が設定されていません。")
		os.Exit(1)
	}

	// 1) 音声ファイルを読み込み、Base64エンコードする
	raw, err := os.ReadFile(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: ファイルが読み込めませんでした: %v\n", err)
		os.Exit(1)
	}
	audioB64 := base64.StdEncoding.EncodeToString(raw)
	fmt.Printf("音声ファイルを読み込みました (%d bytes)\n", len(raw))

	// ファイル拡張子からオーディオフォーマットを判定
	audioFormat := getMimeType(inFile)
	fmt.Printf("オーディオフォーマット: %s\n", audioFormat)

	// 2) gpt-4o-audio API にリクエストを送信
	result, apiResponseBody, err := analyzeEmotionWithGPT4oAudio(apiKey, audioB64, audioFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 感情分析に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 3) Markdownレポートを生成してファイルに書き込む
	markdownContent := generateMarkdown(*result, inFile, "gpt-4o-audio-preview", apiResponseBody)
	if err := os.WriteFile(*outFile, []byte(markdownContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: レポートの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正常に完了しました。レポートが", *outFile, "に保存されました。")
}

func analyzeEmotionWithGPT4oAudio(apiKey string, audioB64 string, audioFormat string) (*SentimentReport, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// プロンプトを生成
	prompt := buildPrompt()

	// リクエストボディを構築
	requestBody := map[string]interface{}{
		"model": "gpt-4o-audio-preview",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "input_audio",
						"input_audio": map[string]interface{}{
							"data":   audioB64,
							"format": audioFormat,
						},
					},
				},
			},
		},
	}

	// JSONにエンコード
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, "", fmt.Errorf("リクエストボディのJSONエンコードに失敗: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "", fmt.Errorf("HTTPリクエストの作成に失敗: %w", err)
	}

	// ヘッダを設定
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	fmt.Println("gpt-4o-audio APIにリクエストを送信中...")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("APIリクエストに失敗: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み込む
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("レスポンスボディの読み込みに失敗: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("APIエラー (ステータスコード: %d): %s", resp.StatusCode, string(respBody))
	}

	// レスポンスをパース
	var chatResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &chatResponse); err != nil {
		return nil, "", fmt.Errorf("レスポンスJSONのパースに失敗: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, "", fmt.Errorf("APIがmessagesを返しませんでした")
	}

	// JSONの抽出とパース
	content := chatResponse.Choices[0].Message.Content
	fullJSON := extractJSON(content)

	// APIレスポンスボディを整形して保存
	apiResponseBody := fullJSON

	// JSONをパース（コードブロック削除）
	cleanedJSON := strings.TrimPrefix(fullJSON, "```json")
	cleanedJSON = strings.TrimSuffix(cleanedJSON, "```")
	cleanedJSON = strings.TrimSpace(cleanedJSON)

	var result SentimentReport
	if err := json.Unmarshal([]byte(cleanedJSON), &result); err != nil {
		return nil, apiResponseBody, fmt.Errorf("結果JSONのパースに失敗: %w (内容: %s)", err, cleanedJSON)
	}

	fmt.Println("分析が完了しました。")
	return &result, apiResponseBody, nil
}

// extractJSON は、APIからの応答に含まれるJSON部分を抽出します。
func extractJSON(raw string) string {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start != -1 && end != -1 && start < end {
		return raw[start : end+1]
	}
	return raw
}
