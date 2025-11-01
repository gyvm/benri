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

// EmotionResult はAPIから返される感情分析結果のJSON構造体です。
type EmotionResult struct {
	Version   string             `json:"version"`
	Model     string             `json:"model"`
	Language  string             `json:"language"`
	Emotions  map[string]float64 `json:"emotions"`
	Valence   float64            `json:"valence"`
	Arousal   float64            `json:"arousal"`
	Dominance float64            `json:"dominance"`
	Notes     []string           `json:"notes"`
	// Optional: 推定された音響系の指標
	SpeakingRate *float64 `json:"speaking_rate,omitempty"`
	AvgPitchHz   *float64 `json:"avg_pitch_hz,omitempty"`
	EnergyProxy  *float64 `json:"energy_proxy,omitempty"`
	SilenceRatio *float64 `json:"silence_ratio,omitempty"`
}

// buildPrompt はプロンプトを生成します。
func buildPrompt() string {
	return `
以下の音声データを分析し、指定された形式で感情分析レポートを生成してください。

出力形式はマークダウンのコードブロックを使わず、純粋なJSONオブジェクトのみとしてください。

感情分析について：
- emotions フィールドには、以下の 7 つの基本感情を 0.0～1.0 の値で指定してください：
  - joy（喜び）
  - sadness（悲しみ）
  - anger（怒り）
  - fear（恐怖）
  - surprise（驚き）
  - disgust（嫌悪）
  - neutral（中立）
  各感情スコアの合計が 1.0 になるようにしてください。

- valence：感情価（-1～+1）
  - 負の値：不快な感情
  - 正の値：快い感情
  - 0：中立

- arousal：覚醒度（0～1）
  - 低い値：落ち着いた、リラックスした状態
  - 高い値：興奮した、活気のある状態

- dominance：優位度（0～1）
  - 低い値：受け身的、従属的な印象
  - 高い値：支配的、主導的な印象

- speaking_rate：推定話速（単語/分）
- avg_pitch_hz：推定平均ピッチ（Hz）
- energy_proxy：エネルギー指標（0～1）
- silence_ratio：無音比率（0～1）
- notes：分析の根拠を日本語で簡潔に記載

JSONスキーマ例：
{
  "version": "1.0",
  "model": "gpt-4o-audio-preview",
  "language": "ja",
  "emotions": {
    "joy": 0.35,
    "sadness": 0.05,
    "anger": 0.0,
    "fear": 0.0,
    "surprise": 0.15,
    "disgust": 0.0,
    "neutral": 0.45
  },
  "valence": 0.6,
  "arousal": 0.7,
  "dominance": 0.5,
  "speaking_rate": 150,
  "avg_pitch_hz": 200,
  "energy_proxy": 0.8,
  "silence_ratio": 0.1,
  "notes": ["音声は明るく、活気のあるトーン。複数の感情が混在しているが、全体的には前向きな印象。"]
}

重要：
- emotions オブジェクトのスコアの合計が 1.0 になることを確認してください。
- notes は配列で、複数の根拠がある場合は複数要素を含められます。
`
}

func writeMarkdownReport(er EmotionResult, audioFilePath string, apiResponseBody string) string {
	var builder strings.Builder

	builder.WriteString("# 感情分析レポート\n\n")
	builder.WriteString(fmt.Sprintf("- **ファイル名:** `%s`\n", filepath.Base(audioFilePath)))
	builder.WriteString(fmt.Sprintf("- **分析日時:** `%s`\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("- **使用モデル:** `%s`\n", er.Model))
	builder.WriteString(fmt.Sprintf("- **言語推定:** `%s`\n\n", er.Language))

	builder.WriteString("## 連続感情モデル (VAD)\n\n")
	builder.WriteString(fmt.Sprintf("- **Valence (快-不快):** %.2f (負=不快、正=快)\n", er.Valence))
	builder.WriteString(fmt.Sprintf("- **Arousal (覚醒-睡眠):** %.2f (低=リラックス、高=興奮)\n", er.Arousal))
	builder.WriteString(fmt.Sprintf("- **Dominance (優位-劣位):** %.2f (低=受け身、高=支配的)\n\n", er.Dominance))

	builder.WriteString("## 離散感情スコア\n\n")
	builder.WriteString("| 感情 | スコア |\n")
	builder.WriteString("| :--- | :---: |\n")

	// 感情スコアを表示
	emotionLabels := []string{"joy", "sadness", "anger", "fear", "surprise", "disgust", "neutral"}
	emotionNames := map[string]string{
		"joy":      "喜び",
		"sadness":  "悲しみ",
		"anger":    "怒り",
		"fear":     "恐怖",
		"surprise": "驚き",
		"disgust":  "嫌悪",
		"neutral":  "中立",
	}
	for _, label := range emotionLabels {
		if score, ok := er.Emotions[label]; ok {
			builder.WriteString(fmt.Sprintf("| %s (%s) | %.3f |\n", emotionNames[label], label, score))
		}
	}
	builder.WriteString("\n")

	builder.WriteString("## 音響分析（推定値）\n\n")
	if er.SpeakingRate != nil {
		builder.WriteString(fmt.Sprintf("- **話速:** %.1f 単語/分\n", *er.SpeakingRate))
	}
	if er.AvgPitchHz != nil {
		builder.WriteString(fmt.Sprintf("- **平均ピッチ:** %.1f Hz\n", *er.AvgPitchHz))
	}
	if er.EnergyProxy != nil {
		builder.WriteString(fmt.Sprintf("- **エネルギー:** %.3f\n", *er.EnergyProxy))
	}
	if er.SilenceRatio != nil {
		builder.WriteString(fmt.Sprintf("- **無音比率:** %.3f\n", *er.SilenceRatio))
	}
	builder.WriteString("\n")

	if len(er.Notes) > 0 {
		builder.WriteString("## 分析の根拠\n\n")
		for _, note := range er.Notes {
			builder.WriteString(fmt.Sprintf("- %s\n", note))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("---\n\n")
	builder.WriteString("## APIレスポンス\n\n")
	builder.WriteString("```json\n")
	builder.WriteString(apiResponseBody)
	builder.WriteString("\n```\n")

	return builder.String()
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
	ext := strings.ToLower(filepath.Ext(inFile))
	audioFormat := "wav" // デフォルトは wav
	if ext == ".mp3" {
		audioFormat = "mp3"
	}
	fmt.Printf("オーディオフォーマット: %s\n", audioFormat)

	// 2) gpt-4o-audio API にリクエストを送信
	result, apiResponseBody, err := analyzeEmotionWithGPT4oAudio(apiKey, audioB64, audioFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 感情分析に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 3) Markdownレポートを生成してファイルに書き込む
	markdownContent := writeMarkdownReport(*result, inFile, apiResponseBody)
	if err := os.WriteFile(*outFile, []byte(markdownContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: レポートの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正常に完了しました。レポートが", *outFile, "に保存されました。")
}

func analyzeEmotionWithGPT4oAudio(apiKey string, audioB64 string, audioFormat string) (*EmotionResult, string, error) {
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

	var result EmotionResult
	if err := json.Unmarshal([]byte(fullJSON), &result); err != nil {
		return nil, apiResponseBody, fmt.Errorf("結果JSONのパースに失敗: %w (内容: %s)", err, fullJSON)
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
