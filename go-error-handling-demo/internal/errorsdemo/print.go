package errorsdemo

import (
	"errors"
	"fmt"
	"go-error-demo/internal/repository"

	"go.uber.org/multierr"
)

// DumpError は、エラーに関する詳細情報をコンソールに出力します。
// 様々なエラーハンドリング戦略による出力の違いをデモンストレーションします。
func DumpError(title string, err error) {
	fmt.Printf("=== %s ===\n", title)

	// --- 1. 標準的なエラー出力 (`%v`) ---
	// これはユーザー向けのシンプルなエラーメッセージを表示します。`err.Error()` の呼び出しチェーンをたどります。
	fmt.Println("\n--- `%v` による出力 (標準エラーメッセージ) ---")
	fmt.Printf("%v\n", err)

	// --- 2. スタックトレース付きの詳細出力 (`%+v`) ---
	// `%+v` というフォーマット指定子は、`pkg/errors` や `cockroachdb/errors` のようなライブラリが
	// スタックトレースを出力するために使用する規約です。`%w` を使った標準の `fmt.Errorf` は
	// スタックトレースを生成せず、ネストされたエラーメッセージのみを生成します。
	// これが観察すべき重要な違いです。
	fmt.Println("\n--- `%+v` による出力 (スタックトレース付き詳細ビュー) ---")
	fmt.Printf("%+v\n", err)

	// --- 3. `errors.Is` のデモンストレーション ---
	// `errors.Is` はエラーチェーンをたどり、チェーン内のいずれかのエラーが特定のエラーインスタンスと
	// 一致するかどうかを確認します。これはセンチネルエラーをチェックする現代的な方法です。
	fmt.Println("\n--- `errors.Is` のデモンストレーション ---")
	if errors.Is(err, repository.ErrNotFound) {
		fmt.Println("✅ このエラーチェーンには `repository.ErrNotFound` が含まれています。")
	} else {
		fmt.Println("❌ このエラーチェーンには `repository.ErrNotFound` は含まれていません。")
	}

	// --- 4. `multierr` のデモンストレーション ---
	// `multierr.Errors` ヘルパー関数は、エラーが `multierr` インスタンスであれば、
	// そこに含まれる個々のエラーのスライスを返します。そうでない場合は、元のエラーのみを含む
	// 長さ1のスライスを返します。この挙動を利用して、複数のエラーが存在するかを判定します。
	fmt.Println("\n--- `multierr` のデモンストレーション (`multierr.Errors`) ---")
	errs := multierr.Errors(err)
	if len(errs) > 1 {
		fmt.Printf("✅ このエラーは %d 個のエラーを含む `multierr` です:\n", len(errs))
		for i, e := range errs {
			fmt.Printf("  - エラー %d: %v\n", i+1, e)
		}
	} else {
		fmt.Println("❌ このエラーは `multierr` ではありません。")
	}

	// --- 5. `errors.As` のデモンストレーション ---
	// `errors.As` はエラーチェーンをたどり、特定のインターフェースまたは型に一致するエラーを探します。
	// これにより、カスタムエラー型が持つ追加情報にアクセスできます。
	fmt.Println("\n--- `errors.As` のデモンストレーション ---")
	var permErr *repository.PermissionError
	if errors.As(err, &permErr) {
		fmt.Printf("✅ このエラーは `PermissionError` 型です。UserID: %d, Action: %s\n", permErr.UserID, permErr.Action)
	} else {
		fmt.Println("❌ このエラーチェーンには `PermissionError` 型のエラーは含まれていません。")
	}

	fmt.Printf("======================================================================\n\n")
}
