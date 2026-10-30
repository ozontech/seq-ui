package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/buildinfo"
	"github.com/ozontech/seq-ui/internal/api"
	admin_v1 "github.com/ozontech/seq-ui/internal/api/admin/v1"
	dashboards_v1 "github.com/ozontech/seq-ui/internal/api/dashboards/v1"
	errorgroups_v1 "github.com/ozontech/seq-ui/internal/api/errorgroups/v1"
	massexport_v1 "github.com/ozontech/seq-ui/internal/api/massexport/v1"
	seqapi_v1 "github.com/ozontech/seq-ui/internal/api/seqapi/v1"
	userprofile_v1 "github.com/ozontech/seq-ui/internal/api/userprofile/v1"
	"github.com/ozontech/seq-ui/internal/app/config/loader"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/server"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
	"github.com/ozontech/seq-ui/internal/pkg/service/admin"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
	"github.com/ozontech/seq-ui/internal/pkg/service/dashboards"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport"
	"github.com/ozontech/seq-ui/internal/pkg/service/profiles"
	"github.com/ozontech/seq-ui/internal/pkg/service/userprofile"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/tracing"
)

const defaultConfig = "config/config.example.yaml"

func main() {
	// Load .env file for local development (optional).
	// OS environment variables take precedence.
	_ = godotenv.Load()

	configPath := flag.String("config", defaultConfig, "application config")
	flag.Parse()

	logger.Info("hi, I am seq-ui",
		zap.String("version", buildinfo.Version),
		zap.String("build_time", buildinfo.BuildTime),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	run(ctx, *configPath)
}

func run(ctx context.Context, configPath string) {
	if configPath == defaultConfig {
		logger.Warn("app uses the default config file, to provide your own config use -config flag")
	}

	cfg, err := loader.FromFile(configPath)
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
	}

	if err := tracing.Init(); err != nil {
		logger.Error("failed to init tracing", zap.Error(err))
	}

	registrar := initApp(ctx, cfg)

	serv, err := server.New(ctx, cfg, registrar)
	if err != nil {
		logger.Fatal("failed to init servers", zap.Error(err))
	}

	// Run launches configured grpc/http servers and debug server. On successful
	// shutdown http.ErrServerClosed is returned because of http.Server.Serve.
	if err = serv.Run(ctx); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("failed to run servers", zap.Error(err))
	}
}

func initApp(ctx context.Context, cfg *config.Config) *api.Registrar {
	logger.Info("initializing caches")
	if err := cache.Init(ctx, &cfg.Cache); err != nil {
		logger.Fatal("failed to init caches", zap.Error(err))
	}

	logger.Info("initializing seq-db clients")
	seqDBClients, err := initSeqDBClients(ctx, cfg)
	if err != nil {
		logger.Fatal("failed to init seq-db clients", zap.Error(err))
	}

	logger.Info("initializing clickhouse clients")
	chClients, err := initClickHouseClients(ctx, cfg)
	if err != nil {
		logger.Fatal("failed to init clickhouse clients", zap.Error(err))
	}

	logger.Info("initializing db")
	db, err := initDb(ctx, cfg.DB)
	if err != nil {
		logger.Fatal("failed to init db", zap.Error(err))
	}

	var massExportV1 *massexport_v1.MassExport
	if cfg.Handlers.MassExport != nil {
		logger.Info("initializing mass export")
		massExportSvc, err := massexport.New(ctx, cfg.Handlers.MassExport, seqDBClients)
		if err != nil {
			logger.Fatal("failed to init mass export", zap.Error(err))
		}
		massExportV1 = massexport_v1.New(massExportSvc)
	}

	var (
		adminV1          *admin_v1.Admin
		userProfileV1    *userprofile_v1.UserProfile
		dashboardsV1     *dashboards_v1.Dashboards
		asyncSearchesSvc map[string]asyncsearches.Service
	)
	if db != nil {
		repo := repository.New(db, cfg.DB.RequestTimeout)
		userProfilesSvc := userprofile.New(repo.UserProfiles, repo.FavoriteQueries)
		dashboardsSvc := dashboards.New(repo.Dashboards)
		profiles.Init(repo.UserProfiles.GetOrCreate)

		logger.Info("initializing user profile")
		userProfileV1 = userprofile_v1.New(userProfilesSvc)
		logger.Info("initializing dashboards")
		dashboardsV1 = dashboards_v1.New(dashboardsSvc)

		if cfg.Handlers.AsyncSearch != nil {
			logger.Info("initializing async search")
			asyncSearchesSvc = asyncsearches.New(ctx, cfg.Handlers.AsyncSearch, repo.AsyncSearches, seqDBClients)
		}

		if cfg.Handlers.Admin != nil {
			logger.Info("initializing admin")
			adminSvc := admin.New(cfg.Handlers.Admin, repo.Admin, repo.UserProfiles)
			adminV1 = admin_v1.New(adminSvc)
		}
	}

	logger.Info("initializing seqapi")
	seqApiV1 := seqapi_v1.New(&cfg.Handlers.SeqAPI, seqDBClients, asyncSearchesSvc)

	var errorGroupsV1 *errorgroups_v1.ErrorGroups
	if cfg.Handlers.ErrorGroups != nil {
		logger.Info("initializing error groups")
		errorGroupsSvc := errorgroups.New(cfg.Handlers.ErrorGroups, &cfg.Clients, chClients)
		errorGroupsV1 = errorgroups_v1.New(errorGroupsSvc)
	}

	return api.NewRegistrar(adminV1, seqApiV1, userProfileV1, dashboardsV1, massExportV1, errorGroupsV1)
}

