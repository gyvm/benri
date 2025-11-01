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

func writeMarkdownReport(er EmotionResult, path string) error {
	var builder strings.Builder

	builder.WriteString("# 感情分析レポート\n\n")
	builder.WriteString(fmt.Sprintf("- **モデル:** %s\n", er.Model))
	builder.WriteString(fmt.Sprintf("- **言語推定:** %s\n", er.Language))
	builder.WriteString(fmt.Sprintf("- **バージョン:** %s\n\n", er.Version))

	builder.WriteString("## 連続感情モデル (VAD)\n\n")
	builder.WriteString(fmt.Sprintf("- **Valence (快-不快):** %.2f\n", er.Valence))
	builder.WriteString(fmt.Sprintf("- **Arousal (覚醒-睡眠):** %.2f\n", er.Arousal))
	builder.WriteString(fmt.Sprintf("- **Dominance (優位-劣位):** %.2f\n\n", er.Dominance))

	builder.WriteString("## 離散感情スコア\n\n")
	for k, v := range er.Emotions {
		builder.WriteString(fmt.Sprintf("- %s: %.2f\n", k, v))
	}
	builder.WriteString("\n")

	if len(er.Notes) > 0 {
		builder.WriteString("## 注記\n\n")
		for _, n := range er.Notes {
			builder.WriteString(fmt.Sprintf("- %s\n", n))
		}
		builder.WriteString("\n")
	}

	hasAcoustics := er.SpeakingRate != nil || er.AvgPitchHz != nil || er.EnergyProxy != nil || er.SilenceRatio != nil
	if hasAcoustics {
		builder.WriteString("## 音響分析（推定値）\n\n")
		if er.SpeakingRate != nil {
			builder.WriteString(fmt.Sprintf("- **話速 (単語/分):** %.2f\n", *er.SpeakingRate))
		}
		if er.AvgPitchHz != nil {
			builder.WriteString(fmt.Sprintf("- **平均ピッチ (Hz):** %.2f\n", *er.AvgPitchHz))
		}
		if er.EnergyProxy != nil {
			builder.WriteString(fmt.Sprintf("- **エネルギープロキシ:** %.2f\n", *er.EnergyProxy))
		}
		if er.SilenceRatio != nil {
			builder.WriteString(fmt.Sprintf("- **無音比率:** %.2f\n", *er.SilenceRatio))
		}
		builder.WriteString("\n")
	}

	return os.WriteFile(path, []byte(builder.String()), 0644)
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

	// 2) gpt-4o-audio API にリクエストを送信
	result, err := analyzeEmotionWithGPT4oAudio(apiKey, audioB64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 感情分析に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 3) Markdownレポートを生成してファイルに書き込む
	if err := writeMarkdownReport(*result, *outFile); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: レポートの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正常に完了しました。レポートが", *outFile, "に保存されました。")
}

func analyzeEmotionWithGPT4oAudio(apiKey string, audioB64 string) (*EmotionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// リクエストボディを構築
	requestBody := map[string]interface{}{
		"model": "gpt-4o-audio-preview",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "次の音声の話者感情を分析し、以下の情報をJSON形式で返してください。JSONオブジェクトのみを返し、他のテキストは不要です:\n{\n  \"version\": \"1.0\",\n  \"model\": \"gpt-4o-audio-preview\",\n  \"language\": \"ja\",\n  \"emotions\": {\n    \"joy\": 0.0,\n    \"sadness\": 0.0,\n    \"anger\": 0.0,\n    \"fear\": 0.0,\n    \"surprise\": 0.0,\n    \"disgust\": 0.0,\n    \"neutral\": 0.5\n  },\n  \"valence\": 0.0,\n  \"arousal\": 0.5,\n  \"dominance\": 0.5,\n  \"notes\": [\"分析の根拠を日本語で簡潔に記載\"]\n}",
					},
					{
						"type": "input_audio",
						"input_audio": map[string]interface{}{
							"data":   audioB64,
							"format": "wav",
						},
					},
				},
			},
		},
	}

	// JSONにエンコード
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("リクエストボディのJSONエンコードに失敗: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエストの作成に失敗: %w", err)
	}

	// ヘッダを設定
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	fmt.Println("gpt-4o-audio APIにリクエストを送信中...")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストに失敗: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み込む
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスボディの読み込みに失敗: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("APIエラー (ステータスコード: %d): %s", resp.StatusCode, string(respBody))
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
		return nil, fmt.Errorf("レスポンスJSONのパースに失敗: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("APIがmessagesを返しませんでした")
	}

	// JSONの抽出とパース
	content := chatResponse.Choices[0].Message.Content
	fullJSON := extractJSON(content)

	var result EmotionResult
	if err := json.Unmarshal([]byte(fullJSON), &result); err != nil {
		return nil, fmt.Errorf("結果JSONのパースに失敗: %w (内容: %s)", err, fullJSON)
	}

	fmt.Println("分析が完了しました。")
	return &result, nil
}

func getSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"version":       map[string]interface{}{"type": "string"},
			"model":         map[string]interface{}{"type": "string"},
			"language":      map[string]interface{}{"type": "string"},
			"emotions":      map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{"type": "number"}},
			"valence":       map[string]interface{}{"type": "number"},
			"arousal":       map[string]interface{}{"type": "number"},
			"dominance":     map[string]interface{}{"type": "number"},
			"notes":         map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			"speaking_rate": map[string]interface{}{"type": "number"},
			"avg_pitch_hz":  map[string]interface{}{"type": "number"},
			"energy_proxy":  map[string]interface{}{"type": "number"},
			"silence_ratio": map[string]interface{}{"type": "number"},
		},
		"required":             []string{"version", "model", "language", "emotions", "valence", "arousal", "dominance"},
		"additionalProperties": false,
	}
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
