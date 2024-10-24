package services

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"triple-s/internal/models"
)

// WriteObjectInfo writes or updates object metadata to a CSV file and returns an error if something goes wrong.
func WriteObjectInfo(r *http.Request, dirPath, bucketName, objectKey string) error {
	// Get file information
	fileInfo, err := os.Stat(dirPath + bucketName + "/" + objectKey)
	if err != nil {
		return errors.New("cannot read file: " + err.Error())
	}
	contType := r.Header.Get("Content-Type")

	// Initialize object metadata
	localObject := models.Object{
		ObjectKey:        objectKey,
		Size:             fileInfo.Size(),
		ContentType:      contType,
		LastModifiedTime: fileInfo.ModTime().Format(time.RFC3339),
	}

	// Prepare the CSV file path
	csvFilePath := dirPath + bucketName + "/objects.csv"

	// Read existing CSV data
	var updatedRecords [][]string
	encodedName := base64.StdEncoding.EncodeToString([]byte(localObject.ObjectKey))
	encodedModTime := base64.StdEncoding.EncodeToString([]byte(localObject.LastModifiedTime))

	// Open the CSV file for reading
	objectInfo, err := os.Open(csvFilePath)
	if err != nil {
		return errors.New("error opening CSV file: " + err.Error())
	}
	defer objectInfo.Close()

	// Create a CSV reader
	reader := csv.NewReader(objectInfo)
	records, err := reader.ReadAll()
	if err != nil {
		return errors.New("error reading CSV file: " + err.Error())
	}

	// Update existing records or append new record
	found := false
	for _, record := range records {
		if record[0] == encodedName {
			// Update the existing record with new values
			updatedRecords = append(updatedRecords, []string{
				encodedName,
				strconv.Itoa(int(localObject.Size)),
				localObject.ContentType,
				encodedModTime,
			})
			found = true
		} else {
			// Keep the existing record
			updatedRecords = append(updatedRecords, record)
		}
	}

	if !found {
		// Append new metadata if the object is not found
		updatedRecords = append(updatedRecords, []string{
			encodedName,
			strconv.Itoa(int(localObject.Size)),
			localObject.ContentType,
			encodedModTime,
		})
	}

	// Open the CSV file for writing (this will truncate the file)
	objectInfo, err = os.OpenFile(csvFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return errors.New("error opening CSV file for writing: " + err.Error())
	}
	defer objectInfo.Close()

	// Write updated records to the CSV file
	writer := csv.NewWriter(objectInfo)
	if err := writer.WriteAll(updatedRecords); err != nil {
		return errors.New("error writing updated CSV data: " + err.Error())
	}
	writer.Flush()

	// Check if there were any errors during the flush process
	if err := writer.Error(); err != nil {
		return errors.New("error flushing CSV data: " + err.Error())
	}

	// Return nil if everything succeeded
	return nil
}
