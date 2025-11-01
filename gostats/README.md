# gostats

`gostats`は、Goプロジェクトのコード統計情報を表示するためのCLIツールです。Railsの`rake stats`に似た機能を提供し、パッケージごとのファイル数、コード行数、関数、構造体の数を集計して、整形された表形式で出力します。

## 機能

-   指定されたディレクトリ内のGoソースコードを再帰的に解析
-   パッケージ（ディレクトリ）ごとに統計情報を集計
-   以下の情報を表示:
    -   ファイル数
    -   総行数 (空行・コメント含む)
    -   コード行数 (LOC)
    -   関数 (Function) の数
    -   構造体 (Struct) の数
    -   1ファイルあたりの平均コード行数
-   テストファイル (`_test.go`) を集計に含めるかどうかのオプション

## インストール

```bash
go install github.com/jules/gostats@latest
```

## 使い方

```bash
gostats [options] [path]
```

-   `path`: 解析したいGoプロジェクトのルートディレクトリへのパス。（デフォルト: カレントディレクトリ）

### オプション

-   `-tests`: テストファイル (`_test.go`) を統計情報に含めます。（デフォルト: `false`）
-   `-exclude`: 統計から除外するディレクトリ名をカンマ区切りで指定します（例: `vendor,generated`、デフォルト: `vendor`）。
-   `-depth`: 統計情報の表示階層を制限します。例: `-depth 2` で2階層まで集計し、それより深いディレクトリは親にまとめます（デフォルト: 無制限）。

### 使用例

現在のディレクトリの統計情報を表示:

```bash
gostats
```

特定のディレクトリ (`~/src/my-project`) の統計情報を、テストファイルも含めて表示:

```bash
gostats -tests ~/src/my-project
```

## 出力例

```
Name                 Files   Lines   LOC   Funcs   Structs   Avg LOC/File
pkg/controller       10      500     350   15      5         35.0
pkg/model            5       250     180   5       10        36.0
main.go              1       100     80    1       0         80.0

Total                16      850     610   21      15        38.1
```
