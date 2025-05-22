package s3uploader

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config 存储 S3 客户端的配置信息
type S3Config struct {
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	BucketName      string `yaml:"bucket_name"`
	UsePathStyle    bool   `yaml:"use_path_style"` // For MinIO or other S3 compatible services
	FileName        string `yaml:"file_name"`
}

// Uploader 定义了上传接口
type Uploader interface {
	UploadFile(filePath string, objectKey string) error
}

type s3Uploader struct {
	client *s3.Client
	config S3Config
}

// NewS3Uploader 创建一个新的 S3 uploader 实例
func NewS3Uploader(cfg S3Config) (Uploader, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithBaseEndpoint(cfg.Endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.UsePathStyle {
			o.UsePathStyle = true
		}
	})

	return &s3Uploader{
		client: client,
		config: cfg,
	}, nil
}

// UploadFile 将文件上传到 S3 bucket
func (u *s3Uploader) UploadFile(filePath string, objectKey string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = u.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(u.config.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	return err
}
