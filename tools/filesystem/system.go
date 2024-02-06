package filesystem

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/s3blob"
)

type System struct {
	ctx    context.Context
	bucket *blob.Bucket
}

func NewWithS3(bucketName, region, accessKey, secretKey string) (*System, error) {
	ctx := context.Background()

	awsConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(awsConfig)
	bucket, err := s3blob.OpenBucketV2(ctx, s3Client, bucketName, nil)
	if err != nil {
		return nil, err
	}

	return &System{ctx: ctx, bucket: bucket}, nil
}

func NewLocal(dirPath string) (*System, error) {
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, err
	}

	bucket, err := fileblob.OpenBucket(dirPath, nil)
	if err != nil {
		return nil, err
	}

	return &System{bucket: bucket}, nil
}

func (s *System) Upload(content []byte, fileKey string) error {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	opts := &blob.WriterOptions{
		ContentType: mimetype.Detect(content).String(),
	}

	w, err := s.bucket.NewWriter(ctx, fileKey, opts)
	if err != nil {
		return err
	}

	if _, err := w.Write(content); err != nil {
		w.Close()
		return err
	}

	return w.Close()
}

func (s *System) Delete(fileKey string) error {
	return s.bucket.Delete(s.ctx, fileKey)
}
