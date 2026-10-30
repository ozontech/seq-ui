package migrate

import (
	"time"

	v1 "github.com/ozontech/seq-ui/internal/app/config/v1"
	v2 "github.com/ozontech/seq-ui/internal/app/config/v2"
)

const (
	v1DefaultEventsCacheTTL       = 24 * time.Hour
	v1DefaultLogsLifespanCacheTTL = 10 * time.Minute
	v1DefaultAdminCacheTTL        = time.Hour
)

type v1tov2 struct {
	src v1.Config
}

func V1ToV2(src v1.Config) v2.Config {
	m := v1tov2{src: src}
	return m.migrate()
}

func (m *v1tov2) migrate() v2.Config {
	server := m.src.Server
	if server == nil {
		// in v1 are always enabled
		server = &v1.Server{}
	}

	handlers := m.src.Handlers
	if handlers == nil {
		// in v1 are always enabled
		handlers = &v1.Handlers{}
	}

	var v1SessionStore *v1.SessionStore
	if handlers.MassExport != nil {
		v1SessionStore = handlers.MassExport.SessionStore
	}

	return v2.Config{
		Version:  2,
		Server:   m.migrateServer(server),
		Cache:    m.migrateCache(server.Cache, v1SessionStore),
		Clients:  m.migrateClients(m.src.Clients, server.CH),
		DB:       m.migrateDB(server.DB),
		Handlers: m.migrateHandlers(handlers, server),
	}
}

func (m *v1tov2) migrateServer(src *v1.Server) v2.Server {
	dst := v2.Server{
		HTTP: &v2.HTTP{
			Addr:              src.HTTPAddr,
			ReadTimeout:       src.HTTPReadTimeout,
			ReadHeaderTimeout: src.HTTPReadHeaderTimeout,
			WriteTimeout:      src.HTTPWriteTimeout,
			CORS:              m.migrateCORS(src.CORS),
		},
		GRPC: &v2.GRPC{
			Addr:              src.GRPCAddr,
			ConnectionTimeout: src.GRPCConnectionTimeout,
		},
		Debug: v2.Debug{
			Addr:              src.DebugAddr,
			ReadTimeout:       src.HTTPReadTimeout,
			ReadHeaderTimeout: src.HTTPReadHeaderTimeout,
			WriteTimeout:      src.HTTPWriteTimeout,
			CORS:              m.migrateCORS(src.CORS),
		},
		RateLimiting: m.migrateRateLimiting(src.RateLimiters),
	}

	if src.OIDC != nil {
		oidc := &v2.OIDC{
			SkipVerify:     src.OIDC.SkipVerify,
			AuthURLs:       src.OIDC.AuthURLs,
			TLS:            m.migrateTLS(src.OIDC),
			AllowedClients: src.OIDC.AllowedClients,
		}

		if src.OIDC.CacheSecretKey != "" {
			cacheID, _ := m.defaultCacheIDs(src.Cache)
			oidc.CacheID = cacheID
			oidc.CacheSecretKey = src.OIDC.CacheSecretKey
		}

		dst.Auth = &v2.Auth{Options: v2.AuthOptions{OIDC: oidc}}
	}

	if src.JWTSecretKey != "" {
		if dst.Auth == nil {
			dst.Auth = &v2.Auth{}
		}
		dst.Auth.Options.JWT = &v2.JWT{SecretKey: src.JWTSecretKey}
	}

	return dst
}

func (m *v1tov2) migrateTLS(src *v1.OIDC) *v2.TLS {
	if src.RootCA == "" && src.CACert == "" &&
		src.PrivateKey == "" && !src.SSLSkipVerify {
		return nil
	}

	return &v2.TLS{
		CACert:     src.RootCA,
		ClientCert: src.CACert,
		ClientKey:  src.PrivateKey,
		Insecure:   src.SSLSkipVerify,
	}
}

