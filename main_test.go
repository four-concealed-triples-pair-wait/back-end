package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// テスト項目
// - 期待通りにHTTPサーバーが起動しているか
// - テストコードが意図通りに終了するか
func TestRun(t *testing.T) {
	// キャンセル可能な「context.Context」のオブジェクトを作る。
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	// 別ゴルーチンでテスト対象の「run」関数を実行してHTTPサーバー起動。
	eg.Go(func() error {
		return run(ctx)
	})

	// Small delay to ensure the HTTP server is up and running
	time.Sleep(1 * time.Second)

	in := "message"
	rsp, err := http.Get("http://localhost:18080/" + in)
	if err != nil {
		t.Fatalf("failed to get: %+v", err) // Use Fatalf to stop the test immediately
	}
	defer rsp.Body.Close()

	got, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	// HTTPサーバーの戻り値を検証
	want := fmt.Sprintf("Hello, %s!", in)
	if string(got) != want {
		t.Errorf("want %q, but got %q", want, got)
	}

	// run関数に終了通知を送信
	cancel()
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
