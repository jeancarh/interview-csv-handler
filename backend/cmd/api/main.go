package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeancarh/interview-csv-handler/backend/api"
	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/csvimport"
	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/httpapi"
	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/payment"
	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/sqlite"
	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
	"github.com/jeancarh/interview-csv-handler/backend/internal/config"
	"github.com/jeancarh/interview-csv-handler/backend/migrations"
	"github.com/jeancarh/interview-csv-handler/backend/web"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("shutting down", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	if err := sqlite.Migrate(ctx, db, migrations.Files); err != nil {
		return err
	}

	now := func() time.Time { return time.Now().UTC() }
	products := sqlite.NewProductRepository(db)
	tx := sqlite.NewTransactor(db)

	catalogService := catalog.NewService(products, now)
	importer := catalog.NewImporter(products, tx, now)
	checkoutService := checkout.NewService(products, sqlite.NewOrderRepository(db), payment.NewFake(), tx, now)

	if cfg.SeedCSV != "" {
		if err := seed(ctx, catalogService, importer, cfg.SeedCSV, cfg.MaxImportRows); err != nil {
			return fmt.Errorf("seed: %w", err)
		}
	}

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: httpapi.NewRouter(httpapi.Deps{
			Catalog:        catalogService,
			Importer:       importer,
			Checkout:       checkoutService,
			Ping:           db.PingContext,
			Assets:         web.Assets(),
			Spec:           api.Spec,
			MaxUploadBytes: cfg.MaxUploadBytes,
			MaxImportRows:  cfg.MaxImportRows,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func seed(ctx context.Context, svc *catalog.Service, importer *catalog.Importer, path string, maxRows int) error {
	page, err := svc.Search(ctx, catalog.SearchQuery{PageSize: 1})
	if err != nil {
		return err
	}
	if page.Total > 0 {
		slog.Info("seed skipped, catalog already has products", "products", page.Total)
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := csvimport.Parse(f, maxRows)
	if err != nil {
		return err
	}
	report, err := importer.Import(ctx, rows)
	if err != nil {
		return err
	}
	s := report.Summary
	slog.Info("seeded catalog", "file", path, "created", s.Created, "skipped", s.Skipped, "rejected", s.Rejected)
	return nil
}

func healthcheck() int {
	client := http.Client{Timeout: 2 * time.Second}
	res, err := client.Get("http://127.0.0.1:" + config.Load().Port + "/healthz")
	if err != nil {
		return 1
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
