package filestore

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ozontech/seq-ui/internal/app/config"
)

type s3FileStore struct {
	tm *transfermanager.Client

	bucketName string
}

func NewS3(ctx context.Context, cfg *config.S3) (FileStore, error) {
	scheme := "http"
	if cfg.EnableSSl {
		scheme = "https"
	}

	awsCfg, err := aws_cfg.LoadDefaultConfig(
		ctx,
		aws_cfg.WithBaseEndpoint(scheme+"://"+cfg.Endpoint),
		aws_cfg.WithRegion("us-east-1"),
		aws_cfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &s3FileStore{
		tm:         transfermanager.New(client),
		bucketName: cfg.BucketName,
	}, nil
}

func (s *s3FileStore) PutObject(ctx context.Context, objectName string, reader io.Reader) error {
	_, err := s.tm.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Key:    aws.String(objectName),
		Bucket: aws.String(s.bucketName),
		Body:   reader,
	})

	return err
}
