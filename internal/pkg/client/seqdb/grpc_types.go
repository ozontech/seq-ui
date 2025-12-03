package seqdb

import (
	"math"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type proxyError seqproxyapi.Error

func (p *proxyError) toProto() *seqapi.Error {
	if p == nil {
		return nil
	}

	var code seqapi.ErrorCode
	switch p.Code {
	case seqproxyapi.ErrorCode_ERROR_CODE_NO:
		code = seqapi.ErrorCode_ERROR_CODE_NO
	case seqproxyapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE:
		code = seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE
	case seqproxyapi.ErrorCode_ERROR_CODE_TOO_MANY_FRACTIONS_HIT:
		code = seqapi.ErrorCode_ERROR_CODE_TOO_MANY_FRACTIONS_HIT
	default:
		code = seqapi.ErrorCode_ERROR_CODE_UNSPECIFIED
	}

	return &seqapi.Error{
		Code:    code,
		Message: p.Message,
	}
}

type proxyDoc seqproxyapi.Document

func (p *proxyDoc) toProto(event *seqapi.Event) error {
	event.Id = p.Id
	event.Time = p.Time
	if len(p.Data) == 0 {
		metric.SeqDBClientEmptyDataResponse.Inc()
		logger.Warn("empty response event data", zap.String("id", event.Id))
		return nil
	}
	data, err := newMapStringString(p.Data)
	if err != nil {
		return err
	}
	event.Data = data
	return nil
}

type proxyDocSlice []*seqproxyapi.Document

func (p proxyDocSlice) toProto() ([]*seqapi.Event, error) {
	events := make([]*seqapi.Event, len(p))
	buf := make([]seqapi.Event, len(p))
	for i, doc := range p {
		event := &buf[i]
		if err := (*proxyDoc)(doc).toProto(event); err != nil {
			return nil, err
		}
		events[i] = event
	}
	return events, nil
}

type proxyAggBucket seqproxyapi.Aggregation_Bucket

func (p *proxyAggBucket) toProto(b *seqapi.Aggregation_Bucket) {
	b.Key = p.Key
	b.NotExists = p.NotExists

	if !math.IsNaN(p.Value) && !math.IsInf(p.Value, 0) {
		b.Value = new(float64)
		*b.Value = p.Value
	}

	if len(p.Quantiles) > 0 {
		b.Quantiles = make([]float64, len(p.Quantiles))
		copy(b.Quantiles, p.Quantiles)
	}

	if p.Ts != nil {
		b.Ts = timestamppb.New(p.Ts.AsTime())
	}
}

type proxyAgg seqproxyapi.Aggregation

func (p *proxyAgg) toProto(agg *seqapi.Aggregation) {
	agg.Buckets = make([]*seqapi.Aggregation_Bucket, len(p.Buckets))
	buf := make([]seqapi.Aggregation_Bucket, len(p.Buckets))
	for i, bucket := range p.Buckets {
		b := &buf[i]
		(*proxyAggBucket)(bucket).toProto(b)
		agg.Buckets[i] = b
	}
	agg.NotExists = p.NotExists
}

type proxyAggSlice []*seqproxyapi.Aggregation

func (p proxyAggSlice) toProto() []*seqapi.Aggregation {
	aggs := make([]*seqapi.Aggregation, len(p))
	buf := make([]seqapi.Aggregation, len(p))
	for i, agg := range p {
		a := &buf[i]
		(*proxyAgg)(agg).toProto(a)
		aggs[i] = a
	}
	return aggs
}

type proxyHist seqproxyapi.Histogram

func (p *proxyHist) toProto() *seqapi.Histogram {
	if p == nil {
		return nil
	}

	buckets := make([]*seqapi.Histogram_Bucket, len(p.Buckets))
	buf := make([]seqapi.Histogram_Bucket, len(p.Buckets))
	for i, bucket := range p.Buckets {
		b := &buf[i]
		b.DocCount = bucket.DocCount
		b.Key = uint64(bucket.Ts.AsTime().UnixMilli())
		buckets[i] = b
	}
	return &seqapi.Histogram{
		Buckets: buckets,
	}
}

// used in conversion functions for different types of requests
type (
	seqAPISearchQ interface {
		GetQuery() string
		GetFrom() *timestamppb.Timestamp
		GetTo() *timestamppb.Timestamp
	}

	seqAPIHistQ interface {
		GetInterval() string
	}
)

func newProxySearchQuery(q seqAPISearchQ) *seqproxyapi.SearchQuery {
	return &seqproxyapi.SearchQuery{
		Query: q.GetQuery(),
		From:  q.GetFrom(),
		To:    q.GetTo(),
	}
}

func newProxyAggQuery(q *seqapi.AggregationQuery, aggQ *seqproxyapi.AggQuery) {
	aggQ.Field = q.GetField()
	aggQ.GroupBy = q.GetGroupBy()
	aggQ.Func = seqproxyapi.AggFunc(q.GetFunc()) // same enum

	if len(q.Quantiles) > 0 {
		aggQ.Quantiles = make([]float64, len(q.Quantiles))
		copy(aggQ.Quantiles, q.Quantiles)
	}

	if q.Interval != nil {
		aggQ.Interval = new(string)
		*aggQ.Interval = *q.Interval
	}
}

func newProxyAggQuerySlice(aggs []*seqapi.AggregationQuery) []*seqproxyapi.AggQuery {
	if len(aggs) == 0 {
		return nil
	}

	proxyAggs := make([]*seqproxyapi.AggQuery, len(aggs))
	buf := make([]seqproxyapi.AggQuery, len(aggs))
	for i, query := range aggs {
		q := &buf[i]
		newProxyAggQuery(query, q)
		proxyAggs[i] = q
	}
	return proxyAggs
}

func newProxyHistQuery(q seqAPIHistQ) *seqproxyapi.HistQuery {
	if q == nil || q.GetInterval() == "" {
		return nil
	}
	return &seqproxyapi.HistQuery{
		Interval: q.GetInterval(),
	}
}

func newProxyGetAggReq(req *seqapi.GetAggregationRequest) *seqproxyapi.GetAggregationRequest {
	var aggQ []*seqproxyapi.AggQuery
	// backward compatibility support with old API
	// if both old and new fields are used, new fields override old one
	if len(req.Aggregations) == 0 && req.AggField != "" {
		aggQ = append(aggQ, &seqproxyapi.AggQuery{Field: req.AggField})
	} else {
		aggQ = newProxyAggQuerySlice(req.Aggregations)
	}
	return &seqproxyapi.GetAggregationRequest{
		Query: newProxySearchQuery(req),
		Aggs:  aggQ,
	}
}

type proxyGetAggResp seqproxyapi.GetAggregationResponse

func (p *proxyGetAggResp) toProto() *seqapi.GetAggregationResponse {
	// backward compatibility support with old API
	var agg *seqapi.Aggregation
	var aggs []*seqapi.Aggregation
	if len(p.Aggs) > 0 {
		aggs = proxyAggSlice(p.Aggs).toProto()
		agg = aggs[0]
	}
	return &seqapi.GetAggregationResponse{
		Aggregation:     agg,
		Aggregations:    aggs,
		Error:           (*proxyError)(p.Error).toProto(),
		PartialResponse: p.PartialResponse,
	}
}

func newProxyGetEventReq(req *seqapi.GetEventRequest) *seqproxyapi.FetchRequest {
	return &seqproxyapi.FetchRequest{
		Ids: []string{req.Id},
	}
}

type proxyGetFieldsResp seqproxyapi.MappingResponse

func (p *proxyGetFieldsResp) toProto() (*seqapi.GetFieldsResponse, error) {
	data := make(map[string]string)
	if err := json.Unmarshal(p.Data, &data); err != nil {
		return nil, err
	}
	fields := make([]*seqapi.Field, len(data))
	buf := make([]seqapi.Field, len(data))
	i := 0
	for k, v := range data {
		field := &buf[i]
		field.Name = k
		field.Type = FieldTypeToProto(v)
		fields[i] = field
		i++
	}
	return &seqapi.GetFieldsResponse{
		Fields: fields,
	}, nil
}

func newProxyGetHistReq(req *seqapi.GetHistogramRequest) *seqproxyapi.GetHistogramRequest {
	return &seqproxyapi.GetHistogramRequest{
		Query: newProxySearchQuery(req),
		Hist:  newProxyHistQuery(req),
	}
}

type proxyGetHistResp seqproxyapi.GetHistogramResponse

func (p *proxyGetHistResp) toProto() *seqapi.GetHistogramResponse {
	return &seqapi.GetHistogramResponse{
		Histogram:       (*proxyHist)(p.Hist).toProto(),
		Error:           (*proxyError)(p.Error).toProto(),
		PartialResponse: p.PartialResponse,
	}
}

func newProxySearchReq(req *seqapi.SearchRequest) *seqproxyapi.ComplexSearchRequest {
	return &seqproxyapi.ComplexSearchRequest{
		Query:     newProxySearchQuery(req),
		Hist:      newProxyHistQuery(req.Histogram),
		Aggs:      newProxyAggQuerySlice(req.Aggregations),
		Size:      int64(req.Limit),
		Offset:    int64(req.Offset),
		WithTotal: req.WithTotal,
		Order:     seqproxyapi.Order(req.Order),
	}
}

type proxySearchResp seqproxyapi.ComplexSearchResponse

func (p *proxySearchResp) toProto() (*seqapi.SearchResponse, error) {
	docs, err := proxyDocSlice(p.Docs).toProto()
	if err != nil {
		return nil, err
	}
	return &seqapi.SearchResponse{
		Events:          docs,
		Histogram:       (*proxyHist)(p.Hist).toProto(),
		Aggregations:    proxyAggSlice(p.Aggs).toProto(),
		Total:           p.Total,
		Error:           (*proxyError)(p.Error).toProto(),
		PartialResponse: p.PartialResponse,
	}, nil
}

func newProxyExportReq(req *seqapi.ExportRequest) *seqproxyapi.ExportRequest {
	return &seqproxyapi.ExportRequest{
		Query:  newProxySearchQuery(req),
		Size:   int64(req.Limit),
		Offset: int64(req.Offset),
	}
}

func asyncSearchStatusToProto(s seqproxyapi.AsyncSearchStatus) seqapi.AsyncSearchStatus {
	switch s {
	case seqproxyapi.AsyncSearchStatus_AsyncSearchStatusInProgress:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_IN_PROGRESS
	case seqproxyapi.AsyncSearchStatus_AsyncSearchStatusDone:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE
	case seqproxyapi.AsyncSearchStatus_AsyncSearchStatusCanceled:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_CANCELED
	case seqproxyapi.AsyncSearchStatus_AsyncSearchStatusError:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_ERROR
	default:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_UNSPECIFIED
	}
}

func asyncSearchStatusFromProto(s seqapi.AsyncSearchStatus) seqproxyapi.AsyncSearchStatus {
	switch s {
	case seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_IN_PROGRESS:
		return seqproxyapi.AsyncSearchStatus_AsyncSearchStatusInProgress
	case seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE:
		return seqproxyapi.AsyncSearchStatus_AsyncSearchStatusDone
	case seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_CANCELED:
		return seqproxyapi.AsyncSearchStatus_AsyncSearchStatusCanceled
	case seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_ERROR:
		return seqproxyapi.AsyncSearchStatus_AsyncSearchStatusError
	default:
		return seqproxyapi.AsyncSearchStatus_AsyncSearchStatusInProgress
	}
}

func newProxyStartAsyncSearchRequest(req *seqapi.StartAsyncSearchRequest) *seqproxyapi.StartAsyncSearchRequest {
	return &seqproxyapi.StartAsyncSearchRequest{
		Retention: req.Retention,
		Query: &seqproxyapi.SearchQuery{
			Query: req.Query,
			From:  req.From,
			To:    req.To,
		},
		Aggs:     newProxyAggQuerySlice(req.Aggs),
		Hist:     newProxyHistQuery(req.Hist),
		WithDocs: req.WithDocs,
		Size:     int64(req.Size),
	}
}

type proxyStartAsyncSearchResp seqproxyapi.StartAsyncSearchResponse

func (p *proxyStartAsyncSearchResp) toProto() *seqapi.StartAsyncSearchResponse {
	return &seqapi.StartAsyncSearchResponse{
		SearchId: p.SearchId,
	}
}

func newProxyFetchAsyncSearchResultRequest(
	req *seqapi.FetchAsyncSearchResultRequest,
) *seqproxyapi.FetchAsyncSearchResultRequest {
	return &seqproxyapi.FetchAsyncSearchResultRequest{
		SearchId: req.SearchId,
		Size:     req.Limit,
		Offset:   req.Offset,
		Order:    seqproxyapi.Order(req.Order),
	}
}

type proxyFetchAsyncSearchResultResp seqproxyapi.FetchAsyncSearchResultResponse

func (p *proxyFetchAsyncSearchResultResp) toProto() (*seqapi.FetchAsyncSearchResultResponse, error) {
	resp, err := (*proxySearchResp)(p.Response).toProto()
	if err != nil {
		return nil, err
	}

	return &seqapi.FetchAsyncSearchResultResponse{
		Status:     asyncSearchStatusToProto(p.Status),
		Request:    newSeqapiStartAsyncSearchRequest(p.Request),
		Response:   resp,
		StartedAt:  p.StartedAt,
		ExpiresAt:  p.ExpiresAt,
		CanceledAt: p.CanceledAt,
		Progress:   p.Progress,
		DiskUsage:  p.DiskUsage,
	}, nil
}

func newProxyGetAsyncSearchesListRequest(
	req *seqapi.GetAsyncSearchesListRequest,
	ids []string,
) *seqproxyapi.GetAsyncSearchesListRequest {
	var status *seqproxyapi.AsyncSearchStatus
	if req.Status != nil {
		s := asyncSearchStatusFromProto(*req.Status)
		status = &s
	}

	return &seqproxyapi.GetAsyncSearchesListRequest{
		Status: status,
		Size:   req.Limit,
		Offset: req.Offset,
		Ids:    ids,
	}
}

type proxyGetAsyncSearchesListResp seqproxyapi.GetAsyncSearchesListResponse

func (p *proxyGetAsyncSearchesListResp) toProto() *seqapi.GetAsyncSearchesListResponse {
	searches := make([]*seqapi.GetAsyncSearchesListResponse_ListItem, 0, len(p.Searches))

	for _, s := range p.Searches {
		searches = append(searches, &seqapi.GetAsyncSearchesListResponse_ListItem{
			SearchId:   s.SearchId,
			Status:     asyncSearchStatusToProto(s.Status),
			Request:    newSeqapiStartAsyncSearchRequest(s.Request),
			StartedAt:  s.StartedAt,
			ExpiresAt:  s.ExpiresAt,
			CanceledAt: s.CanceledAt,
			Progress:   s.Progress,
			DiskUsage:  s.DiskUsage,
			Error:      s.Error,
		})
	}

	return &seqapi.GetAsyncSearchesListResponse{
		Searches: searches,
	}
}

func newSeqapiStartAsyncSearchRequest(r *seqproxyapi.StartAsyncSearchRequest) *seqapi.StartAsyncSearchRequest {
	var hist *seqapi.StartAsyncSearchRequest_HistQuery
	if r.Hist != nil {
		hist = &seqapi.StartAsyncSearchRequest_HistQuery{
			Interval: r.Hist.Interval,
		}
	}

	return &seqapi.StartAsyncSearchRequest{
		Retention: r.Retention,
		Query:     r.Query.Query,
		From:      r.Query.From,
		To:        r.Query.To,
		Aggs:      newSeqapiAggQuerySlice(r.Aggs),
		Hist:      hist,
		WithDocs:  r.WithDocs,
		Size:      int32(r.Size),
	}
}

func newSeqapiAggQuerySlice(aggs []*seqproxyapi.AggQuery) []*seqapi.AggregationQuery {
	result := make([]*seqapi.AggregationQuery, 0, len(aggs))

	for _, agg := range aggs {
		result = append(result, &seqapi.AggregationQuery{
			Field:     agg.Field,
			GroupBy:   agg.GroupBy,
			Func:      seqapi.AggFunc(agg.Func),
			Quantiles: agg.Quantiles,
			Interval:  agg.Interval,
		})
	}

	return result
}

func newProxyCancelAsyncSearchRequest(req *seqapi.CancelAsyncSearchRequest) *seqproxyapi.CancelAsyncSearchRequest {
	return &seqproxyapi.CancelAsyncSearchRequest{
		SearchId: req.SearchId,
	}
}

func newProxyDeleteAsyncSearchRequest(req *seqapi.DeleteAsyncSearchRequest) *seqproxyapi.DeleteAsyncSearchRequest {
	return &seqproxyapi.DeleteAsyncSearchRequest{
		SearchId: req.SearchId,
	}
}

type proxyStatusResp seqproxyapi.StatusResponse

func (p *proxyStatusResp) toProto() *seqapi.StatusResponse {
	return &seqapi.StatusResponse{
		NumberOfStores:    p.NumberOfStores,
		OldestStorageTime: p.OldestStorageTime,
		Stores:            proxyStoreSlice(p.Stores).toProto(),
	}
}

type proxyStoreSlice []*seqproxyapi.StoreStatus

func (s proxyStoreSlice) toProto() []*seqapi.StoreStatus {
	result := make([]*seqapi.StoreStatus, len(s))
	for i := range s {
		result[i] = (*proxyStore)(s[i]).toProto()
	}

	return result
}

type proxyStore seqproxyapi.StoreStatus

func (s *proxyStore) toProto() *seqapi.StoreStatus {
	return &seqapi.StoreStatus{
		Host:   s.Host,
		Values: (*proxyStoreStatusValues)(s.Values).toProto(),
		Error:  s.Error,
	}
}

type proxyStoreStatusValues seqproxyapi.StoreStatusValues

func (v *proxyStoreStatusValues) toProto() *seqapi.StoreStatusValues {
	if v == nil {
		return nil
	}

	return &seqapi.StoreStatusValues{
		OldestTime: v.OldestTime,
	}
}
