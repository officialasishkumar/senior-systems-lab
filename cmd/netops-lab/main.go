package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/officialasishkumar/senior-systems-lab/internal/broker"
	"github.com/officialasishkumar/senior-systems-lab/internal/config"
	"github.com/officialasishkumar/senior-systems-lab/internal/logging"
	"github.com/officialasishkumar/senior-systems-lab/internal/observability"
	"github.com/officialasishkumar/senior-systems-lab/internal/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		if err := healthcheck(config.FromEnv().HTTPAddr); err != nil {
			os.Exit(1)
		}
		return
	}

	cfg := config.FromEnv()
	logger := logging.New(cfg.LogLevel)
	metrics := observability.NewMetrics()
	queue := broker.New(cfg.QueueCapacity, cfg.DeadLetterCapacity, metrics)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	handler := broker.HandlerFunc(func(ctx context.Context, msg broker.Message) error {
		logger.InfoContext(ctx, "processed message", "message_id", msg.ID, "topic", msg.Topic, "attempt", msg.Attempts)
		return nil
	})
	queue.Start(ctx, cfg.Workers, handler)

	httpSrv := server.NewHTTP(cfg.HTTPAddr, queue, metrics, logger)
	tcpSrv := server.NewTCP(cfg.TCPAddr, queue, metrics, logger)
	udpSrv := server.NewUDP(cfg.UDPAddr, metrics, logger)

	var wg sync.WaitGroup
	start := func(name string, run func(context.Context) error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("starting service", "service", name)
			if err := run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
				logger.Error("service stopped with error", "service", name, "error", err)
				stop()
			}
		}()
	}

	start("http", httpSrv.Run)
	start("tcp", tcpSrv.Run)
	start("udp", udpSrv.Run)

	<-ctx.Done()
	logger.Info("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	_ = tcpSrv.Shutdown(shutdownCtx)
	_ = udpSrv.Shutdown(shutdownCtx)
	queue.Close()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("shutdown complete")
	case <-time.After(cfg.ShutdownTimeout):
		logger.Warn("shutdown timeout elapsed")
	}
}

func healthcheck(addr string) error {
	host := addr
	if strings.HasPrefix(host, ":") {
		host = "127.0.0.1" + host
	}
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + host + "/healthz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("unhealthy")
	}
	return nil
}
