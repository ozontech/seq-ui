package seqdb

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"go.uber.org/zap"
)

const (
	exportJSONLFormat  = `{"id":%q,"data":%s,"time":%s}`
	exportCSVSeparator = ','

	batchSize = 50
)

func (c *GRPCClient) Export(ctx context.Context, req *seqapi.ExportRequest, cw *httputil.ChunkedWriter) error {
	proxyReq := newProxyExportReq(req)
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.Export(ctx, proxyReq)
		},
	)
	if err != nil {
		return err
	}

	csvWriter := newCsvWriter(cw)

	if req.Format == seqapi.ExportFormat_EXPORT_FORMAT_CSV {
		err = csvWriter.Write(req.Fields)
		if err != nil {
			return err
		}
	}

	proxyStream := proxyResp.(seqproxyapi.SeqProxyApi_ExportClient)
	i := 0
	for {
		eResp, err := proxyStream.Recv()
		i++
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			logger.Error("grpc client stream recv failed",
				zap.String("method", "Export"),
				zap.Error(err),
			)
			metric.SeqDBClientStreamError.WithLabelValues("Export", "recv").Inc()
			return err
		}

		if eResp == nil || eResp.Doc == nil {
			continue
		}

		if req.Format == seqapi.ExportFormat_EXPORT_FORMAT_CSV {
			m, err := newMapStringString(eResp.Doc.Data)
			if err != nil {
				continue
			}
			if c.masker != nil {
				c.masker.Mask(m)
			}
			if err := csvWriter.Write(m.getValues(req.Fields, false)); err != nil {
				continue
			}
		} else {
			timeJson, _ := json.Marshal(eResp.Doc.Time.AsTime())

			data := eResp.Doc.Data
			if c.masker != nil {
				if m, err := newMapStringString(data); err == nil {
					c.masker.Mask(m)
					if d, err := json.Marshal(m); err == nil {
						data = d
					}
				}
			}

			cw.WriteString(fmt.Sprintf(exportJSONLFormat, eResp.Doc.Id, data, timeJson))
		}

		if i%batchSize == 0 {
			csvWriter.Flush()
			cw.Flush()
		}
	}

	csvWriter.Flush()
	cw.Flush()

	return nil
}

func newCsvWriter(w io.Writer) *csv.Writer {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = exportCSVSeparator
	csvWriter.UseCRLF = true
	return csvWriter
}
