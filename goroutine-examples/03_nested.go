package main

import (
	"fmt"
	"sync"
	"time"
	"math/rand"
)

// childWorker は、入れ子の内側で実行されるゴルーチンです。
func childWorker(id int, parentId int, wg *sync.WaitGroup) {
	defer wg.Done()

	// 処理時間をランダムにする
	sleepDuration := time.Duration(rand.Intn(500) + 100) * time.Millisecond

	fmt.Printf("  [子 %d-%d] 開始 (親: %d)\n", parentId, id, parentId)
	time.Sleep(sleepDuration)
	fmt.Printf("  [子 %d-%d] 完了 (親: %d, 実行時間: %v)\n", parentId, id, parentId, sleepDuration)
}

// parentWorker は、複数の子ゴルーチンを起動するゴルーチンです。
func parentWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("[親 %d] 開始\n", id)

	// 子ゴルーチンのための新しいWaitGroup
	var childWg sync.WaitGroup

	// 3つの子ゴルーチンを起動
	for i := 1; i <= 3; i++ {
		childWg.Add(1)
		go childWorker(i, id, &childWg)
	}

	// この親ゴルーチンが起動したすべての子ゴルーチンが終わるのを待つ
	childWg.Wait()

	fmt.Printf("[親 %d] 完了 (すべての子が終了)\n", id)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup

	fmt.Println("入れ子ゴルーチンのデモを開始します。")

	// 2つの親ゴルーチンを起動
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go parentWorker(i, &wg)
	}

	// すべての親ゴルーチンが終わるのを待つ
	wg.Wait()

	fmt.Println("すべての処理が完了しました。")
}
