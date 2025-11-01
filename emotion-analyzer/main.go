package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

	// 1) 音声ファイルを読み込み、Base64エンコードする
	raw, err := os.ReadFile(inFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: ファイルが読み込めませんでした: %v\n", err)
		os.Exit(1)
	}
	audioB64 := base64.StdEncoding.EncodeToString(raw)

	// 2) Realtime WebSocket 接続
	dialer := websocket.Dialer{
		TLSClientConfig:  &tls.Config{},
		HandshakeTimeout: 30 * time.Second,
	}
	url := "wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview"
	headers := map[string][]string{
		"Authorization": {"Bearer " + apiKey},
		"OpenAI-Beta":   {"realtime=v1"},
	}
	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: WebSocket接続に失敗しました: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("WebSocket接続が確立されました。")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 3) セッション初期化（構造化出力のスキーマを提示）
	schema := getSchema()
	initMsg := map[string]any{
		"type": "response.create",
		"response": map[string]any{
			"modalities":   []string{"text"}, // 返答はJSONテキスト想定
			"instructions": "次の音声の話者感情を推定し、指定のJSONスキーマにstrict準拠で返してください。\n- emotions: joy,sadness,anger,fear,surprise,disgust,neutral を 0.0〜1.0\n- valence(-1..+1), arousal(0..1), dominance(0..1)\n- 音響上の傾向も可能なら出す（speaking_rate, avg_pitch_hz など）\n- notes は日本語で短く根拠を書く",
			"response_format": map[string]any{
				"type": "json_schema",
				"json_schema": map[string]any{
					"name":   "EmotionResult",
					"strict": true,
					"schema": schema,
				},
			},
		},
	}
	if err := conn.WriteJSON(initMsg); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 初期化メッセージの送信に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("初期化メッセージを送信しました。")

	// 4) 音声ペイロード送信
	audioEvt := map[string]any{
		"type":  "input_audio_buffer.append",
		"audio": audioB64,
	}
	if err := conn.WriteJSON(audioEvt); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 音声データの送信に失敗しました: %v\n", err)
		os.Exit(1)
	}
	if err := conn.WriteJSON(map[string]any{"type": "input_audio_buffer.commit"}); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 音声コミットの送信に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("音声データを送信しました。応答を待っています...")

	// 5) 応答受信
	var result EmotionResult
	jsonStr := ""
readLoop:
	for {
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "エラー: タイムアウトしました。")
			os.Exit(1)
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: メッセージの受信に失敗しました: %v\n", err)
			os.Exit(1)
		}

		var env struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &env); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 不明なメッセージをデコードできませんでした: %v\n", err)
			continue
		}

		switch env.Type {
		case "response.text.delta":
			var delta struct {
				Delta string `json:"delta"`
			}
			if err := json.Unmarshal(msg, &delta); err == nil {
				jsonStr += delta.Delta
			}
		case "response.done", "response.completed": // 互換性のため両方見る
			fmt.Println("分析が完了しました。")
			break readLoop
		case "error":
			fmt.Fprintf(os.Stderr, "エラー: APIからエラーが返されました: %s\n", string(msg))
			os.Exit(1)
		}
	}

	// 6) JSONの抽出とパース
	fullJSON := extractJSON(jsonStr)
	if err := json.Unmarshal([]byte(fullJSON), &result); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 結果JSONのパースに失敗しました: %v\n", err)
		fmt.Fprintf(os.Stderr, "受信した文字列: %s\n", jsonStr)
		os.Exit(1)
	}

	// 7) Markdownレポートを生成してファイルに書き込む
	if err := writeMarkdownReport(result, *outFile); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: レポートの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正常に完了しました。レポートが", *outFile, "に保存されました。")
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

// extractJSON は、APIからのストリームデータに含まれるJSON部分を抽出します。
func extractJSON(raw string) string {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start != -1 && end != -1 && start < end {
		return raw[start : end+1]
	}
	return raw
}
