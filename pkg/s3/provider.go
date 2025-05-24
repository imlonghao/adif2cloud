package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config 定义了 S3 配置
type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	BucketName      string `mapstructure:"bucket_name"`
	UsePathStyle    bool   `mapstructure:"use_path_style"`
	FileName        string `mapstructure:"file_name"`
}

// S3Provider 实现了 Provider 接口，用于 S3 服务
type S3Provider struct {
	client     *s3.Client
	bucketName string
	fileName   string
}

// NewS3Provider 创建一个新的 S3Provider 实例
func NewS3Provider(cfg S3Config) (*S3Provider, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithBaseEndpoint(cfg.Endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return &S3Provider{
		client:     client,
		bucketName: cfg.BucketName,
		fileName:   cfg.FileName,
	}, nil
}

// GetSize 获取 S3 上 ADIF 文件的大小
func (p *S3Provider) GetSize() (int64, error) {
	head, err := p.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(p.fileName),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get object size: %w", err)
	}
	return *head.ContentLength, nil
}

// Download 从 S3 下载 ADIF 文件
func (p *S3Provider) Download(w io.Writer) error {
	result, err := p.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(p.fileName),
	})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	_, err = io.Copy(w, result.Body)
	return err
}

// Upload 上传文件到 S3
func (p *S3Provider) Upload(filename string, _ string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	_, err = p.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(p.fileName),
		Body:   bytes.NewReader(content),
	})
	return err
}

// GetName 获取提供商的名称
func (p *S3Provider) GetName() string {
	return fmt.Sprintf("S3-%s-%s", *p.client.Options().BaseEndpoint, p.bucketName)
}