func (m *v1tov2) migrateCORS(src *v1.CORS) *v2.CORS {
	if src == nil {
		return nil
	}

	return &v2.CORS{
		AllowedOrigins:     src.AllowedOrigins,
		AllowedMethods:     src.AllowedMethods,
		AllowedHeaders:     src.AllowedHeaders,
		ExposedHeaders:     src.ExposedHeaders,
		AllowCredentials:   src.AllowCredentials,
		MaxAge:             src.MaxAge,
		OptionsPassthrough: src.OptionsPassthrough,
	}
}

func (m *v1tov2) migrateRateLimiting(src v1.ApiToRateLimiters) *v2.RateLimiting {
	if len(src) == 0 {
		return nil
	}

	options := make(v2.ApiToRateLimiters, len(src))
	for api, rl := range src {
		options[api] = v2.RateLimiters{
			Default:      m.migrateRateLimiter(rl.Default),
			SpecialUsers: m.migrateUserToRateLimiter(rl.SpecialUsers),
		}
	}

	return &v2.RateLimiting{Options: options}
}

func (m *v1tov2) migrateUserToRateLimiter(src v1.UserToRateLimiter) map[string]v2.RateLimiter {
	if src == nil {
		return nil
	}

	dst := make(map[string]v2.RateLimiter, len(src))
	for user, rl := range src {
		dst[user] = m.migrateRateLimiter(rl)
	}

	return dst
}

func (m *v1tov2) migrateRateLimiter(src v1.RateLimiter) v2.RateLimiter {
	return v2.RateLimiter{
		RatePerSec:   src.RatePerSec,
		MaxBurst:     src.MaxBurst,
		StoreMaxKeys: src.StoreMaxKeys,
		PerHandler:   src.PerHandler,
	}
}

func (m *v1tov2) migrateGRPCParams(src *v1.GRPCKeepaliveParams) *v2.GRPCParams {
	if src == nil {
		return nil
	}

	return &v2.GRPCParams{
		Keepalive: &v2.KeepaliveParams{
			Time:                src.Time,
			Timeout:             src.Timeout,
			PermitWithoutStream: src.PermitWithoutStream,
		},
	}
}

func (m *v1tov2) migrateClients(src *v1.Clients, ch *v1.CH) v2.Clients {
	dst := v2.Clients{}

	if src != nil {
		if len(src.SeqDB) > 0 {
			dst.SeqDB = make(map[string]*v2.SeqDBClient, len(src.SeqDB))
			for i := range src.SeqDB {
				dst.SeqDB[src.SeqDB[i].ID] = m.migrateSeqDBClient(&src.SeqDB[i])
			}
		} else {
			dst.SeqDB = map[string]*v2.SeqDBClient{v2.DefaultSeqDBClientID: {
				Timeout:               src.SeqDBTimeout,
				AvgDocSize:            src.SeqDBAvgDocSize,
				Addrs:                 src.SeqDBAddrs,
				RequestRetries:        src.RequestRetries,
				RequestRetriesBackoff: m.migrateBackoff(src.InitialRetryBackoff, src.MaxRetryBackoff),
				ClientMode:            src.ProxyClientMode,
				GRPCParams:            m.migrateGRPCParams(src.GRPCKeepaliveParams),
			}}
		}
	}

	if ch != nil {
		dst.ClickHouse = map[string]*v2.CHClient{v2.DefaultCHClientID: {
			Addrs:       ch.Addrs,
			Database:    ch.Database,
			Username:    ch.Username,
			Password:    ch.Password,
			Sharded:     ch.Sharded,
			DialTimeout: ch.DialTimeout,
			ReadTimeout: ch.ReadTimeout,
		}}
	}

	return dst
}

func (m *v1tov2) migrateBackoff(initialInterval, maxInterval time.Duration) *v2.Backoff {
	if initialInterval == 0 && maxInterval == 0 {
		return nil
	}

	return &v2.Backoff{
		InitialInterval: initialInterval,
		MaxInterval:     maxInterval,
	}
}

func (m *v1tov2) migrateSeqDBClient(src *v1.SeqDBClient) *v2.SeqDBClient {
	return &v2.SeqDBClient{
		Timeout:               src.Timeout,
		AvgDocSize:            src.AvgDocSize,
		Addrs:                 src.Addrs,
		RequestRetries:        src.RequestRetries,
		RequestRetriesBackoff: m.migrateBackoff(src.InitialRetryBackoff, src.MaxRetryBackoff),
		ClientMode:            src.ClientMode,
		GRPCParams:            m.migrateGRPCParams(src.GRPCKeepaliveParams),
	}
}

