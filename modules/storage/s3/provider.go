package s3

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/volvlabs/nebularcore/modules/storage/models"
)

// s3ClientAPI defines the interface for S3 client operations
type s3ClientAPI interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

type Provider struct {
	client     s3ClientAPI
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucket     string
}

type Config struct {
	Bucket          string `yaml:"bucket" validate:"required"`
	Region          string `yaml:"region" validate:"required"`
	AccessKeyID     string `yaml:"accessKeyId" validate:"required_with=SecretAccessKey"`
	SecretAccessKey string `yaml:"secretAccessKey" validate:"required_with=AccessKeyID"`
	SessionToken    string `yaml:"sessionToken"`
	Endpoint        string `yaml:"endpoint"`
	ForcePathStyle  bool   `yaml:"forcePathStyle"`
}

func New(cfg Config) (*Provider, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	// Add credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		creds := credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
				SessionToken:    cfg.SessionToken,
			},
		}
		opts = append(opts, config.WithCredentialsProvider(creds))
	}

	// Add custom endpoint if provided
	if cfg.Endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: cfg.Endpoint, SigningRegion: cfg.Region}, nil
		})
		opts = append(opts, config.WithEndpointResolverWithOptions(customResolver))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if cfg.ForcePathStyle {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)

	return &Provider{
		client:     client,
		uploader:   manager.NewUploader(client),
		downloader: manager.NewDownloader(client),
		bucket:     cfg.Bucket,
	}, nil
}

func (p *Provider) Upload(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error) {
	if input.Key == "" {
		return nil, fmt.Errorf("key is required")
	}

	// Prepare metadata
	metadata := make(map[string]string)
	metadata["original-name"] = input.FileName
	for k, v := range input.Metadata {
		metadata[k] = v
	}

	// Upload the file
	result, err := p.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(input.Key),
		Body:        input.File,
		ContentType: aws.String(input.ContentType),
		Metadata:    metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Get object info to return size
	head, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(input.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return &models.UploadOutput{
		Path:        input.Key,
		URL:         result.Location,
		ContentType: input.ContentType,
		Size:        *head.ContentLength,
		Metadata:    metadata,
	}, nil
}

func (p *Provider) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return result.Body, nil
}

func (p *Provider) Delete(ctx context.Context, path string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

func (p *Provider) List(ctx context.Context, prefix string) ([]models.FileInfo, error) {
	var files []models.FileInfo
	var continuationToken *string

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(p.bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		}

		result, err := p.client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range result.Contents {
			contentType := ""
			if obj.Key != nil {
				contentType = getContentType(*obj.Key)
			}

			files = append(files, models.FileInfo{
				Path:        aws.ToString(obj.Key),
				Size:        *obj.Size,
				ContentType: contentType,
				ModTime:     aws.ToTime(obj.LastModified),
				IsDir:       strings.HasSuffix(aws.ToString(obj.Key), "/"),
			})
		}

		if result.IsTruncated == nil || !*result.IsTruncated {
			break
		}
		continuationToken = result.NextContinuationToken
	}

	return files, nil
}

func getContentType(key string) string {
	ext := path.Ext(key)
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
