package seqdb

import (
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	mock "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type streamErrorType int8

const (
	streamErrNo streamErrorType = iota
	streamErrProxy
	streamErrRecv
	streamErrConvert
)

func initGRPCClient(client *mock.MockSeqProxyApiClient) *GRPCClient {
	return &GRPCClient{
		clients: []seqproxyapi.SeqProxyApiClient{client},
		timeout: 3 * time.Second,
	}
}

func makeProxySearchQuery(query string, from, to *timestamppb.Timestamp) *seqproxyapi.SearchQuery {
	return &seqproxyapi.SearchQuery{
		Query: query,
		From:  from,
		To:    to,
	}
}

func makeEvent(id string, countData int, t *timestamppb.Timestamp) *seqapi.Event {
	e := &seqapi.Event{
		Id:   id,
		Data: make(map[string]string),
		Time: t,
	}
	for i := 0; i < countData; i++ {
		e.Data[fmt.Sprintf("field%d", i+1)] = fmt.Sprintf("val%d", i+1)
	}
	return e
}

func makeHistogram(bucketCount int) *seqapi.Histogram {
	hist := &seqapi.Histogram{
		Buckets: make([]*seqapi.Histogram_Bucket, 0, bucketCount),
	}
	for i := 0; i < bucketCount; i++ {
		hist.Buckets = append(hist.Buckets, &seqapi.Histogram_Bucket{
			Key:      uint64(i * 100),
			DocCount: uint64(i + 1),
		})
	}
	return hist
}

type makeAggOpts struct {
	NotExists int64
	Quantiles []float64
}

func makeAggregation(bucketCount int, opts *makeAggOpts) *seqapi.Aggregation {
	agg := &seqapi.Aggregation{
		Buckets: make([]*seqapi.Aggregation_Bucket, 0, bucketCount),
	}
	for i := 0; i < bucketCount; i++ {
		v := new(float64)
		*v = float64(i + 1)
		b := &seqapi.Aggregation_Bucket{
			Key:   fmt.Sprintf("test%d", i+1),
			Value: v,
		}
		if opts != nil {
			b.NotExists = opts.NotExists
			if len(opts.Quantiles) > 0 {
				b.Quantiles = opts.Quantiles
			}
		}
		agg.Buckets = append(agg.Buckets, b)
	}
	return agg
}

func makeAggregations(aggCount, bucketCount int, opts *makeAggOpts) []*seqapi.Aggregation {
	aggs := make([]*seqapi.Aggregation, 0, aggCount)
	for i := 0; i < aggCount; i++ {
		aggs = append(aggs, makeAggregation(bucketCount, opts))
	}
	return aggs
}

func checkEventsEqual(event1, event2 *seqapi.Event) bool {
	if event1 == nil && event2 == nil {
		return true
	}
	if event1.Id != event2.Id {
		return false
	}
	if len(event1.Data) != len(event2.Data) {
		return false
	}
	for k, v := range event1.Data {
		if event2.Data[k] != fmt.Sprintf("%q", v) {
			return false
		}
	}
	return true
}
