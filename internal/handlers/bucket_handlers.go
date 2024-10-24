package handlers

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/xml"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"triple-s/internal/models"
	"triple-s/internal/services"
)

// ErrorResponse defines the structure of the error response in XML format
type ErrorResponse struct {
	XMLName xml.Name `xml:"error"`
	Message string   `xml:"message"`
}

// writeErrorResponse marshals the error message into XML and writes it to the response
func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)
	errorResponse := ErrorResponse{Message: message}
	xmlData, err := xml.MarshalIndent(errorResponse, "", "  ")
	if err != nil {
		log.Println("Error generating XML response:", err)
		w.Write([]byte("<error>Internal Server Error</error>"))
		return
	}
	w.Write(xmlData)
}

// HandlePutBuckets handles PUT requests for creating a bucket
func HandlePutBuckets(w http.ResponseWriter, r *http.Request, directoryPath string) {
	// Ensure the directory exists
	err := os.MkdirAll(directoryPath, os.ModePerm)
	if err != nil {
		writeErrorResponse(w, "Cannot create directory", http.StatusInternalServerError)
		return
	}

	// Extract bucket name from the URL path
	bucketName := strings.TrimPrefix(r.URL.Path, "/")
	// Validate bucket name
	if !ValidateBucketName(bucketName) {
		writeErrorResponse(w, "Not a valid bucket name", http.StatusBadRequest)
		return
	}

	// Call service to create the bucket and handle errors
	if err := services.BucketAndFileCreation(directoryPath + bucketName); err != nil {
		writeErrorResponse(w, err.Error(), http.StatusConflict)
		return
	}

	// Write bucket info and return XML response, handle errors
	xmlData, err := services.WriteBucketInfo(bucketName, directoryPath)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response type to XML and send response
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xmlData))
}

// HandleDeleteBuckets handles DELETE requests for deleting a bucket
func HandleDeleteBuckets(w http.ResponseWriter, r *http.Request, directoryPath string) {
	// Extract bucket name from the URL path
	bucketName := strings.TrimPrefix(r.URL.Path, "/")
	if bucketName == "buckets.csv" {
		w.Write([]byte("Can not delete metadata file"))
		return
	}
	if bucketName == "" {
		writeErrorResponse(w, "Empty bucket name", http.StatusBadRequest)
		return
	}

	// Open the CSV file to read bucket data
	file, err := os.Open(directoryPath + "buckets.csv")
	if err != nil {
		writeErrorResponse(w, "Error reading bucket file", http.StatusInternalServerError)
		log.Println("Error reading CSV:", err)
		return
	}
	defer file.Close()

	// Read CSV records
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		writeErrorResponse(w, "Error reading CSV data", http.StatusInternalServerError)
		log.Println("Error reading CSV:", err)
		return
	}

	// Check if the bucket exists by looking at the directory
	fileInfo, err := os.Stat(directoryPath + bucketName)
	if err != nil {
		writeErrorResponse(w, "Bucket does not exist", http.StatusNotFound)
		return
	}

	// Encode modification time
	encodedModTime := base64.StdEncoding.EncodeToString([]byte(fileInfo.ModTime().Format(time.RFC3339)))

	// Update bucket status in the CSV records
	for i, record := range records {
		val, _ := base64.StdEncoding.DecodeString(record[0])
		if string(val) == bucketName {
			records[i][3] = "false" // Mark the bucket as inactive
			records[i][2] = encodedModTime
		}
	}

	// Delete the bucket directory
	err = os.RemoveAll(directoryPath + bucketName)
	if err != nil {
		writeErrorResponse(w, "Error deleting bucket", http.StatusInternalServerError)
		log.Println("Error deleting bucket:", err)
		return
	}

	// Open the CSV file to write updated data
	bucketFile, err := os.OpenFile(directoryPath+"buckets.csv", os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		writeErrorResponse(w, "Error opening bucket file", http.StatusInternalServerError)
		log.Println("Error opening file for writing:", err)
		return
	}
	defer bucketFile.Close()

	// Write updated records back to the CSV file
	writer := csv.NewWriter(bucketFile)
	err = writer.WriteAll(records)
	if err != nil {
		writeErrorResponse(w, "Error writing CSV data", http.StatusInternalServerError)
		log.Println("Error writing CSV:", err)
		return
	}

	// Return success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<message>Bucket successfully deleted</message>"))
}

// HandleGetBuckets handles GET requests for listing buckets
func HandleGetBuckets(w http.ResponseWriter, r *http.Request, directoryPath string) {
	// Ensure the request is for the root path
	if r.URL.Path != "/" {
		writeErrorResponse(w, "Not found", http.StatusNotFound)
		return
	}

	// Open the buckets.csv file
	file, err := os.Open(directoryPath + "buckets.csv")
	if err != nil {
		writeErrorResponse(w, "No such bucket", http.StatusConflict)
		return
	}
	defer file.Close()

	// Read all records from the CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		writeErrorResponse(w, "Error reading CSV file", http.StatusInternalServerError)
		return
	}

	// Set the response type to XML
	w.Header().Set("Content-Type", "application/xml")

	// Create a slice to hold the bucket models
	var buckets []models.Bucket

	// Iterate through each record and convert it into a Bucket model
	for _, record := range records {
		// Decode base64-encoded fields
		name, err := base64.StdEncoding.DecodeString(record[0])
		if err != nil {
			writeErrorResponse(w, "Error decoding bucket name", http.StatusInternalServerError)
			return
		}
		creationTime, err := base64.StdEncoding.DecodeString(record[1])
		if err != nil {
			writeErrorResponse(w, "Error decoding creation time", http.StatusInternalServerError)
			return
		}
		lastModifiedTime, err := base64.StdEncoding.DecodeString(record[2])
		if err != nil {
			writeErrorResponse(w, "Error decoding last modified time", http.StatusInternalServerError)
			return
		}
		status := record[3]

		// Create a Bucket model
		localBucket := models.Bucket{
			Name:             string(name),
			CreationTime:     string(creationTime),
			LastModifiedTime: string(lastModifiedTime),
			Status:           status,
		}

		buckets = append(buckets, localBucket)
	}

	// Marshal the slice of buckets into XML
	xmlData, err := xml.MarshalIndent(buckets, "", "  ")
	if err != nil {
		writeErrorResponse(w, "Error generating XML", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Write XML data to the response
	w.Write(xmlData)
}
