# Goルーチン動作検証サンプル集

このディレクトリは、Go言語の並行処理機能である「ゴルーチン」の様々な動作パターンを学習・検証するためのサンプルコード集です。

## 各サンプルの概要

### 1. `01_waitgroup.go`：基本編 `sync.WaitGroup`

`sync.WaitGroup` を使い、複数のゴルーチンの処理がすべて完了するまで待機する最も基本的な同期パターンです。

**実行方法:**
```bash
go run goroutine-examples/01_waitgroup.go
```

### 2. `02_channel.go`：チャネル編

チャネル（`chan`）を使って、ゴルーチン間で安全にデータを送受信するプロデューサー・コンシューマーモデルを実装しています。

**実行方法:**
```bash
go run goroutine-examples/02_channel.go
```

### 3. `03_nested.go`：実践編① 入れ子と実行時間の差

ゴルーチンの中からさらに新しいゴルーチンを起動する「入れ子」構造のサンプルです。各処理の実行時間をランダムにすることで、非同期処理の完了順序が予測できないことを示します。

**実行方法:**
```bash
go run goroutine-examples/03_nested.go
```

### 4. `04_worker_pool.go`：実践編② ワーカープール

実際のアプリケーションでよく利用されるパターンです。限られた数のワーカーゴルーチンを起動し、大量のタスクを効率的に処理します。

**実行方法:**
```bash
go run goroutine-examples/04_worker_pool.go
```

### 5. `05_race_condition.go`：競合編 レースコンディションとMutex

複数のゴルーチンが同じ共有メモリに同時にアクセスすることで発生する「データ競合（Race Condition）」を意図的に発生させ、その問題を `sync.Mutex` を使って解決する方法を示します。

**実行方法（競合を検出する場合）:**
Goの組み込みツールであるレースディテクタを有効にして実行します。
```bash
go run -race goroutine-examples/05_race_condition.go
```
