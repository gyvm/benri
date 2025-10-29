package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// .envファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		log.Println("警告: .envファイルが見つかりません。")
	}

	// コマンドライン引数を定義する
	output := flag.String("o", "", "出力ファイル名（マークダウン形式）")
	lang := flag.String("lang", "auto", "音声の言語（ja, en, または auto）")
	flag.Parse()

	// 音声ファイルパスを取得する
	if flag.NArg() == 0 {
		fmt.Println("エラー: 文字起こしする音声ファイルを指定してください。")
		fmt.Println("使用法: whisper [オプション] <音声ファイルパス>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	audioFilepath := flag.Arg(0)

	// 出力ファイル名が指定されていない場合は、入力ファイル名から生成する
	if *output == "" {
		base := filepath.Base(audioFilepath)
		ext := filepath.Ext(base)
		*output = strings.TrimSuffix(base, ext) + ".md"
	}

	// APIキーの存在チェック
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("エラー: 環境変数 OPENAI_API_KEY が設定されていません。")
	}

	// 音声ファイルの存在チェック
	if _, err := os.Stat(audioFilepath); os.IsNotExist(err) {
		log.Fatalf("エラー: 音声ファイル '%s' が見つかりません。", audioFilepath)
	}

	// Whisper APIへのリクエスト処理
	ctx := context.Background()
	client := openai.NewClient(apiKey)
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioFilepath,
	}

	// "auto" 以外の場合のみ言語を指定
	if *lang != "auto" {
		req.Language = *lang
	}

	fmt.Printf("音声ファイル '%s' の文字起こしを開始します...\n", audioFilepath)
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		log.Fatalf("文字起こしに失敗しました: %v\n", err)
	}

	fmt.Println("文字起こしが完了しました。")

	// 結果をファイルに書き込む
	err = os.WriteFile(*output, []byte(resp.Text), 0644)
	if err != nil {
		log.Fatalf("ファイル '%s' への書き込みに失敗しました: %v", *output, err)
	}

	fmt.Printf("結果を '%s' に書き込みました。\n", *output)
}
