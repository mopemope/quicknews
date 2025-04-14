package storage

import (
	"context"
	"fmt"
	"io"

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
	if cfg.Cloudflare == nil {
		return nil, errors.New("Cloudflare R2 configuration is missing")
	}
	r2Config := cfg.Cloudflare
	if r2Config.AccessKeyID == "" || r2Config.SecretAccessKey == "" || r2Config.BucketName == "" || r2Config.EndpointURL == "" {
		return nil, errors.New("missing required Cloudflare R2 configuration fields (AccessKeyID, SecretAccessKey, BucketName, EndpointURL)")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2Config.AccessKeyID, r2Config.SecretAccessKey, "")),
		awsconfig.WithRegion("auto"),
		//		awsconfig.WithRequestChecksumCalculation(0),
		// awsconfig.WithResponseChecksumValidation(0),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load AWS config")
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Config.EndpointURL)
	})

	return &R2Storage{
		client:     s3Client,
		bucketName: r2Config.BucketName,
	}, nil
}

// Upload uploads data to the specified key in the R2 bucket.
func (r *R2Storage) Upload(ctx context.Context, key string, reader io.Reader) error {
	fmt.Println("Uploading to R2 bucket:", r.bucketName, "with key:", key)

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String("audio/mpeg"),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to upload object %q to R2 bucket %q", key, r.bucketName)
	}
	return nil
}
