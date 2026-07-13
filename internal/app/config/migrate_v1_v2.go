package config

import (
	v1 "github.com/ozontech/seq-ui/internal/app/config/v1"
	v2 "github.com/ozontech/seq-ui/internal/app/config/v2"
)

func migrateV1ToV2(src v1.Config) v2.Config {
	dst := v2.Config{Version: currentVersion}

	dst.Server = migrateServer(src.Server)
	dst.Cache = migrateCache(src.Server)
	dst.Clients = migrateClients(src)
	dst.DB = migrateDB(src.Server)
	dst.Handlers = migrateHandlers(src.Handlers, &dst)

	return dst
}

func migrateServer(src *v1.Server) *v2.Server {
	dst := &v2.Server{
		HTTP: v2.HTTP{
			Addr:              src.HTTPAddr,
			ReadTimeout:       src.HTTPReadTimeout,
			ReadHeaderTimeout: src.HTTPReadHeaderTimeout,
			WriteTimeout:      src.HTTPWriteTimeout,
			CORS:              migrateCORS(src.CORS),
		},
		GRPC: v2.GRPC{
			Addr:              src.GRPCAddr,
			ConnectionTimeout: src.GRPCConnectionTimeout,
		},
		Debug: v2.Debug{
			Addr: src.DebugAddr,
		},
		RateLimiters: migrateApiRateLimiters(src.RateLimiters),
	}

	if src.OIDC != nil {
		dst.Auth = &v2.Auth{
			OIDC: &v2.OIDC{
				SkipVerify:     src.OIDC.SkipVerify,
				AuthURLs:       src.OIDC.AuthURLs,
				RootCA:         src.OIDC.RootCA,
				CACert:         src.OIDC.CACert,
				PrivateKey:     src.OIDC.PrivateKey,
				SSLSkipVerify:  src.OIDC.SSLSkipVerify,
				AllowedClients: src.OIDC.AllowedClients,
				CacheSecretKey: src.OIDC.CacheSecretKey,
			},
		}
	}

	if src.JWTSecretKey != "" {
		dst.Auth.JWT = &v2.JWT{SecretKey: src.JWTSecretKey}
	}

	return dst
}

