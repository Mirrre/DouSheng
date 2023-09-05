package utils

import (
	"app/config"
	"app/consts"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/minio/minio-go/v7"
	"os"
	"time"
)

func UploadFileToS3(filePath, key string) error {
	fileContent, err := os.Open(filePath)
	if err != nil {
		return err
	}

	_, err = config.S3Client.PutObject(&s3.PutObjectInput{
		ACL:    aws.String("public-read"),
		Body:   fileContent,
		Bucket: aws.String(consts.AwsBucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}

func UploadFileToMinIO(filePath, objectName, bucketName string, expiration time.Duration) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = config.MinioClient.PutObject(
		context.Background(), bucketName, objectName, file, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}

	presignedURL, err := config.MinioClient.PresignedGetObject(
		context.Background(), bucketName, objectName, expiration, nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}
