package massexport

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type loadTask struct {
	fileStorePath string
	query         string

	partID int

	from   time.Time
	to     time.Time
	window time.Duration
}

func (s *exportService) exportAllParts(
	ctx context.Context,
	sessionID string,
) {
	ctx, span := tracing.StartSpan(ctx, "export_all_parts")
	defer span.End()

	logger.Info("export started", zap.String("session_id", sessionID))

	start := time.Now()

	info, err := s.sessionStore.CheckExport(ctx, sessionID)
	if err != nil {
		logger.Error("can't get export info to start/continue export", zap.Error(err), zap.String("session_id", sessionID))
		return
	}

	var (
		from = info.From
		to   = info.To

		fileStorePathPrefix = info.FileStorePathPrefix
		linksPath           = fmt.Sprintf("%s/links", info.FileStorePathPrefix)

		query  = info.Query
		window = info.Window
	)

	// buffered channel is required to prevent producer blocking
	// (see exportOnePartWrk func: consumer stops reading from channel after first error)
	ch := make(chan *loadTask, s.tasksChannelSize)

	wg := &sync.WaitGroup{}
	for i := 0; i < s.workersCount; i++ {
		wg.Add(1)
		go s.exportOnePartWrk(ctx, sessionID, ch, wg)
	}

	// links to file with logs
	links := make([]string, 0, to.Sub(from)/s.partLength+1)
	partID := 0

	const timeFormat = "2006-01-02T15-04"
	for subTo := to; subTo.After(from); subTo = subTo.Add(-s.partLength) {
		subFrom := subTo.Add(-s.partLength)
		fileStorePath := fmt.Sprintf(
			"%s/%s_%s.json.gz",
			fileStorePathPrefix,
			subFrom.Format(timeFormat),
			subTo.Format(timeFormat)[len(timeFormat)-len("10-00"):],
		)
		links = append(links, s.getFileLink(fileStorePath))

		if !info.PartIsUploaded[partID] {
			ch <- &loadTask{
				fileStorePath: fileStorePath,
				query:         query,
				from:          subFrom,
				to:            subTo,
				window:        window,
				partID:        partID,
			}
		}

		partID++
	}

	close(ch)
	wg.Wait()

	err = s.sessionStore.FinishExport(ctx, sessionID)
	if err != nil {
		logger.Error(
			"export failed",
			zap.String("session_id", sessionID),
			zap.Duration("duration", time.Since(start)),
			zap.Error(fmt.Errorf("finish export: %w", err)),
		)
		return
	}

	// upload file that contains links to files with logs
	links = append(links, "")
	err = s.fileStore.PutObject(
		ctx,
		linksPath,
		strings.NewReader(strings.Join(links, "\n")),
	)

	if err != nil {
		logger.Error("export finished, but can't upload links",
			zap.String("session_id", sessionID),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return
	}

	logger.Info("export successfully finished",
		zap.String("session_id", sessionID),
		zap.Duration("duration", time.Since(start)),
		zap.String("links", s.getFileLink(linksPath)),
	)
}

func (s *exportService) exportOnePartWrk(ctx context.Context, sessionID string, tasks chan *loadTask, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		err := s.exportOnePart(ctx, sessionID, task)
		if err != nil {
			logger.Error("can't export one part", zap.Error(err), zap.String("session_id", sessionID))
			err2 := s.sessionStore.FailExport(ctx, sessionID, err.Error())
			if err2 != nil {
				logger.Error("can't fail export", zap.Error(err2), zap.String("session_id", sessionID))
			}
			break
		}
	}
}

func (s *exportService) exportOnePart(ctx context.Context, sessionID string, task *loadTask) error {
	ctx, span := tracing.StartSpan(ctx, "export_one_part")
	defer span.End()

	defer func(start time.Time) {
		duration := time.Since(start)
		logger.Info(
			"one part exported",
			zap.Duration("duration", duration),
			zap.String("session_id", sessionID),
		)
		metric.MassExportOnePartExportDuration.WithLabelValues(sessionID).Observe(duration.Seconds())
	}(time.Now())

	reader, writer := io.Pipe()

	packedWriter := NewSizeWriter(writer)
	gzWriter, _ := gzip.NewWriterLevel(packedWriter, gzip.BestCompression)
	unpackedWriter := NewSizeWriter(gzWriter)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		defer func() {
			err := writer.Close()
			if err != nil {
				logger.Error("can't close pipe writer", zap.Error(err), zap.String("session_id", sessionID))
			}
		}()

		defer func() {
			err := gzWriter.Close()
			if err != nil {
				logger.Error("can't close gzip writer", zap.Error(err), zap.String("session_id", sessionID))
			}
		}()

		err := s.downloadOnePart(groupCtx, unpackedWriter, task, sessionID)
		return err
	})
	group.Go(func() error {
		defer func() {
			err := reader.Close()
			if err != nil {
				logger.Error("can't close pipe reader", zap.Error(err), zap.String("session_id", sessionID))
			}
		}()

		err := s.uploadOnePart(groupCtx, reader, task)
		return err
	})

	err := group.Wait()
	if err != nil {
		return fmt.Errorf("download/upload one part: %w", err)
	}

	err = s.sessionStore.ConfirmPart(ctx, sessionID, task.partID, types.Size{
		Unpacked: unpackedWriter.Size(),
		Packed:   packedWriter.Size(),
	})
	if err != nil {
		return fmt.Errorf("confirm part: %w", err)
	}

	return nil
}

func (s *exportService) downloadOnePart(ctx context.Context, writer io.Writer, task *loadTask, sessionID string) error {
	for subTo := task.to; subTo.After(task.from); subTo = subTo.Add(-task.window) {
		info, err := s.sessionStore.CheckExport(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("check export status: %w", err)
		}

		switch info.Status {
		case types.ExportStatusCancel:
			return errors.New("worker stops: export canceled")
		case types.ExportStatusFail:
			return errors.New("worker stops: export failed")
		case types.ExportStatusFinish:
			return errors.New("impossible: export finished before worker stopped")
		case types.ExportStatusStart:
			break
		default:
			return fmt.Errorf("unknown export status: %s", info.Status.String())
		}

		subFrom := subTo.Add(-task.window)
		if subFrom.Before(task.from) {
			subFrom = task.from
		}

		err = s.fetchBatched(
			ctx,
			sessionID,
			task.query,
			subFrom,
			subTo.Add(-time.Millisecond),
			writer,
		)
		if err != nil {
			return fmt.Errorf("batch fetch: %w", err)
		}
	}

	return nil
}

func (s *exportService) fetchBatched(ctx context.Context, sessionID string, query string, from, to time.Time, writer io.Writer) error {
	offset := 0

	for {
		resp, err := s.downloader.Search(ctx,
			sessionID,
			&seqapi.SearchRequest{
				Query:     query,
				From:      timestamppb.New(from),
				To:        timestamppb.New(to),
				Limit:     int32(s.batchSize),
				Offset:    int32(offset),
				WithTotal: false,
			})
		if err != nil {
			return err
		}

		for _, event := range resp.Events {
			data, err := json.Marshal(event)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(writer, string(data))
			if err != nil {
				return err
			}
		}

		if len(resp.Events) < int(s.batchSize) {
			break
		}

		offset += len(resp.Events)
	}

	return nil
}

func (s *exportService) uploadOnePart(
	ctx context.Context,
	reader io.Reader,
	task *loadTask,
) error {
	err := s.fileStore.PutObject(ctx, task.fileStorePath, reader)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}
