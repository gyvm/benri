package main

import (
	"fmt"
	"time"
	"sync"
)

// worker は、タスクチャネルからタスクを受け取り、処理を実行します。
func worker(id int, tasks <-chan int, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		fmt.Printf("ワーカー %d: タスク %d を開始\n", id, task)
		// 処理に時間がかかることをシミュレート
		time.Sleep(time.Second)
		result := fmt.Sprintf("タスク %d の結果", task)
		fmt.Printf("ワーカー %d: タスク %d を完了\n", id, task)
		results <- result
	}
}

func main() {
	const numTasks = 10
	const numWorkers = 3

	tasks := make(chan int, numTasks)
	results := make(chan string, numTasks)

	var wg sync.WaitGroup

	fmt.Printf("%d個のワーカーで%d個のタスクを処理します。\n", numWorkers, numTasks)

	// ワーカーゴルーチンを起動
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, tasks, results, &wg)
	}

	// タスクをタスクチャネルに送信
	for i := 1; i <= numTasks; i++ {
		tasks <- i
	}
	close(tasks) // すべてのタスクを送信したらチャネルを閉じる

	// すべてのワーカーが完了するのを待つ
	wg.Wait()
	close(results)

	fmt.Println("\nすべてのタスクが完了しました。結果:")
	// 結果を収集
	for result := range results {
		fmt.Println(result)
	}
}
