package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/ozontech/seq-ui/internal/api"
	dashboards_v1 "github.com/ozontech/seq-ui/internal/api/dashboards/v1"
	errorgroups_v1 "github.com/ozontech/seq-ui/internal/api/errorgroups/v1"
	massexport_v1 "github.com/ozontech/seq-ui/internal/api/massexport/v1"
	"github.com/ozontech/seq-ui/internal/api/profiles"
	seqapi_v1 "github.com/ozontech/seq-ui/internal/api/seqapi/v1"
	userprofile_v1 "github.com/ozontech/seq-ui/internal/api/userprofile/v1"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/server"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
	repositorych "github.com/ozontech/seq-ui/internal/pkg/repository_ch"
	"github.com/ozontech/seq-ui/internal/pkg/service"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport"
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport/filestore"
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport/sessionstore"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/tracing"
	"go.uber.org/zap"
)

const (
	defaultConfig = "config/config.example.yaml"

	defaultClientMaxRecvMsgSize = 4 * 1024 * 1024 // 4MB
)

var (
	configPath = flag.String("config", defaultConfig, "application config")
)

func init() {
	// Load .env file for local development (optional).
	// OS environment variables take precedence.
	_ = godotenv.Load()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGTERM,
		os.Interrupt,
	)
	defer cancel()

	run(ctx)
}

func run(ctx context.Context) {
	flag.Parse()

	if *configPath == defaultConfig {
		logger.Warn("app uses the default config file, to provide your own config use -config flag")
	}

	cfg, err := config.FromFile(*configPath)
	if err != nil {
		logger.Fatal("read config file error", zap.Error(err))
	}

	if tracingCfg, err := tracing.Initialize(); err != nil {
		logger.Error("tracing initialization failed", zap.Error(err))
	} else {
		logger.Info(
			"tracing initialization success",
			zap.String("service_name", tracingCfg.ServiceName),
			zap.String("agent_host", tracingCfg.AgentHost),
			zap.String("agent_port", tracingCfg.AgentPort),
			zap.Float64("sampler_param", tracingCfg.SamplerParam))
	}

	registrar := initApp(ctx, cfg)

	serv, err := server.New(ctx, cfg.Server, registrar)
	if err != nil {
		logger.Fatal("app init error", zap.Error(err))
	}

	// Run launches both grpc and http servers. On successful
	// http.ErrServerClosed is returned because of http.Server.Serve.
	if err = serv.Run(ctx); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("app run", zap.Error(err))
	}
}

func initApp(ctx context.Context, cfg config.Config) *api.Registrar {
	logger.Info("initializing seq-db clients")
	seqDBClients, err := initSeqDBClients(ctx, cfg)
	if err != nil {
		logger.Fatal("failed to init seq-db client", zap.Error(err))
	}

	var defaultClient seqdb.Client

	if len(cfg.Handlers.SeqAPI.Envs) > 0 {
		client, exists := seqDBClients[cfg.Handlers.SeqAPI.DefaultEnv]
		if !exists {
			logger.Fatal("client for default environment not found",
				zap.String("defaultEnv", cfg.Handlers.SeqAPI.DefaultEnv),
			)
		}
		defaultClient = client
	} else {
		client, exists := seqDBClients[config.DefaultSeqDBClientID]
		if !exists {
			logger.Fatal("default client not found",
				zap.String("defaultClientID", config.DefaultSeqDBClientID),
			)
		}
		defaultClient = client
	}

	var massExportV1 *massexport_v1.MassExport
	if cfg.Handlers.MassExport != nil {
		exportServer, err := initExportService(ctx, *cfg.Handlers.MassExport, defaultClient)
		if err != nil {
			logger.Fatal("can't init export server", zap.Error(err))
		}

		massExportV1 = massexport_v1.New(exportServer)
	}

	logger.Info("initializing inmemory with redis seqapi cache")
	inmemWithRedisCache, err := cache.NewInmemoryWithRedisOrInmemory(ctx, cfg.Server.Cache)
	if err != nil {
		logger.Fatal("failed to init inmemory with redis seqapi cache", zap.Error(err))
	}

	logger.Info("initializing redis seqapi cache")
	redisCache, err := cache.NewRedisOrInmemory(ctx, cfg.Server.Cache)
	if err != nil {
		logger.Fatal("failed to init redis seqapi cache", zap.Error(err))
	}

	logger.Info("initializing db")
	db, err := initDb(ctx, cfg.Server.DB)
	if err != nil {
		logger.Fatal("failed to init db", zap.Error(err))
	}

	var (
		asyncSearchesService *asyncsearches.Service
		p                    *profiles.Profiles
		userProfileV1        *userprofile_v1.UserProfile
		dashboardsV1         *dashboards_v1.Dashboards
	)
	if db != nil {
		repo := repository.New(db, cfg.Server.DB.RequestTimeout)
		svc := service.New(repo)
		p = profiles.New(svc)

		userProfileV1 = userprofile_v1.New(svc, p)
		dashboardsV1 = dashboards_v1.New(svc, p)

		asyncSearchesService = asyncsearches.New(ctx, repo, defaultClient, cfg.Handlers.AsyncSearch.AdminUsers)
	}

	seqApiV1 := seqapi_v1.New(cfg.Handlers.SeqAPI, seqDBClients, inmemWithRedisCache, redisCache, asyncSearchesService, p)

	logger.Info("initializing clickhouse")
	ch, err := initClickHouse(ctx, cfg.Server.CH)
	if err != nil {
		logger.Fatal("failed to init clickhouse", zap.Error(err))
	}

	var errorGroupsV1 *errorgroups_v1.ErrorGroups
	if ch != nil {
		repo := repositorych.New(ch, cfg.Server.CH.Sharded, cfg.Handlers.ErrorGroups.QueryFilter)
		svc := errorgroups.New(repo, cfg.Handlers.ErrorGroups.LogTagsMapping)

		errorGroupsV1 = errorgroups_v1.New(svc)
	}

	return api.NewRegistrar(seqApiV1, userProfileV1, dashboardsV1, massExportV1, errorGroupsV1)
}