func (m *v1tov2) migrateDB(src *v1.DB) *v2.DB {
	if src == nil {
		return nil
	}

	return &v2.DB{
		Name:                   src.Name,
		Host:                   src.Host,
		Port:                   src.Port,
		User:                   src.User,
		Pass:                   src.Pass,
		RequestTimeout:         src.RequestTimeout,
		ConnectionPoolCapacity: src.ConnectionPoolCapacity,
		UsePreparedStatements:  src.UsePreparedStatements,
	}
}

func (m *v1tov2) migrateCache(src v1.Cache, sessionStore *v1.SessionStore) v2.Cache {
	// in config v1 inmemory cache is always created
	cache := v2.Cache{
		Inmemory: map[string]*v2.Inmemory{v2.DefaultInmemID: {
			NumCounters: src.Inmemory.NumCounters,
			MaxCost:     src.Inmemory.MaxCost,
			BufferItems: src.Inmemory.BufferItems,
		}},
	}

	if src.Redis != nil || sessionStore != nil {
		cache.Redis = make(map[string]*v2.Redis, 3)
	}

	if src.Redis != nil {
		cache.Redis[v2.DefaultRedisWithInmemID] = m.migrateRedis(src.Redis, v2.DefaultInmemID)
		cache.Redis[v2.DefaultRedisWithoutInmemID] = m.migrateRedis(src.Redis, "")
	}

	if sessionStore != nil {
		// in v1 session store ignored key_prefix
		redis := m.migrateRedis(&sessionStore.Redis, "")
		redis.KeyPrefix = ""
		cache.Redis[v2.DefaultMassExportRedisID] = redis
	}

	return cache
}

func (m *v1tov2) migrateRedis(src *v1.Redis, withInmemID string) *v2.Redis {
	return &v2.Redis{
		WithInmemID:     withInmemID,
		Addr:            src.Addr,
		Username:        src.Username,
		Password:        src.Password,
		Timeout:         src.Timeout,
		MaxRetries:      src.MaxRetries,
		MinRetryBackoff: src.MinRetryBackoff,
		MaxRetryBackoff: src.MaxRetryBackoff,
		KeyPrefix:       src.KeyPrefix,
	}
}

func (m *v1tov2) migrateHandlers(src *v1.Handlers, server *v1.Server) v2.Handlers {
	return v2.Handlers{
		SeqAPI:      m.migrateSeqAPI(&src.SeqAPI, server.Cache),
		AsyncSearch: m.migrateAsyncSearch(&src.AsyncSearch, &src.SeqAPI),
		ErrorGroups: m.migrateErrorGroups(&src.ErrorGroups, server.CH),
		Admin:       m.migrateAdmin(src.Admin, server.Cache),
		MassExport:  m.migrateMassExport(src.MassExport, &src.SeqAPI),
	}
}

func (m *v1tov2) migrateSeqAPI(src *v1.SeqAPI, v1Cache v1.Cache) v2.SeqAPI {
	dst := v2.SeqAPI{
		Options:    m.migrateSeqAPIOptions(src.SeqAPIOptions, v1Cache),
		DefaultEnv: src.DefaultEnv,
	}

	if len(src.Envs) == 0 {
		dst.Options.SeqDBID = v2.DefaultSeqDBClientID
		return dst
	}

	dst.Envs = make(map[string]*v2.SeqAPIOptions, len(src.Envs))
	for envName, envCfg := range src.Envs {
		// in v1 env without options uses root options, in v2 root
		// options are merged into env options, so only seq_db_id is set
		options := &v2.SeqAPIOptions{}
		if envCfg.Options != nil {
			migratedOpts := m.migrateSeqAPIOptions(envCfg.Options, v1Cache)
			options = &migratedOpts
		}
		options.SeqDBID = envCfg.SeqDB
		dst.Envs[envName] = options
	}

	return dst
}

