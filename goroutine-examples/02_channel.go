package main

import (
	"fmt"
	"time"
)

// producer は、チャネルにデータを送信するゴルーチンです。
func producer(ch chan<- int) {
	for i := 1; i <= 5; i++ {
		fmt.Printf("送信: %d\n", i)
		ch <- i // データをチャネルに送信
		time.Sleep(500 * time.Millisecond)
	}
	close(ch) // すべてのデータを送信したらチャネルを閉じる
}

// consumer は、チャネルからデータを受信するゴルーチンです。
func consumer(ch <-chan int, done chan<- bool) {
	// チャネルが閉じるまで、データを受信し続けます。
	for v := range ch {
		fmt.Printf("受信: %d\n", v)
		time.Sleep(1 * time.Second)
	}
	done <- true // 処理が完了したことを通知
}

func main() {
	// int型のデータを送受信するチャネルを作成します。
	dataChannel := make(chan int)
	// 処理の完了を待つためのチャネルを作成します。
	doneChannel := make(chan bool)

	fmt.Println("プロデューサー・コンシューマーモデルを開始します。")

	// データを生成して送信するプロデューサーをゴルーチンとして起動します。
	go producer(dataChannel)
	// データを受信して処理するコンシューマーをゴルーチンとして起動します。
	go consumer(dataChannel, doneChannel)

	// コンシューマーの処理が完了するのを待ちます。
	<-doneChannel

	fmt.Println("すべての処理が完了しました。")
}