func initSeqDBClients(ctx context.Context, cfg config.Config) (map[string]seqdb.Client, error) {
	clients := make(map[string]seqdb.Client)
	for _, clientCfg := range cfg.Clients.SeqDB {
		client, err := createSeqBDClient(ctx, clientCfg, cfg.Handlers.SeqAPI)
		if err != nil {
			return nil, fmt.Errorf("failed to create seq_db client %s: %w", clientCfg.ID, err)
		}
		clients[clientCfg.ID] = client
	}

	return clients, nil
}

func createSeqBDClient(ctx context.Context, cfg config.SeqDBClient, seqAPI config.SeqAPI) (seqdb.Client, error) {
	clientMaxRecvMsgSize := cfg.AvgDocSize * 1024 * int(seqAPI.MaxSearchLimit)
	if clientMaxRecvMsgSize < defaultClientMaxRecvMsgSize {
		clientMaxRecvMsgSize = defaultClientMaxRecvMsgSize
	}

	clientParams := seqdb.ClientParams{
		Addrs:               cfg.Addrs,
		Timeout:             cfg.Timeout,
		MaxRetries:          cfg.RequestRetries,
		InitialRetryBackoff: cfg.InitialRetryBackoff,
		MaxRetryBackoff:     cfg.MaxRetryBackoff,
		MaxRecvMsgSize:      clientMaxRecvMsgSize,
	}
	if cfg.GRPCKeepaliveParams != nil {
		clientParams.GRPCKeepaliveParams = &seqdb.GRPCKeepaliveParams{
			Time:                cfg.GRPCKeepaliveParams.Time,
			Timeout:             cfg.GRPCKeepaliveParams.Timeout,
			PermitWithoutStream: cfg.GRPCKeepaliveParams.PermitWithoutStream,
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
		return nil, fmt.Errorf("can't parse connection string: %w", err)
	}

	if !*cfg.UsePreparedStatements {
		// By default, pgx uses the QueryExecModeCacheStatement and automatically prepares and caches prepared statements.
		// However, this may be incompatible with proxies such as PGBouncer.
		pgxCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		logger.Info("db running without prepared statements")
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("can't create pgx pool: %w", err)
	}

	return pool, nil
}

func initExportService(ctx context.Context, cfg config.MassExport, client seqdb.Client) (massexport.Service, error) {
	sessionStore, err := sessionstore.NewRedisSessionStore(ctx, cfg.SessionStore)
	if err != nil {
		return nil, fmt.Errorf("init session store: %w", err)
	}
	logger.Info("session store initialized")

	fileStore, err := filestore.NewS3(cfg.FileStore.S3)
	if err != nil {
		return nil, fmt.Errorf("init file store: %w", err)
	}
	logger.Info("file store initialized")

	return massexport.NewService(ctx, cfg, sessionStore, fileStore, client)
}

func initClickHouse(ctx context.Context, cfg *config.CH) (driver.Conn, error) {
	if cfg == nil {
		logger.Warn("clickhouse config is nil, running without clickhouse")
		return nil, nil
	}

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
