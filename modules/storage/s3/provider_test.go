package s3

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"gitlab.com/jideobs/nebularcore/modules/storage/models"
)

// mockS3Client implements the s3ClientAPI interface
type mockS3Client struct {
	putObjectOutput           *s3.PutObjectOutput
	putObjectErr              error
	getObjectOutput           *s3.GetObjectOutput
	getObjectErr              error
	deleteObjectOutput        *s3.DeleteObjectOutput
	deleteObjectErr           error
	listObjectsOutput         *s3.ListObjectsV2Output
	listObjectsErr            error
	headObjectOutput          *s3.HeadObjectOutput
	headObjectErr             error
	createMultipartOutput     *s3.CreateMultipartUploadOutput
	createMultipartErr        error
	uploadPartOutput          *s3.UploadPartOutput
	uploadPartErr             error
	completeMultipartOutput   *s3.CompleteMultipartUploadOutput
	completeMultipartErr      error
	abortMultipartOutput      *s3.AbortMultipartUploadOutput
	abortMultipartErr         error
	listMultipartPartsOutput  *s3.ListPartsOutput
	listMultipartPartsErr     error
	uploadPartCopyOutput      *s3.UploadPartCopyOutput
	uploadPartCopyErr         error
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectOutput, m.putObjectErr
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.getObjectOutput, m.getObjectErr
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return m.deleteObjectOutput, m.deleteObjectErr
}

func (m *mockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return m.listObjectsOutput, m.listObjectsErr
}

func (m *mockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return m.headObjectOutput, m.headObjectErr
}

func (m *mockS3Client) CreateMultipartUpload(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
	return m.createMultipartOutput, m.createMultipartErr
}

func (m *mockS3Client) UploadPart(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
	return m.uploadPartOutput, m.uploadPartErr
}

func (m *mockS3Client) CompleteMultipartUpload(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
	return m.completeMultipartOutput, m.completeMultipartErr
}

func (m *mockS3Client) AbortMultipartUpload(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
	return m.abortMultipartOutput, m.abortMultipartErr
}

func (m *mockS3Client) ListParts(ctx context.Context, params *s3.ListPartsInput, optFns ...func(*s3.Options)) (*s3.ListPartsOutput, error) {
	return m.listMultipartPartsOutput, m.listMultipartPartsErr
}

func (m *mockS3Client) UploadPartCopy(ctx context.Context, params *s3.UploadPartCopyInput, optFns ...func(*s3.Options)) (*s3.UploadPartCopyOutput, error) {
	return m.uploadPartCopyOutput, m.uploadPartCopyErr
}

func TestProvider_Upload(t *testing.T) {
	mockClient := &mockS3Client{
		putObjectOutput: &s3.PutObjectOutput{},
		headObjectOutput: &s3.HeadObjectOutput{
			ContentLength: aws.Int64(100),
		},
	}

	provider := &Provider{
		client:     mockClient,
		uploader:   manager.NewUploader(mockClient),
		downloader: manager.NewDownloader(mockClient),
		bucket:     "test-bucket",
	}

	ctx := context.Background()
	input := &models.UploadInput{
		File:        bytes.NewReader([]byte("test content")),
		FileName:    "test.txt",
		ContentType: "text/plain",
		Key:        "test/file.txt",
		Metadata:    map[string]string{"test": "value"},
	}

	output, err := provider.Upload(ctx, input)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	if output.Path != input.Key {
		t.Errorf("Upload() Path = %v, want %v", output.Path, input.Key)
	}

	if output.Size != 100 {
		t.Errorf("Upload() Size = %v, want %v", output.Size, 100)
	}
}

func TestProvider_Download(t *testing.T) {
	mockBody := io.NopCloser(bytes.NewReader([]byte("test content")))
	mockClient := &mockS3Client{
		getObjectOutput: &s3.GetObjectOutput{
			Body: mockBody,
		},
	}

	provider := &Provider{
		client:     mockClient,
		uploader:   manager.NewUploader(mockClient),
		downloader: manager.NewDownloader(mockClient),
		bucket:     "test-bucket",
	}

	ctx := context.Background()
	reader, err := provider.Download(ctx, "test/file.txt")
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read downloaded content: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Download() content = %v, want %v", string(content), "test content")
	}
}

func TestProvider_Delete(t *testing.T) {
	mockClient := &mockS3Client{
		deleteObjectOutput: &s3.DeleteObjectOutput{},
	}

	provider := &Provider{
		client:     mockClient,
		uploader:   manager.NewUploader(mockClient),
		downloader: manager.NewDownloader(mockClient),
		bucket:     "test-bucket",
	}

	ctx := context.Background()
	err := provider.Delete(ctx, "test/file.txt")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestProvider_List(t *testing.T) {
	now := time.Now()
	mockClient := &mockS3Client{
		listObjectsOutput: &s3.ListObjectsV2Output{
			Contents: []types.Object{
				{
					Key:          aws.String("test/file1.txt"),
					Size:         aws.Int64(100),
					LastModified: aws.Time(now),
				},
				{
					Key:          aws.String("test/folder/"),
					Size:         aws.Int64(0),
					LastModified: aws.Time(now),
				},
			},
		},
	}

	provider := &Provider{
		client:     mockClient,
		uploader:   manager.NewUploader(mockClient),
		downloader: manager.NewDownloader(mockClient),
		bucket:     "test-bucket",
	}

	ctx := context.Background()
	files, err := provider.List(ctx, "test/")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("List() got %v files, want %v", len(files), 2)
	}

	// Check file properties
	if files[0].Path != "test/file1.txt" {
		t.Errorf("List() file path = %v, want %v", files[0].Path, "test/file1.txt")
	}
	if files[0].Size != 100 {
		t.Errorf("List() file size = %v, want %v", files[0].Size, 100)
	}
	if files[0].IsDir {
		t.Error("List() file IsDir = true, want false")
	}

	// Check folder properties
	if !files[1].IsDir {
		t.Error("List() folder IsDir = false, want true")
	}
	if files[1].Size != 0 {
		t.Errorf("List() folder size = %v, want %v", files[1].Size, 0)
	}
}
