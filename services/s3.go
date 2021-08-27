package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type PresignedURLOption struct {
	Keys []string
}

func GeneratePresignedURLs(keys []string) ([]string, error) {
	size := len(keys)
	urls := make([]string, size)
	if size < 1 {
		return nil, fmt.Errorf("length of urls should be at least 1")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client, func(options *s3.PresignOptions) {
		options.Expires = time.Minute * 10
	})

	for index, key := range keys {
		presignedRequest, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("AWS_BUCKET")),
			Key:    aws.String(key),
		})
		if err != nil {
			return nil, err
		}

		urls[index] = presignedRequest.URL
	}

	return urls, nil
}
