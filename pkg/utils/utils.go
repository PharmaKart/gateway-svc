package utils

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetIntQueryParam(c *gin.Context, key string, defaultValue int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return defaultValue
	}
	return value
}

func UploadImageToS3(c *gin.Context, bucket string, file *multipart.FileHeader) (string, error) {
	// Open the uploaded file
	// src, err := file.Open()
	// if err != nil {
	// 	return "", err
	// }
	// defer src.Close()

	_ = c // Remove this line

	// Generate a unique file name
	fileName := fmt.Sprintf("%s/%d%s", bucket, time.Now().UnixNano(), filepath.Ext(file.Filename))

	Info("Uploading file to S3", map[string]interface{}{
		"file_name": fileName,
	})

	// Upload file to S3
	// url, err := s3Client.UploadFile(c.Request.Context(), src, fileName)
	// if err != nil {
	// 	return "", err
	// }

	url := "https://picsum.photos/500"

	return url, nil
}