func migrateCORS(src *v1.CORS) *v2.CORS {
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

func migrateApiRateLimiters(src v1.ApiToRateLimiters) v2.ApiToRateLimiters {
	if src == nil {
		return nil
	}

	dst := make(v2.ApiToRateLimiters, len(src))
	for api, rl := range src {
		dst[api] = v2.ApiRateLimiters{
			Default: v2.RateLimiter{
				RatePerSec:   rl.Default.RatePerSec,
				MaxBurst:     rl.Default.MaxBurst,
				StoreMaxKeys: rl.Default.StoreMaxKeys,
				PerHandler:   rl.Default.PerHandler,
			},
			SpecialUsers: migrateUserToRateLimiter(rl.SpecialUsers),
		}
	}

	return dst
}

func migrateUserToRateLimiter(src v1.UserToRateLimiter) v2.UserToRateLimiter {
	if src == nil {
		return nil
	}

	dst := make(v2.UserToRateLimiter, len(src))
	for k, rl := range src {
		dst[k] = v2.RateLimiter{
			RatePerSec:   rl.RatePerSec,
			MaxBurst:     rl.MaxBurst,
			StoreMaxKeys: rl.StoreMaxKeys,
			PerHandler:   rl.PerHandler,
		}
	}

	return dst
}

func migrateGRPCKeepaliveParams(src *v1.GRPCKeepaliveParams) *v2.GRPCKeepaliveParams {
	if src == nil {
		return nil
	}

	return &v2.GRPCKeepaliveParams{
		Time:                src.Time,
		Timeout:             src.Timeout,
		PermitWithoutStream: src.PermitWithoutStream,
	}
}

func migrateClients(src v1.Config) *v2.Clients {
	dst := &v2.Clients{}

	if src.Clients != nil {
		if len(src.Clients.SeqDB) > 0 {
			dst.SeqDB = make([]v2.SeqDBClient, 0, len(src.Clients.SeqDB))
			for _, s := range src.Clients.SeqDB {
				dst.SeqDB = append(dst.SeqDB, migrateSeqDBClient(s))
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
				GRPCKeepaliveParams: migrateGRPCKeepaliveParams(src.Clients.GRPCKeepaliveParams),
			}}
		}
	}

	if src.Server != nil && src.Server.CH != nil {
		ch := src.Server.CH
		dst.ClickHouse = []v2.CHClient{{
			ID:          v2.DefaultSeqDBClientID,
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

func migrateSeqDBClient(src v1.SeqDBClient) v2.SeqDBClient {
	return v2.SeqDBClient{
		ID:                  src.ID,
		Timeout:             src.Timeout,
		AvgDocSize:          src.AvgDocSize,
		Addrs:               src.Addrs,
		RequestRetries:      src.RequestRetries,
		InitialRetryBackoff: src.InitialRetryBackoff,
		MaxRetryBackoff:     src.MaxRetryBackoff,
		ClientMode:          src.ClientMode,
		GRPCKeepaliveParams: migrateGRPCKeepaliveParams(src.GRPCKeepaliveParams),
	}
}

func migrateDB(src *v1.Server) *v2.DB {
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

func migrateCache(src *v1.Server) *v2.Cache {
	if src == nil {
		return &v2.Cache{}
	}

	cache := &v2.Cache{}
	cache.Inmemory = append(cache.Inmemory, v2.InmemoryCache{
		ID:          v2.DefaultInmemCacheID,
		NumCounters: src.Cache.Inmemory.NumCounters,
		MaxCost:     src.Cache.Inmemory.MaxCost,
		BufferItems: src.Cache.Inmemory.BufferItems,
	})

	if src.Cache.Redis != nil {
		cache.Redis = append(cache.Redis, migrateRedis(src.Cache.Redis, v2.DefaultRedisID, v2.DefaultInmemCacheID))
	}

	return cache
}

func migrateRedis(src *v1.Redis, id, withInmemID string) v2.Redis {
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
	}
}

func migrateHandlers(src *v1.Handlers, cfg *v2.Config) *v2.Handlers {
	if src == nil {
		return &v2.Handlers{}
	}

	dst := &v2.Handlers{
		SeqAPI:      migrateSeqAPI(src.SeqAPI),
		ErrorGroups: migrateErrorGroups(src.ErrorGroups),
		AsyncSearch: migrateAsyncSearch(src.AsyncSearch),
	}

	if src.MassExport != nil {
		dst.MassExport = migrateMassExport(src.MassExport, cfg)
	}

	return dst
}

func migrateSeqAPI(src v1.SeqAPI) v2.SeqAPI {
	return v2.SeqAPI{
		SeqAPIOptions: migrateSeqAPIOptions(src.SeqAPIOptions),
		Envs:          migrateSeqAPIEnvs(src.Envs),
		DefaultEnv:    src.DefaultEnv,
	}
}

func migrateSeqAPIEnvs(envs map[string]v1.SeqAPIEnv) map[string]v2.SeqAPIEnv {
	if len(envs) == 0 {
		return nil
	}

	dst := make(map[string]v2.SeqAPIEnv, len(envs))
	for name, cfg := range envs {
		dst[name] = v2.SeqAPIEnv{
			SeqDB:   cfg.SeqDB,
			Options: migrateSeqAPIOptions(cfg.Options),
		}
	}

	return dst
}

func migrateSeqAPIOptions(options *v1.SeqAPIOptions) *v2.SeqAPIOptions {
	if options == nil {
		return nil
	}

	return &v2.SeqAPIOptions{
		MaxSearchLimit:             options.MaxSearchLimit,
		MaxSearchTotalLimit:        options.MaxSearchTotalLimit,
		MaxSearchOffsetLimit:       options.MaxSearchOffsetLimit,
		MaxExportLimit:             options.MaxExportLimit,
		SeqCLIMaxSearchLimit:       options.SeqCLIMaxSearchLimit,
		MaxParallelExportRequests:  options.MaxParallelExportRequests,
		MaxAggregationsPerRequest:  options.MaxAggregationsPerRequest,
		MaxBucketsPerAggregationTs: options.MaxBucketsPerAggregationTs,
		EventsCacheTTL:             options.EventsCacheTTL,
		PinnedFields:               migrateFields(options.PinnedFields),
		SystemFields:               migrateFields(options.SystemFields),
		LogsLifespanCacheKey:       options.LogsLifespanCacheKey,
		LogsLifespanCacheTTL:       options.LogsLifespanCacheTTL,
		FieldsCacheTTL:             options.FieldsCacheTTL,
		Masking:                    migrateMasking(options.Masking),
	}
}

func migrateFields(fs []v1.Field) []v2.Field {
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

func migrateMasking(src *v1.Masking) *v2.Masking {
	if src == nil {
		return nil
	}

	return &v2.Masking{
		Masks:         migrateMasks(src.Masks),
		ProcessFields: src.ProcessFields,
		IgnoreFields:  src.IgnoreFields,
	}
}

func migrateMasks(ms []v1.Mask) []v2.Mask {
	if ms == nil {
		return nil
	}

	dst := make([]v2.Mask, len(ms))
	for i, m := range ms {
		dst[i] = v2.Mask{
			Re:            m.Re,
			Groups:        m.Groups,
			Mode:          m.Mode,
			ReplaceWord:   m.ReplaceWord,
			ProcessFields: m.ProcessFields,
			IgnoreFields:  m.IgnoreFields,
			FieldFilters:  migrateFieldFilters(m.FieldFilters),
		}
	}

	return dst
}

func migrateFieldFilters(src *v1.FieldFilterSet) *v2.FieldFilterSet {
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

func migrateErrorGroups(eg v1.ErrorGroups) v2.ErrorGroups {
	return v2.ErrorGroups{
		LogTagsMapping: v2.LogTagsMapping{
			Env:     eg.LogTagsMapping.Env,
			Service: eg.LogTagsMapping.Service,
			Release: eg.LogTagsMapping.Release,
		},
		QueryFilter: eg.QueryFilter,
	}
}

func migrateAsyncSearch(a v1.AsyncSearch) v2.AsyncSearch {
	return v2.AsyncSearch{
		AdminUsers:           a.AdminUsers,
		ListQueryLengthLimit: a.ListQueryLengthLimit,
	}
}

func migrateMassExport(me *v1.MassExport, cfg *v2.Config) *v2.MassExport {
	dst := &v2.MassExport{
		BatchSize:        me.BatchSize,
		WorkersCount:     me.WorkersCount,
		TasksChannelSize: me.TasksChannelSize,
		PartLength:       me.PartLength,
		URLPrefix:        me.URLPrefix,
		AllowedUsers:     me.AllowedUsers,
		FileStore:        migrateFileStore(me.FileStore),
		DownloadParams:   migrateDownloadParams(me.SeqProxyDownloader),
	}

	if me.SessionStore != nil {
		cfg.Cache.Redis = append(cfg.Cache.Redis, migrateRedis(&me.SessionStore.Redis, v2.DefaultMassExportRedisID, ""))

		dst.SessionStore = &v2.SessionStore{
			RedisID:        v2.DefaultMassExportRedisID,
			ExportLifetime: me.SessionStore.ExportLifetime,
		}
	}

	return dst
}

func migrateFileStore(fs *v1.FileStore) *v2.FileStore {
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

func migrateDownloadParams(src *v1.SeqProxyDownloader) *v2.DownloadParams {
	if src == nil {
		return nil
	}

	return &v2.DownloadParams{
		Delay:               src.Delay,
		InitialRetryBackoff: src.InitialRetryBackoff,
		MaxRetryBackoff:     src.MaxRetryBackoff,
	}
}
