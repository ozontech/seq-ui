package seqdb

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
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

	stream := proxyResp.(seqproxyapi.SeqProxyApi_ExportClient)
	return c.writeExportStream(stream, req.Format, req.Fields, cw, "Export")
}

// exportStream is the commonRecv part of both Export and ExportAsyncSearch
// gRPC client streams, which both yield *seqproxyapi.ExportResponse messages.
type exportStream interface {
	Recv() (*seqproxyapi.ExportResponse, error)
}

// writeExportStream reads documents from the given gRPC stream and writes them
// to cw in the requested format (JSONL or CSV), applying masking when enabled.
// method is used for logging and metrics labels.
func (c *GRPCClient) writeExportStream(
	stream exportStream,
	format seqapi.ExportFormat,
	fields []string,
	cw *httputil.ChunkedWriter,
	method string,
) error {
	csvWriter := newCsvWriter(cw)

	if format == seqapi.ExportFormat_EXPORT_FORMAT_CSV {
		if err := csvWriter.Write(fields); err != nil {
			return err
		}
	}

	i := 0
	for {
		eResp, err := stream.Recv()
		i++
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			logger.Error("grpc client stream recv failed",
				zap.String("method", method),
				zap.Error(err),
			)
			metric.SeqDBClientStreamError.WithLabelValues(method, "recv").Inc()
			return err
		}

		if eResp == nil || eResp.Doc == nil {
			continue
		}

		if format == seqapi.ExportFormat_EXPORT_FORMAT_CSV {
			m, err := newMapStringString(eResp.Doc.Data)
			if err != nil {
				continue
			}
			if c.masker != nil {
				c.masker.Mask(m)
			}
			if err := csvWriter.Write(m.getValues(fields, false)); err != nil {
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