func (m *v1tov2) migrateSeqAPIOptions(src *v1.SeqAPIOptions, v1Cache v1.Cache) v2.SeqAPIOptions {
	if src == nil {
		src = &v1.SeqAPIOptions{}
	}

	redisWithInmemID, redisOrInmemID := m.defaultCacheIDs(v1Cache)
	dst := v2.SeqAPIOptions{
		Limits: v2.SeqAPILimits{
			MaxSearchLimit:             src.MaxSearchLimit,
			MaxSearchTotal:             src.MaxSearchTotalLimit,
			MaxSearchOffset:            src.MaxSearchOffsetLimit,
			MaxExportLimit:             src.MaxExportLimit,
			SeqCLIMaxSearchLimit:       src.SeqCLIMaxSearchLimit,
			MaxParallelExportRequests:  src.MaxParallelExportRequests,
			MaxAggregationsPerRequest:  src.MaxAggregationsPerRequest,
			MaxBucketsPerAggregationTs: src.MaxBucketsPerAggregationTs,
		},
		Masking:      m.migrateMasking(src.Masking),
		PinnedFields: m.migrateFields(src.PinnedFields),
		SystemFields: m.migrateFields(src.SystemFields),
		Caches: v2.SeqAPICaches{
			Events: &v2.HandlerCache{
				ID:  redisWithInmemID,
				TTL: m.migrateTTL(src.EventsCacheTTL, v1DefaultEventsCacheTTL),
			},
			LogsLifespan: &v2.HandlerCache{
				ID:  redisOrInmemID,
				TTL: m.migrateTTL(src.LogsLifespanCacheTTL, v1DefaultLogsLifespanCacheTTL),
			},
		},
	}

	// in v1 fields are cached only if ttl is set
	if src.FieldsCacheTTL > 0 {
		dst.Caches.Fields = &v2.HandlerCache{
			ID:  v2.DefaultInmemID,
			TTL: src.FieldsCacheTTL,
		}
	}

	return dst
}

func (m *v1tov2) migrateFields(fs []v1.Field) []v2.Field {
	if fs == nil {
		return nil
	}

	dst := make([]v2.Field, len(fs))
	for i, f := range fs {
		dst[i] = v2.Field{
			Name: f.Name,
			Type: f.Type,
		}
	}

	return dst
}

func (m *v1tov2) migrateMasking(src *v1.Masking) *v2.Masking {
	if src == nil {
		return nil
	}

	return &v2.Masking{
		Masks:         m.migrateMasks(src.Masks),
		ProcessFields: src.ProcessFields,
		IgnoreFields:  src.IgnoreFields,
	}
}

func (m *v1tov2) migrateMasks(ms []v1.Mask) []v2.Mask {
	if ms == nil {
		return nil
	}

	dst := make([]v2.Mask, len(ms))
	for i := range ms {
		mask := &ms[i]
		dst[i] = v2.Mask{
			Re:            mask.Re,
			Groups:        mask.Groups,
			Mode:          mask.Mode,
			ReplaceWord:   mask.ReplaceWord,
			ProcessFields: mask.ProcessFields,
			IgnoreFields:  mask.IgnoreFields,
			FieldFilters:  m.migrateFieldFilters(mask.FieldFilters),
		}
	}

	return dst
}

func (m *v1tov2) migrateFieldFilters(src *v1.FieldFilterSet) *v2.FieldFilterSet {
	if src == nil {
		return nil
	}

	dst := &v2.FieldFilterSet{
		Condition: src.Condition,
	}
	if src.Filters != nil {
		dst.Filters = make([]v2.FieldFilter, len(src.Filters))
		for i, f := range src.Filters {
			dst.Filters[i] = v2.FieldFilter{
				Field:  f.Field,
				Mode:   f.Mode,
				Values: f.Values,
			}
		}
	}

	return dst
}

func (m *v1tov2) migrateErrorGroups(src *v1.ErrorGroups, ch *v1.CH) *v2.ErrorGroups {
	if ch == nil {
		return nil
	}

	return &v2.ErrorGroups{
		Options: v2.ErrorGroupsOptions{
			CHID: v2.DefaultCHClientID,
			LogTagsMapping: v2.LogTagsMapping{
				Env:     src.LogTagsMapping.Env,
				Service: src.LogTagsMapping.Service,
				Release: src.LogTagsMapping.Release,
			},
			QueryFilter: src.QueryFilter,
		},
	}
}

