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

// OpenAI API Response structs
type MessageResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Content []Content `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func writeMarkdownReport(er EmotionResult, path string) error {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# 感情分析レポート\n\n"))
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

	// 1) 音声ファイルを読み込む
	raw, err := os.ReadFile(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: ファイルが読み込めませんでした: %v\n", err)
		os.Exit(1)
	}

	// 2) MIMEタイプを決定
	ext := strings.ToLower(filepath.Ext(inFile))
	mimeType := getMimeType(ext)
	fmt.Printf("ファイル形式: %s (MIME: %s)\n", ext, mimeType)

	// 3) Base64エンコード
	audioB64 := base64.StdEncoding.EncodeToString(raw)

	// 4) 感情分析リクエストを実行
	result, err := analyzeEmotion(apiKey, audioB64, mimeType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 感情分析に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 5) Markdownレポートを生成してファイルに書き込む
	if err := writeMarkdownReport(*result, *outFile); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: レポートの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正常に完了しました。レポートが", *outFile, "に保存されました。")
}

func getMimeType(ext string) string {
	switch ext {
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	case ".flac":
		return "audio/flac"
	case ".ogg":
		return "audio/ogg"
	default:
		return "audio/wav"
	}
}

func analyzeEmotion(apiKey, audioB64, mimeType string) (*EmotionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// スキーマを定義
	schema := getSchema()

	// リクエストボディを構築
	requestBody := map[string]any{
		"model": "gpt-4o-audio-preview",
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "text",
						"text": "この音声の話者感情を詳細に分析し、以下の指定JSONスキーマに完全に準拠して返してください。\n- emotions: joy, sadness, anger, fear, surprise, disgust, neutral を 0.0～1.0\n- valence(-1..+1), arousal(0..1), dominance(0..1)\n- 音響上の傾向も可能なら出す（speaking_rate, avg_pitch_hz など）\n- notes は日本語で短く根拠を書く",
					},
					{
						"type":       "input_audio",
						"data":       audioB64,
						"media_type": mimeType,
					},
				},
			},
		},
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "EmotionResult",
				"strict": true,
				"schema": schema,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("リクエストの構築に失敗: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	// リクエストを送信
	fmt.Println("OpenAI APIに音声データを送信中...")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストに失敗: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスを読み込む
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み込みに失敗: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API エラー (ステータス %d): %s", resp.StatusCode, string(body))
	}

	// MessageResponse をパース
	var msgResp MessageResponse
	if err := json.Unmarshal(body, &msgResp); err != nil {
		return nil, fmt.Errorf("レスポンスのパースに失敗: %w", err)
	}

	// Content から JSON テキストを抽出
	if len(msgResp.Content) == 0 {
		return nil, fmt.Errorf("レスポンスにコンテンツがありません")
	}

	jsonStr := msgResp.Content[0].Text

	// JSONを抽出してパース
	fullJSON := extractJSON(jsonStr)
	var result EmotionResult
	if err := json.Unmarshal([]byte(fullJSON), &result); err != nil {
		return nil, fmt.Errorf("結果JSONのパースに失敗: %w", err)
	}

	return &result, nil
}

func getSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"version":   map[string]any{"type": "string"},
			"model":     map[string]any{"type": "string"},
			"language":  map[string]any{"type": "string"},
			"emotions":  map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "number"}},
			"valence":   map[string]any{"type": "number"},
			"arousal":   map[string]any{"type": "number"},
			"dominance": map[string]any{"type": "number"},
			"notes":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"speaking_rate": map[string]any{"type": "number"},
			"avg_pitch_hz":  map[string]any{"type": "number"},
			"energy_proxy":  map[string]any{"type": "number"},
			"silence_ratio": map[string]any{"type": "number"},
		},
		"required": []string{"version", "model", "language", "emotions", "valence", "arousal", "dominance"},
		"additionalProperties": false,
	}
}

// extractJSON は、APIからのテキストレスポンスに含まれるJSON部分を抽出します。
func extractJSON(raw string) string {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start != -1 && end != -1 && start < end {
		return raw[start : end+1]
	}
	return raw
}
