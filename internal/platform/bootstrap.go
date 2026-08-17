package platform

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw14/cry044/internal/application"
	"github.com/wyw14/cry044/internal/config"
	"github.com/wyw14/cry044/internal/domain"
	"github.com/wyw14/cry044/internal/middleware"
	"github.com/wyw14/cry044/internal/repository"
	"github.com/wyw14/cry044/internal/service"
	httptransport "github.com/wyw14/cry044/internal/transport/http"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type reviewAssembly struct {
	server *http.Server
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func Serve(settings config.Config) {
	root, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	assembly, err := assemble(root, settings)
	if err != nil {
		panic(err)
	}
	defer assembly.pool.Close()
	defer assembly.logger.Sync()
	workers, lifetime := errgroup.WithContext(root)
	workers.Go(func() error {
		err := assembly.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})
	workers.Go(func() error {
		<-lifetime.Done()
		deadline, cancel := context.WithTimeout(context.WithoutCancel(root), 10*time.Second)
		defer cancel()
		return assembly.server.Shutdown(deadline)
	})
	if err = workers.Wait(); err != nil {
		assembly.logger.Error("review runtime stopped", zap.Error(err))
	}
}

func assemble(ctx context.Context, settings config.Config) (*reviewAssembly, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, settings.DatabaseURL)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}
	ledger := repository.NewPGLedger(pool)
	clock := service.ReviewClock{}
	identities := &service.ReviewIdentitySequence{}
	exchange := service.NewLocalExchange(settings.UploadDir)
	components := httptransport.Services{
		Reviews:       application.NewReviewService(ledger, clock, identities),
		Deliberations: application.NewDeliberationService(ledger, clock, domain.PanelPolicy{MinimumReviewers: 2, DivergenceThreshold: 2}),
		Transfers:     application.NewTransferService(ledger, clock, identities, exchange, exchange),
		Ready: func() error {
			probe, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			return pool.Ping(probe)
		},
	}
	server := &http.Server{Addr: settings.Addr, Handler: httptransport.New(components, middleware.ReviewBoundary(settings.Timeout), middleware.RecoverReviewAPI(logger)), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 45 * time.Second}
	return &reviewAssembly{server: server, pool: pool, logger: logger}, nil
}
