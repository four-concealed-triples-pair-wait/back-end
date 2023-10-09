package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"golang.org/x/sync/errgroup"
)

// テスト項目
// - 期待通りにHTTPサーバーが起動しているか
// - テストコードが意図通りに終了するか
func TestRun(t *testing.T) {
	// キャンセル可能な「context.Context」のオブジェクトを作る。
	// ポート番号を0に指定すると利用可能なポートを動的に選択してくれる。
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen port %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	// 別ゴルーチンでテスト対象の「run」関数を実行してHTTPサーバー起動。
	eg.Go(func() error {
		return run(ctx, l)
	})

	in := "message"
	url := fmt.Sprintf("http://%s/%s", l.Addr().String(), in)
	t.Logf("try request to %q", url)
	rsp, err := http.Get(url)

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
