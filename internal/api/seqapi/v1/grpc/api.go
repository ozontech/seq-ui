package grpc

import (
	"time"

	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type API struct {
	seqapi.UnimplementedSeqAPIServiceServer

	config              config.SeqAPI
	seqDB               seqdb.Client
	inmemWithRedisCache cache.Cache
	redisCache          cache.Cache
	nowFn               func() time.Time
	fieldsCache         *fieldsCache
	pinnedFields        []*seqapi.Field
	asyncSearches       *asyncsearches.Service
	profiles            *profiles.Profiles
}

func New(
	cfg config.SeqAPI,
	seqDB seqdb.Client,
	inmemWithRedisCache cache.Cache,
	redisCache cache.Cache,
	asyncSearches *asyncsearches.Service,
	profiles *profiles.Profiles,
) *API {
	var fCache *fieldsCache
	if cfg.FieldsCacheTTL > 0 {
		fCache = newFieldsCache(cfg.FieldsCacheTTL)
	}

	return &API{
		config:              cfg,
		seqDB:               seqDB,
		inmemWithRedisCache: inmemWithRedisCache,
		redisCache:          redisCache,
		nowFn:               time.Now,
		fieldsCache:         fCache,
		pinnedFields:        parsePinnedFields(cfg.PinnedFields),
		asyncSearches:       asyncSearches,
		profiles:            profiles,
	}
}

func parsePinnedFields(fields []config.PinnedField) []*seqapi.Field {
	res := make([]*seqapi.Field, len(fields))
	for i, f := range fields {
		res[i] = &seqapi.Field{
			Name: f.Name,
			Type: seqdb.FieldTypeToProto(f.Type),
		}
	}
	return res
}

type fieldsCache struct {
	ttl time.Duration

	ts     time.Time
	fields []*seqapi.Field
}

func newFieldsCache(ttl time.Duration) *fieldsCache {
	return &fieldsCache{ttl: ttl}
}

func (c *fieldsCache) getFields() ([]*seqapi.Field, bool, bool) {
	return c.fields, !c.ts.IsZero(), time.Since(c.ts) < c.ttl
}

func (c *fieldsCache) setFields(fields []*seqapi.Field) {
	c.fields = fields
	c.ts = time.Now()
}

func checkLimitOffset(limit, offset int32) error {
	if limit < 0 {
		return types.NewErrInvalidRequestField("'limit' must be non-negative")
	}
	if offset < 0 {
		return types.NewErrInvalidRequestField("'offset' must be non-negative")
	}
	return nil
}
