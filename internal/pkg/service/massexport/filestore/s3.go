package filestore

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ozontech/seq-ui/internal/app/config"
)

type s3FileStore struct {
	uploader *s3manager.Uploader

	bucketName string
}

func NewS3(cfg *config.S3) (FileStore, error) {
	disableSSL := !cfg.EnableSSl
	s3Session, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint:         aws.String(cfg.Endpoint),
			S3ForcePathStyle: aws.Bool(true),
			Region:           aws.String("us-east-1"),
			Credentials:      credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
			DisableSSL:       &disableSSL,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 session: %w", err)
	}

	client := s3.New(s3Session, aws.NewConfig())

	uploader := s3manager.NewUploaderWithClient(client)

	return &s3FileStore{
		uploader:   uploader,
		bucketName: cfg.BucketName,
	}, nil
}

func (s *s3FileStore) PutObject(ctx context.Context, objectName string, reader io.Reader) error {
	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Key:    aws.String(objectName),
		Bucket: aws.String(s.bucketName),
		Body:   reader,
	})

	return err
}
