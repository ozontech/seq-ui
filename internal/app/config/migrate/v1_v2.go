package migrate

import (
	v1 "github.com/ozontech/seq-ui/internal/app/config/v1"
	v2 "github.com/ozontech/seq-ui/internal/app/config/v2"
)

type v1tov2 struct {
	src v1.Config
}

func V1ToV2(src v1.Config) v2.Config {
	m := v1tov2{src: src}
	return m.migrate()
}

func (m *v1tov2) migrate() v2.Config {
	dst := v2.Config{Version: 2}

	dst.Server = m.migrateServer(m.src.Server)
	dst.Cache = m.migrateCache(m.src.Server)
	dst.Clients = m.migrateClients(m.src)
	dst.DB = m.migrateDB(m.src.Server)
	dst.Handlers = m.migrateHandlers(m.src.Handlers, &dst, m.src.Server.Cache)

	return dst
}

func (m *v1tov2) migrateServer(src *v1.Server) *v2.Server {
	dst := &v2.Server{
		HTTP: v2.HTTP{
			Addr:              src.HTTPAddr,
			ReadTimeout:       src.HTTPReadTimeout,
			ReadHeaderTimeout: src.HTTPReadHeaderTimeout,
			WriteTimeout:      src.HTTPWriteTimeout,
			CORS:              m.migrateCORS(src.CORS),
		},
		GRPC: v2.GRPC{
			Addr:              src.GRPCAddr,
			ConnectionTimeout: src.GRPCConnectionTimeout,
		},
		Debug: v2.Debug{
			Addr: src.DebugAddr,
		},
		RateLimiters: m.migrateApiRateLimiters(src.RateLimiters),
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
		}

		dst.Auth = &v2.Auth{OIDC: oidc}
	}

	if src.JWTSecretKey != "" {
		if dst.Auth == nil {
			dst.Auth = &v2.Auth{}
		}
		dst.Auth.JWT = &v2.JWT{SecretKey: src.JWTSecretKey}
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

func (m *v1tov2) migrateApiRateLimiters(src v1.ApiToRateLimiters) v2.ApiToRateLimiters {
	if src == nil {
		return nil
	}

	dst := make(v2.ApiToRateLimiters, len(src))
	for api, rl := range src {
		dst[api] = v2.ApiRateLimiters{
			Default:      m.migrateRateLimiter(rl.Default),
			SpecialUsers: m.migrateUserToRateLimiter(rl.SpecialUsers),
		}
	}

	return dst
}

func (m *v1tov2) migrateUserToRateLimiter(src v1.UserToRateLimiter) v2.UserToRateLimiter {
	if src == nil {
		return nil
	}

	dst := make(v2.UserToRateLimiter, len(src))
	for k, rl := range src {
		dst[k] = m.migrateRateLimiter(rl)
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

func (m *v1tov2) migrateClients(src v1.Config) *v2.Clients {
	dst := &v2.Clients{}

	if src.Clients != nil {
		if len(src.Clients.SeqDB) > 0 {
			dst.SeqDB = make([]v2.SeqDBClient, 0, len(src.Clients.SeqDB))
			for i := range src.Clients.SeqDB {
				dst.SeqDB = append(dst.SeqDB, m.migrateSeqDBClient(&src.Clients.SeqDB[i]))
			}
		} else {
			dst.SeqDB = []v2.SeqDBClient{{
				ID:                  v2.DefaultSeqDBClientID,
				Timeout:             src.Clients.SeqDBTimeout,
				AvgDocSize:          src.Clients.SeqDBAvgDocSize,
				Addrs:               src.Clients.SeqDBAddrs,
				RequestRetries:      src.Clients.RequestRetries,
				InitialRetryBackoff: src.Clients.InitialRetryBackoff,
				MaxRetryBackoff:     src.Clients.MaxRetryBackoff,
				ClientMode:          src.Clients.ProxyClientMode,
				GRPCParams:          m.migrateGRPCParams(src.Clients.GRPCKeepaliveParams),
			}}
		}
	}

	if src.Server != nil && src.Server.CH != nil {
		ch := src.Server.CH
		dst.ClickHouse = []v2.CHClient{{
			ID:          v2.DefaultCHClientID,
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

func (m *v1tov2) migrateSeqDBClient(src *v1.SeqDBClient) v2.SeqDBClient {
	return v2.SeqDBClient{
		ID:                  src.ID,
		Timeout:             src.Timeout,
		AvgDocSize:          src.AvgDocSize,
		Addrs:               src.Addrs,
		RequestRetries:      src.RequestRetries,
		InitialRetryBackoff: src.InitialRetryBackoff,
		MaxRetryBackoff:     src.MaxRetryBackoff,
		ClientMode:          src.ClientMode,
		GRPCParams:          m.migrateGRPCParams(src.GRPCKeepaliveParams),
	}
}

func (m *v1tov2) migrateDB(src *v1.Server) *v2.DB {
	if src == nil || src.DB == nil {
		return nil
	}

	return &v2.DB{
		Name:                   src.DB.Name,
		Host:                   src.DB.Host,
		Port:                   src.DB.Port,
		Pass:                   src.DB.Pass,
		User:                   src.DB.User,
		RequestTimeout:         src.DB.RequestTimeout,
		ConnectionPoolCapacity: src.DB.ConnectionPoolCapacity,
		UsePreparedStatements:  src.DB.UsePreparedStatements,
	}
}

func (m *v1tov2) migrateCache(src *v1.Server) *v2.Cache {
	if src == nil {
		return &v2.Cache{}
	}

	cache := &v2.Cache{}

	if !m.isInmemEmpty(src.Cache.Inmemory) {
		cache.Inmemory = append(cache.Inmemory, v2.InmemoryCache{
			ID:          v2.DefaultInmemCacheID,
			NumCounters: src.Cache.Inmemory.NumCounters,
			MaxCost:     src.Cache.Inmemory.MaxCost,
			BufferItems: src.Cache.Inmemory.BufferItems,
		})
	}

	if src.Cache.Redis != nil {
		if len(cache.Inmemory) > 0 {
			cache.Redis = append(cache.Redis,
				m.migrateRedis(src.Cache.Redis, v2.DefaultRedisID, v2.DefaultInmemCacheID),
				m.migrateRedis(src.Cache.Redis, v2.DefaultRedis2ID, ""),
			)
		} else {
			cache.Redis = append(cache.Redis, m.migrateRedis(src.Cache.Redis, v2.DefaultRedisID, ""))
		}
	}

	return cache
}

func (m *v1tov2) migrateRedis(src *v1.Redis, id, withInmemID string) v2.Redis {
	return v2.Redis{
		ID:              id,
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

func (m *v1tov2) migrateHandlers(src *v1.Handlers, cfg *v2.Config, v1Cache v1.Cache) *v2.Handlers {
	if src == nil {
		return &v2.Handlers{}
	}

	dst := &v2.Handlers{
		SeqAPI:      m.migrateSeqAPI(src.SeqAPI, v1Cache),
		ErrorGroups: m.migrateErrorGroups(src.ErrorGroups),
		AsyncSearch: m.migrateAsyncSearch(src.AsyncSearch),
		Admin:       m.migrateAdmin(src.Admin, v1Cache),
	}

	if src.MassExport != nil {
		dst.MassExport = m.migrateMassExport(src.MassExport, cfg)
	}

	return dst
}

func (m *v1tov2) migrateSeqAPI(src v1.SeqAPI, v1Cache v1.Cache) v2.SeqAPI {
	cacheID, redisID := m.defaultCacheIDs(v1Cache)
	dst := v2.SeqAPI{
		Envs:       m.migrateSeqAPIEnvs(src.Envs, cacheID, redisID),
		DefaultEnv: src.DefaultEnv,
		Options: v2.SeqAPIOptions{
			Caches: v2.SeqAPICaches{
				Events: v2.SeqAPICache{
					ID:  cacheID,
					TTL: src.EventsCacheTTL,
				},
				LogsLifespan: v2.SeqAPICache{
					ID:  redisID,
					TTL: src.LogsLifespanCacheTTL,
				},
				Fields: v2.SeqAPICache{
					ID:  cacheID,
					TTL: src.FieldsCacheTTL,
				},
			},
		},
	}

	if len(src.Envs) == 0 {
		dst.SeqDBID = v2.DefaultSeqDBClientID
	}

	if src.SeqAPIOptions != nil {
		dst.Options = m.migrateSeqAPIOptions(src.SeqAPIOptions, cacheID, redisID)
	}

	return dst
}

func (m *v1tov2) migrateSeqAPIEnvs(envs map[string]v1.SeqAPIEnv, cacheID, redisID string) map[string]v2.SeqAPIEnv {
	if len(envs) == 0 {
		return nil
	}

	dst := make(map[string]v2.SeqAPIEnv, len(envs))
	for name, cfg := range envs {
		env := v2.SeqAPIEnv{
			SeqDBID: cfg.SeqDB,
		}
		if cfg.Options != nil {
			envOptions := m.migrateSeqAPIOptions(cfg.Options, cacheID, redisID)
			env.Options = &envOptions
		}

		dst[name] = env
	}

	return dst
}

func (m *v1tov2) migrateSeqAPIOptions(options *v1.SeqAPIOptions, cacheID, redisID string) v2.SeqAPIOptions {
	return v2.SeqAPIOptions{
		Limits: v2.SeqAPILimits{
			MaxSearchLimit:             options.MaxSearchLimit,
			MaxSearchTotal:             options.MaxSearchTotalLimit,
			MaxSearchOffset:            options.MaxSearchOffsetLimit,
			MaxExportLimit:             options.MaxExportLimit,
			SeqCLIMaxSearchLimit:       options.SeqCLIMaxSearchLimit,
			MaxParallelExportRequests:  options.MaxParallelExportRequests,
			MaxAggregationsPerRequest:  options.MaxAggregationsPerRequest,
			MaxBucketsPerAggregationTs: options.MaxBucketsPerAggregationTs,
		},
		Masking:      m.migrateMasking(options.Masking),
		PinnedFields: m.migrateFields(options.PinnedFields),
		SystemFields: m.migrateFields(options.SystemFields),
		Caches: v2.SeqAPICaches{
			Events: v2.SeqAPICache{
				ID:  cacheID,
				TTL: options.EventsCacheTTL,
			},
			LogsLifespan: v2.SeqAPICache{
				ID:  redisID,
				TTL: options.LogsLifespanCacheTTL,
			},
			Fields: v2.SeqAPICache{
				ID:  cacheID,
				TTL: options.FieldsCacheTTL,
			},
		},
	}
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
		m := &ms[i]
		dst[i] = v2.Mask{
			Re:            m.Re,
			Groups:        m.Groups,
			Mode:          m.Mode,
			ReplaceWord:   m.ReplaceWord,
			ProcessFields: m.ProcessFields,
			IgnoreFields:  m.IgnoreFields,
			FieldFilters:  m.migrateFieldFilters(m.FieldFilters),
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

func (m *v1tov2) migrateErrorGroups(eg v1.ErrorGroups) *v2.ErrorGroups {
	if m.isEgEmpty(eg) {
		return nil
	}

	return &v2.ErrorGroups{
		CHID: v2.DefaultCHClientID,
		LogTagsMapping: v2.LogTagsMapping{
			Env:     eg.LogTagsMapping.Env,
			Service: eg.LogTagsMapping.Service,
			Release: eg.LogTagsMapping.Release,
		},
		QueryFilter: eg.QueryFilter,
	}
}

func (m *v1tov2) migrateAsyncSearch(a v1.AsyncSearch) *v2.AsyncSearch {
	if len(a.AdminUsers) == 0 && a.ListQueryLengthLimit == 0 {
		return nil
	}

	return &v2.AsyncSearch{
		SeqDBID:              v2.DefaultSeqDBClientID,
		AdminUsers:           a.AdminUsers,
		ListQueryLengthLimit: a.ListQueryLengthLimit,
	}
}

func (m *v1tov2) migrateMassExport(me *v1.MassExport, cfg *v2.Config) *v2.MassExport {
	dst := &v2.MassExport{
		SeqDBID:          v2.DefaultSeqDBClientID,
		BatchSize:        me.BatchSize,
		WorkersCount:     me.WorkersCount,
		TasksChannelSize: me.TasksChannelSize,
		PartLength:       me.PartLength,
		URLPrefix:        me.URLPrefix,
		AllowedUsers:     me.AllowedUsers,
		FileStore:        m.migrateFileStore(me.FileStore),
		DownloadParams:   m.migrateDownloadParams(me.SeqProxyDownloader),
	}

	if me.SessionStore != nil {
		cfg.Cache.Redis = append(cfg.Cache.Redis, m.migrateRedis(&me.SessionStore.Redis, v2.DefaultMassExportRedisID, ""))

		dst.SessionStore = &v2.SessionStore{
			RedisID:        v2.DefaultMassExportRedisID,
			ExportLifetime: me.SessionStore.ExportLifetime,
		}
	}

	return dst
}

func (m *v1tov2) migrateFileStore(fs *v1.FileStore) *v2.FileStore {
	if fs == nil {
		return nil
	}

	dst := &v2.FileStore{}
	if fs.S3 != nil {
		dst.S3 = &v2.S3{
			Endpoint:        fs.S3.Endpoint,
			AccessKeyID:     fs.S3.AccessKeyID,
			SecretAccessKey: fs.S3.SecretAccessKey,
			BucketName:      fs.S3.BucketName,
			EnableSSl:       fs.S3.EnableSSl,
		}
	}

	return dst
}

func (m *v1tov2) migrateDownloadParams(src *v1.SeqProxyDownloader) *v2.DownloadParams {
	if src == nil {
		return nil
	}

	return &v2.DownloadParams{
		Delay:               src.Delay,
		InitialRetryBackoff: src.InitialRetryBackoff,
		MaxRetryBackoff:     src.MaxRetryBackoff,
	}
}

func (m *v1tov2) migrateAdmin(src *v1.Admin, v1Cache v1.Cache) *v2.Admin {
	if src == nil {
		return nil
	}

	_, redisID := m.defaultCacheIDs(v1Cache)
	return &v2.Admin{
		SuperUsers: src.SuperUsers,
		Options: v2.HandlerCache{
			ID:  redisID,
			TTL: src.CacheTTL,
		},
	}
}

func (m *v1tov2) defaultCacheIDs(src v1.Cache) (cacheID, redisID string) {
	hasInmem := !m.isInmemEmpty(src.Inmemory)
	hasRedis := src.Redis != nil

	switch {
	case hasInmem && hasRedis:
		return v2.DefaultRedisID, v2.DefaultRedis2ID
	case hasRedis:
		return v2.DefaultRedisID, v2.DefaultRedisID
	case hasInmem:
		return v2.DefaultInmemCacheID, ""
	default:
		return "", ""
	}
}

func (m *v1tov2) isInmemEmpty(src v1.InmemoryCache) bool {
	return src.BufferItems == 0 && src.MaxCost == 0 && src.NumCounters == 0
}

func (m *v1tov2) isEgEmpty(src v1.ErrorGroups) bool {
	return len(src.LogTagsMapping.Env) == 0 && len(src.LogTagsMapping.Service) == 0 &&
		len(src.LogTagsMapping.Release) == 0 && len(src.QueryFilter) == 0
}
