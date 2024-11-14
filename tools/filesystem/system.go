package filesystem

import (
	"context"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/memblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

type System struct {
	ctx    context.Context
	bucket *blob.Bucket

	IsBucketClosed bool
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

func NewWithGoogleCloudStorage(bucketName, credfileLocation string) (*System, error) {
	ctx := context.Background()

	credContent, err := os.ReadFile(credfileLocation)
	if err != nil {
		return nil, err
	}
	creds, err := google.CredentialsFromJSON(ctx, credContent, storage.ScopeReadWrite)
	if err != nil {
		return nil, err
	}

	client, err := gcp.NewHTTPClient(
		gcp.DefaultTransport(),
		gcp.CredentialsTokenSource(creds))
	if err != nil {
		return nil, err
	}

	bucket, err := gcsblob.OpenBucket(ctx, client, bucketName, nil)
	if err != nil {
		return nil, err
	}

	return &System{ctx: ctx, bucket: bucket}, nil
}

func NewLocal(dirPath string) (*System, error) {
	ctx := context.Background()
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, err
	}

	bucket, err := fileblob.OpenBucket(dirPath, nil)
	if err != nil {
		return nil, err
	}

	return &System{ctx: ctx, bucket: bucket}, nil
}

func NewMemory() (*System, error) {
	ctx := context.Background()
	bucket := memblob.OpenBucket(nil)

	return &System{ctx: ctx, bucket: bucket}, nil
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

// This is only meant for development purposes only as in production files
// would be served using cloudfront.
func (s *System) Download(fileKey string) ([]byte, string, error) {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	r, err := s.bucket.NewReader(ctx, fileKey, nil)
	if err != nil {
		return nil, "", err
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	return content, r.ContentType(), nil
}

func (s *System) Close() error {
	s.IsBucketClosed = true
	return s.bucket.Close()
}
