package utils

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"time"

	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

func GetIntQueryParam(c *gin.Context, key string, defaultValue int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return defaultValue
	}
	return value
}

func UploadImageToS3(c *gin.Context, cfg *config.Config, bucketFolder string, file *multipart.FileHeader) (string, error) {
	// TODO: Add this to env variables
	bucket := cfg.S3Bucket
	region := cfg.AwsRegion

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate a unique file name
	fileName := fmt.Sprintf("%s/%d%s", bucketFolder, time.Now().UnixNano(), filepath.Ext(file.Filename))

	Info("Uploading file to S3", map[string]interface{}{
		"file_name": fileName,
	})

	// Start a new S3 session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return "", err
	}

	// Upload the file to S3
	s3Service := s3.New(sess)

	_, err = s3Service.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileName),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, fileName)

	return url, nil
}
