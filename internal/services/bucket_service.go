package services

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"os"
	"time"

	"triple-s/internal/models"
)

// BucketAndFileCreation creates a bucket and objects.csv file, returns an error if fails
func BucketAndFileCreation(dirPath string) error {
	err := os.Mkdir(dirPath, os.ModePerm)
	if err != nil {
		return errors.New("bucket already exists")
	}

	file, err := os.Create(dirPath + "/objects.csv")
	if err != nil {
		return errors.New("error creating objects.csv")
	}
	defer file.Close()
	return nil
}

// WriteBucketInfo writes bucket metadata to a CSV and returns an error if it fails
func WriteBucketInfo(bucketName string, directoryPath string) (string, error) {
	fileInfo, err := os.Stat(directoryPath + bucketName)
	if err != nil {
		return "", errors.New("error getting file info")
	}

	localBucket := &models.Bucket{
		Name:             bucketName,
		CreationTime:     fileInfo.ModTime().Format(time.RFC3339),
		LastModifiedTime: fileInfo.ModTime().Format(time.RFC3339),
		Status:           "true",
	}

	bucketInfo, err := os.OpenFile(directoryPath+"buckets.csv", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o644)
	if err != nil {
		return "", errors.New("error creating file")
	}
	defer bucketInfo.Close()

	encodedName := base64.StdEncoding.EncodeToString([]byte(fileInfo.Name()))
	encodedModTime := base64.StdEncoding.EncodeToString([]byte(fileInfo.ModTime().Format(time.RFC3339)))
	writer := csv.NewWriter(bucketInfo)

	info := []string{
		encodedName,
		encodedModTime,
		encodedModTime,
		localBucket.Status,
	}

	if err := writer.Write(info); err != nil {
		return "", errors.New("error writing to CSV")
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return "", errors.New("error flushing CSV data")
	}

	xmlData, err := xml.MarshalIndent(localBucket, "", "   ")
	if err != nil {
		return "", errors.New("error generating XML")
	}

	return string(xmlData), nil
}
