package utils

import (
	"app/config"
	"app/consts"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
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
