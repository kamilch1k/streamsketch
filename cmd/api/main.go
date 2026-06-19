package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kamilch1k/streamsketch/internal/httpapi"
	"github.com/kamilch1k/streamsketch/internal/sketch"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	precision := flag.Int("precision", 12, "HyperLogLog precision from 4 to 18")
	width := flag.Int("width", 2048, "count-min sketch width")
	depth := flag.Int("depth", 5, "count-min sketch depth")
	topK := flag.Int("top-k", 5, "default heavy hitter count")
	flag.Parse()

	handler, err := httpapi.NewHandler(sketch.Config{
		Precision: uint8(*precision),
		Width:     *width,
		Depth:     *depth,
		TopK:      *topK,
	})
	if err != nil {
		slog.Error("invalid sketch config", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		slog.Info("streamsketch api listening", "addr", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}
}
