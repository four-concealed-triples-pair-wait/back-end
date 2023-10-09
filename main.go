package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
)

// プログラムのエントリーポイント。HTTPサーバを初期化して実行する。
func main() {
	if len(os.Args) != 2 {
		log.Printf("need port number\n")
		os.Exit(1)
	}
	p := os.Args[1]
	l, err := net.Listen("tcp", ":"+p)
	if err != nil {
		log.Fatalf("failed to listen port %s: %v", p, err)
	}
	// サーバを起動し、終了エラーがあればログに出力する。
	if err := run(context.Background(), l); err != nil {
		log.Printf("failed to terminate server: %v", err)
		os.Exit(1)
	}
}

// :18080ポートでリッスンするHTTPサーバーを起動する。
func run(ctx context.Context, l net.Listener) error {
	// HTTPサーバを設定し、初期化する。
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
		}),
	}

	// 引数のcontextで新しいerrgroupを作成する。
	eg, ctx := errgroup.WithContext(ctx)

	// 別ゴルーチンでHTTPサーバーを起動する
	eg.Go(func() error {
		// サーバが正常にシャットダウンした場合はエラーを返さない。
		// それ以外の場合、エラーをログに出力し、errgroupにエラーを返す。
		if err := s.Serve(l); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("failed to close: %+v", err)
			return err
		}
		return nil
	})

	// contextのキャンセルシグナルを待機する。
	<-ctx.Done()

	// サーバを正常にシャットダウン
	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown: %+v", err)
	}

	// errgroup内の全てのゴルーチンが終了するのを待機する。
	return eg.Wait()
}
