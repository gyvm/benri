package main

import (
	"fmt"
	"sync"
	"time"
)

// worker は、いくつかの処理を実行する関数です。
// 各ワーカーは、自身のIDと処理にかかる時間を表示します。
func worker(id int, wg *sync.WaitGroup) {
	// この関数が終了する際に、WaitGroupに完了を通知します。
	defer wg.Done()

	fmt.Printf("Worker %d: 処理を開始します。\n", id)
	// 処理に時間がかかることをシミュレートします。
	time.Sleep(time.Second)
	fmt.Printf("Worker %d: 処理を終了しました。\n", id)
}

func main() {
	// WaitGroupを初期化します。
	var wg sync.WaitGroup

	// 5つのゴルーチンを起動します。
	for i := 1; i <= 5; i++ {
		// WaitGroupのカウンターを1増やします。
		wg.Add(1)
		// worker関数をゴルーチンとして起動します。
		go worker(i, &wg)
	}

	fmt.Println("すべてのゴルーチンの処理が完了するのを待ちます。")
	// WaitGroupのカウンターが0になるまで、ここで処理をブロックします。
	wg.Wait()
	fmt.Println("すべてのゴルーチンの処理が完了しました。プログラムを終了します。")
}
