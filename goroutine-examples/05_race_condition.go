package main

import (
	"fmt"
	"sync"
)

// Counter は、ミューテックスで保護されたカウンターです。
type Counter struct {
	mu    sync.Mutex
	value int
}

// Increment は、カウンターの値を安全にインクリメントします。
func (c *Counter) Increment() {
	c.mu.Lock()
	// この関数が終了する際に、必ずアンロックされるようにします。
	defer c.mu.Unlock()
	c.value++
}

// Value は、カウンターの現在の値を安全に取得します。
func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func main() {
	var wg sync.WaitGroup

	// --- 競合状態が発生する例 ---
	var unsafeCounter int
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			// ミューテックスなしで共有変数をインクリメント
			unsafeCounter++
		}()
	}
	wg.Wait()
	// 期待値は1000ですが、競合によりそれより少ない値になる可能性が高いです。
	// 実行するたびに結果が変わることもあります。
	fmt.Printf("競合状態ありの場合の結果: %d (期待値: 1000)\n", unsafeCounter)
	fmt.Println("---------------------------------")

	// --- Mutexを使って競合状態を解決する例 ---
	safeCounter := Counter{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			safeCounter.Increment()
		}()
	}
	wg.Wait()
	// Mutexによって保護されているため、常に期待値の1000になります。
	fmt.Printf("Mutexで保護した場合の結果: %d (期待値: 1000)\n", safeCounter.Value())

	fmt.Println("\n実行方法のヒント:")
	fmt.Println("go run -race goroutine-examples/05_race_condition.go")
	fmt.Println("`-race` フラグを付けて実行すると、Goのランタイムがデータ競合を検出してくれます。")
}
