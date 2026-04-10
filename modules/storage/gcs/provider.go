package gcs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/volvlabs/nebularcore/modules/storage/models"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// gcsStorageAPI defines the interface for GCS storage operations
type gcsStorageAPI interface {
	writeObject(name string, data []byte, attrs *storage.ObjectAttrs) error
	readObject(name string) ([]byte, *storage.ObjectAttrs, error)
	deleteObject(name string) error
	listObjects(prefix string) ([]*storage.ObjectAttrs, error)
}

type Config struct {
	Bucket          string `yaml:"bucket" validate:"required"`
	CredentialsFile string `yaml:"credentialsFile"`
	CredentialsJSON string `yaml:"credentialsJSON"`
}

type Provider struct {
	client gcsStorageAPI
	bucket string
}

func New(cfg Config) (*Provider, error) {
	ctx := context.Background()

	// Set up client options
	opts := []option.ClientOption{}

	// Add credentials if provided
	if cfg.CredentialsJSON != "" {
		creds, err := google.CredentialsFromJSON(ctx, []byte(cfg.CredentialsJSON), storage.ScopeReadWrite)
		if err != nil {
			return nil, fmt.Errorf("failed to parse credentials JSON: %w", err)
		}
		opts = append(opts, option.WithCredentials(creds))
	} else if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
	}

	// Create client
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Wrap the real client in an adapter that implements gcsStorageAPI
	adapter := &gcsClientAdapter{client: client, bucket: cfg.Bucket}

	return &Provider{
		client: adapter,
		bucket: cfg.Bucket,
	}, nil
}

// gcsClientAdapter adapts the real GCS client to our gcsStorageAPI interface
type gcsClientAdapter struct {
	client *storage.Client
	bucket string
}

func (a *gcsClientAdapter) writeObject(name string, data []byte, attrs *storage.ObjectAttrs) error {
	ctx := context.Background()
	obj := a.client.Bucket(a.bucket).Object(name)
	writer := obj.NewWriter(ctx)

	// Set attributes if provided
	if attrs != nil {
		writer.ContentType = attrs.ContentType
		writer.Metadata = attrs.Metadata
	}

	// Write data
	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return err
	}

	return writer.Close()
}

func (a *gcsClientAdapter) readObject(name string) ([]byte, *storage.ObjectAttrs, error) {
	ctx := context.Background()
	obj := a.client.Bucket(a.bucket).Object(name)

	// Get attributes
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Read data
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	return data, attrs, nil
}

func (a *gcsClientAdapter) deleteObject(name string) error {
	ctx := context.Background()
	return a.client.Bucket(a.bucket).Object(name).Delete(ctx)
}

func (a *gcsClientAdapter) listObjects(prefix string) ([]*storage.ObjectAttrs, error) {
	ctx := context.Background()
	var result []*storage.ObjectAttrs

	it := a.client.Bucket(a.bucket).Objects(ctx, &storage.Query{Prefix: prefix})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, attrs)
	}

	return result, nil
}

func (p *Provider) Upload(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error) {
	if input.Key == "" {
		return nil, fmt.Errorf("key is required")
	}

	// Read all input data
	data, err := io.ReadAll(input.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read input data: %w", err)
	}

	// Prepare metadata
	metadata := make(map[string]string)
	metadata["original-name"] = input.FileName
	for k, v := range input.Metadata {
		metadata[k] = v
	}

	// Create object attributes
	attrs := &storage.ObjectAttrs{
		Name:        input.Key,
		ContentType: input.ContentType,
		Metadata:    metadata,
	}

	// Write object
	if err := p.client.writeObject(input.Key, data, attrs); err != nil {
		return nil, fmt.Errorf("failed to write object: %w", err)
	}

	// Generate URL
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", p.bucket, input.Key)

	return &models.UploadOutput{
		Path:        input.Key,
		URL:         url,
		ContentType: input.ContentType,
		Size:        int64(len(data)),
		Metadata:    metadata,
	}, nil
}

func (p *Provider) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	data, _, err := p.client.readObject(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (p *Provider) Delete(ctx context.Context, path string) error {
	if err := p.client.deleteObject(path); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (p *Provider) List(ctx context.Context, prefix string) ([]models.FileInfo, error) {
	attrs, err := p.client.listObjects(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var files []models.FileInfo
	for _, attr := range attrs {
		contentType := attr.ContentType
		if contentType == "" {
			contentType = getContentType(attr.Name)
		}

		files = append(files, models.FileInfo{
			Path:        attr.Name,
			Size:        attr.Size,
			ContentType: contentType,
			ModTime:     attr.Updated,
			IsDir:       strings.HasSuffix(attr.Name, "/"),
		})
	}

	return files, nil
}

func (p *Provider) Close() error {
	// No-op for mock client
	return nil
}

func getContentType(name string) string {
	ext := path.Ext(name)
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
