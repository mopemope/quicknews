package storage

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/config"
)

// R2Storage provides methods for interacting with Cloudflare R2 storage.
type R2Storage struct {
	client     *s3.Client
	bucketName string
}

// NewR2Storage creates a new R2Storage client.
func NewR2Storage(ctx context.Context, cfg *config.Config) (*R2Storage, error) {
	if cfg.CloudflareR2 == nil {
		return nil, errors.New("Cloudflare R2 configuration is missing")
	}
	r2Config := cfg.CloudflareR2

	if r2Config.AccountID == "" || r2Config.AccessKeyID == "" || r2Config.SecretAccessKey == "" || r2Config.BucketName == "" || r2Config.EndpointURL == "" {
		return nil, errors.New("missing required Cloudflare R2 configuration fields (AccountID, AccessKeyID, SecretAccessKey, BucketName, EndpointURL)")
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: r2Config.EndpointURL,
		}, nil
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithEndpointResolverWithOptions(resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2Config.AccessKeyID, r2Config.SecretAccessKey, "")),
		// Note: R2 doesn't use regions in the same way AWS S3 does, but the SDK might require one.
		// Using a placeholder region like "auto" or a specific one if needed. Check R2 documentation.
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load AWS config for R2")
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &R2Storage{
		client:     s3Client,
		bucketName: r2Config.BucketName,
	}, nil
}

// Upload uploads data to the specified key in the R2 bucket.
func (r *R2Storage) Upload(ctx context.Context, key string, data []byte) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
		// Consider setting ContentType based on the data if known
		// ContentType: aws.String("audio/wav"),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to upload object %q to R2 bucket %q", key, r.bucketName)
	}
	return nil
}
