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

var supportedModels = []string{"gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"}

func main() {
	if err := godotenv.Load(); err != nil {
		// .env file is not required, so we don't treat this as a fatal error
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENAI_API_KEY is not set in the environment variables or .env file.")
	}

	model := flag.String("m", "gpt-4o", "Model to use for diarization")
	output := flag.String("o", "", "Output file path")
	flag.Parse()

	if !isSupportedModel(*model) {
		log.Fatalf("Error: Model %s is not supported. Supported models are: %v", *model, supportedModels)
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Usage: diarizer [options] <input_file>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	inputFile := flag.Args()[0]

	content, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Error reading input file %s: %v", inputFile, err)
	}
	inputText := string(content)

	client := openai.NewClient(apiKey)
	fmt.Println("Requesting diarization to OpenAI API...")
	result, err := diarize(client, *model, inputText)
	if err != nil {
		log.Fatalf("Diarization failed: %v", err)
	}
	fmt.Println("Diarization completed.")

	outputFile := *output
	if outputFile == "" {
		outputFile = defaultOutputPath(inputFile)
	}

	err = os.WriteFile(outputFile, []byte(result), 0644)
	if err != nil {
		log.Fatalf("Error writing to output file %s: %v", outputFile, err)
	}

	fmt.Printf("Successfully wrote diarized text to %s\n", outputFile)
}

func isSupportedModel(model string) bool {
	for _, m := range supportedModels {
		if model == m {
			return true
		}
	}
	return false
}

func defaultOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return base + ".diarized.md"
}

func diarize(client *openai.Client, model string, text string) (string, error) {
	prompt := fmt.Sprintf(
		"以下のテキストの話者分離を行ってください。"+
			"各発言が誰によってなされたかを推測し、行の先頭に `[話者A]`、`[話者B]` のようなラベルを付けてください。"+
			"会話の文脈を考慮し、自然な話者の割り当てをお願いします。\n\n---\n\n%s",
		text,
	)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "あなたは、与えられたテキストの話者分離を高い精度で行うアシスタントです。",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return resp.Choices[0].Message.Content, nil
}