func (m *v1tov2) migrateAsyncSearch(src *v1.AsyncSearch, seqAPI *v1.SeqAPI) *v2.AsyncSearch {
	seqDBID := v2.DefaultSeqDBClientID
	if len(seqAPI.Envs) > 0 {
		seqDBID = seqAPI.Envs[seqAPI.DefaultEnv].SeqDB
	}

	return &v2.AsyncSearch{
		Options: v2.AsyncSearchOptions{
			SeqDBID:              seqDBID,
			AdminUsers:           src.AdminUsers,
			ListQueryLengthLimit: src.ListQueryLengthLimit,
		},
	}
}

func (m *v1tov2) migrateMassExport(src *v1.MassExport, seqAPI *v1.SeqAPI) *v2.MassExport {
	if src == nil {
		return nil
	}

	options := v2.MassExportOptions{
		WorkersCount:     src.WorkersCount,
		TasksChannelSize: src.TasksChannelSize,
		PartLength:       src.PartLength,
		URLPrefix:        src.URLPrefix,
		AllowedUsers:     src.AllowedUsers,
		DownloadParams: v2.DownloadParams{
			BatchSize: src.BatchSize,
		},
	}

	if src.FileStore != nil && src.FileStore.S3 != nil {
		options.FileStore.S3 = &v2.S3{
			Endpoint:        src.FileStore.S3.Endpoint,
			AccessKeyID:     src.FileStore.S3.AccessKeyID,
			SecretAccessKey: src.FileStore.S3.SecretAccessKey,
			BucketName:      src.FileStore.S3.BucketName,
			EnableSSL:       src.FileStore.S3.EnableSSl,
		}
	}

	if src.SessionStore != nil {
		options.SessionStore = v2.SessionStore{
			Lifetime:     src.SessionStore.ExportLifetime,
			RedisCacheID: v2.DefaultMassExportRedisID,
		}
	}

	if src.SeqProxyDownloader != nil {
		options.DownloadParams.Delay = src.SeqProxyDownloader.Delay
		options.DownloadParams.DelayBackoff = v2.Backoff{
			InitialInterval: src.SeqProxyDownloader.InitialRetryBackoff,
			MaxInterval:     src.SeqProxyDownloader.MaxRetryBackoff,
		}
	}

	options.SeqDBID = v2.DefaultSeqDBClientID
	if len(seqAPI.Envs) > 0 {
		options.SeqDBID = seqAPI.Envs[seqAPI.DefaultEnv].SeqDB
	}

	return &v2.MassExport{Options: options}
}

func (m *v1tov2) migrateAdmin(src *v1.Admin, v1Cache v1.Cache) *v2.Admin {
	if src == nil {
		return nil
	}

	dst := &v2.Admin{
		Options: v2.AdminOptions{
			SuperUsers: src.SuperUsers,
		},
	}

	if v1Cache.Redis != nil {
		dst.Options.Cache = &v2.HandlerCache{
			ID:  v2.DefaultRedisWithoutInmemID,
			TTL: m.migrateTTL(src.CacheTTL, v1DefaultAdminCacheTTL),
		}
	}

	return dst
}

func (m *v1tov2) migrateTTL(ttl time.Duration, defaultTTL time.Duration) time.Duration {
	if ttl <= 0 {
		return defaultTTL
	}
	return ttl
}

// defaultCacheIDs returns cache ids same as in v1.
// inmemWithRedisID - inmemory+redis or inmemory (v1 events, oidc)
// redisOrInmemID - redis or inmemory (v1 logslifespan)
func (m *v1tov2) defaultCacheIDs(src v1.Cache) (inmemWithRedisID, redisOrInmemID string) {
	if src.Redis != nil {
		return v2.DefaultRedisWithInmemID, v2.DefaultRedisWithoutInmemID
	}
	return v2.DefaultInmemID, v2.DefaultInmemID
}
