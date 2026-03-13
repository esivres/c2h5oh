package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	zeebepb "github.com/camunda/zeebe/clients/go/v8/pkg/pb"
	pb "github.com/esivres/c2h5oh/api/proto"
	"github.com/esivres/c2h5oh/pkg/gateway"
	"github.com/esivres/c2h5oh/pkg/processing/behavior"
	"github.com/esivres/c2h5oh/pkg/processing/engine"
	"github.com/esivres/c2h5oh/pkg/processing/export"
	sqlstore "github.com/esivres/c2h5oh/pkg/processing/storage/sql"
	_ "google.golang.org/grpc/encoding/gzip"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	_ "modernc.org/sqlite"
)

func main() {
	var (
		grpcAddr    = flag.String("addr", ":9090", "gRPC listen address")
		dbPath      = flag.String("db", "c2h5oh.db", "SQLite database path")
		logLevel    = flag.String("log-level", "info", "log level (debug, info, warn, error)")
		zeebeCompat = flag.Bool("zeebe-compat", true, "enable Zeebe Gateway API compatibility proxy")
	)
	flag.Parse()

	var level slog.Level
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	if err := run(logger, *grpcAddr, *dbPath, *zeebeCompat); err != nil {
		logger.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, grpcAddr, dbPath string, zeebeCompat bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := sqlstore.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	logger.Info("database ready", "path", dbPath)

	store := sqlstore.NewStore(db)

	// Restore key generator sequence from the database to avoid key collisions after restart.
	maxKey, err := store.GetMaxKeyId(ctx)
	if err != nil {
		return fmt.Errorf("get max key: %w", err)
	}
	// Strip partition prefix (high 8 bits) to get the raw sequence number.
	startSequence := maxKey & 0x00FFFFFFFFFFFFFF
	if startSequence > 0 {
		logger.Info("restored key sequence", "maxKey", maxKey, "sequence", startSequence)
	}

	jobNotifier := gateway.NewJobNotifier()
	deployNotifier := gateway.NewDeployNotifier()
	completionNotifier := gateway.NewCompletionNotifier()

	registry := behavior.DefaultRegistry()
	logExporter := export.NewLogExporter(logger.With("component", "export"))
	processor := engine.NewProcessor(engine.Config{
		PartitionId:   1,
		Store:         store,
		Registry:      registry,
		Exporter:      logExporter,
		StartSequence: startSequence,
		Observers: []engine.IntentObserver{
			gateway.JobCreatedObserver(jobNotifier),
			gateway.DeployedObserver(deployNotifier),
			gateway.ProcessCompletedObserver(completionNotifier),
		},
		Logger: logger,
	})

	procErr := make(chan error, 1)
	go func() {
		procErr <- processor.RunWithTimerChecker(ctx, &engine.TimerCheckerConfig{
			Interval: time.Second,
		})
	}()
	logger.Info("processor started")

	srv := grpc.NewServer()
	gw := gateway.NewServer(processor, store, jobNotifier, deployNotifier, completionNotifier)
	pb.RegisterGatewayAPIServer(srv, gw)

	if zeebeCompat {
		zp := gateway.NewZeebeProxy(processor, store, jobNotifier, deployNotifier, completionNotifier)
		zeebepb.RegisterGatewayServer(srv, zp)
		logger.Info("zeebe compatibility proxy enabled")
	}

	reflection.Register(srv)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", grpcAddr, err)
	}

	grpcErr := make(chan error, 1)
	go func() {
		logger.Info("gRPC server listening", "addr", grpcAddr)
		grpcErr <- srv.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down...")
	case err := <-procErr:
		return fmt.Errorf("processor: %w", err)
	case err := <-grpcErr:
		return fmt.Errorf("grpc: %w", err)
	}

	srv.GracefulStop()
	logger.Info("stopped")
	return nil
}