func initSeqDBClients(ctx context.Context, cfg *config.Config) (map[string]seqdb.Client, error) {
	clients := make(map[string]seqdb.Client, len(cfg.Clients.SeqDB))
	for id, clientCfg := range cfg.Clients.SeqDB {
		client, err := createSeqDBClient(ctx, clientCfg, &cfg.Handlers.SeqAPI)
		if err != nil {
			return nil, fmt.Errorf("failed to create seq-db client %q: %w", id, err)
		}
		clients[id] = client
	}

	return clients, nil
}

func createSeqDBClient(ctx context.Context, cfg *config.SeqDBClient, seqAPI *config.SeqAPI) (seqdb.Client, error) {
	var maxSearchLimit int32
	if len(seqAPI.Envs) == 0 && seqAPI.Options.SeqDBID == cfg.ID {
		maxSearchLimit = seqAPI.Options.Limits.MaxSearchLimit
	}

	// one seqdb client can be used by several envs, so it must accept the largest one
	for _, env := range seqAPI.Envs {
		if env.Options.SeqDBID == cfg.ID {
			maxSearchLimit = max(maxSearchLimit, env.Options.Limits.MaxSearchLimit)
		}
	}

	clientParams := seqdb.ClientParams{
		Addrs:          cfg.Addrs,
		Timeout:        cfg.Timeout,
		MaxRetries:     cfg.RequestRetries,
		MaxRecvMsgSize: cfg.AvgDocSize * 1024 * int(maxSearchLimit),
	}
	if cfg.RequestRetriesBackoff != nil {
		clientParams.InitialRetryBackoff = cfg.RequestRetriesBackoff.InitialInterval
		clientParams.MaxRetryBackoff = cfg.RequestRetriesBackoff.MaxInterval
	}
	if cfg.GRPCParams != nil && cfg.GRPCParams.Keepalive != nil {
		clientParams.GRPCKeepaliveParams = &seqdb.GRPCKeepaliveParams{
			Time:                cfg.GRPCParams.Keepalive.Time,
			Timeout:             cfg.GRPCParams.Keepalive.Timeout,
			PermitWithoutStream: cfg.GRPCParams.Keepalive.PermitWithoutStream,
		}
	}

	return seqdb.NewGRPCClient(ctx, clientParams)
}

func initDb(ctx context.Context, cfg *config.DB) (*pgxpool.Pool, error) {
	if cfg == nil {
		logger.Warn("db config is nil, running without db")
		return nil, nil
	}
	pgxCfg, err := pgxpool.ParseConfig(cfg.ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	if !*cfg.UsePreparedStatements {
		// By default, pgx uses the QueryExecModeCacheStatement and automatically prepares and caches prepared statements.
		// However, this may be incompatible with proxies such as PGBouncer.
		pgxCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		logger.Info("db running without prepared statements")
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	return pool, nil
}

func initClickHouseClients(ctx context.Context, cfg *config.Config) (map[string]driver.Conn, error) {
	clients := make(map[string]driver.Conn, len(cfg.Clients.ClickHouse))
	for id, clientCfg := range cfg.Clients.ClickHouse {
		client, err := createClickHouseClient(ctx, clientCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create clickhouse client %q: %w", id, err)
		}
		clients[id] = client
	}

	return clients, nil
}

func createClickHouseClient(ctx context.Context, cfg *config.CHClient) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: cfg.Addrs,
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}
