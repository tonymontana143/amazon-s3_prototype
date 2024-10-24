package handlers

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"triple-s/internal/models"
	"triple-s/internal/services"
)

// XMLResponse represents a standard XML response format for errors and messages.
type XMLResponse struct {
	XMLName xml.Name `xml:"response"`
	Message string   `xml:"message"`
}

// WriteXMLResponse writes an XML response with the provided status code and message.
func WriteXMLResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/xml")

	response := XMLResponse{Message: message}
	xmlData, err := xml.MarshalIndent(response, "", "  ")
	if err != nil {
		// In case of an error while marshaling XML, fall back to plain text
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(xmlData)
}

func HandlerPutObject(w http.ResponseWriter, r *http.Request, directoryPath string, bucketName, objectKey string) {
	defer r.Body.Close()

	// Validate the URL structure
	count := strings.Count(r.URL.Path, "/")
	if count != 2 {
		WriteXMLResponse(w, http.StatusBadRequest, "Invalid URL path format")
		return
	}

	// Check if the bucket exists
	bucketPath := directoryPath + bucketName
	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		WriteXMLResponse(w, http.StatusNotFound, "Bucket not found")
		return
	}

	// Create the object
	objectPath := bucketPath + "/" + objectKey
	newFile, err := os.Create(objectPath)
	if err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Error creating object")
		return
	}
	defer newFile.Close()

	// Write object data from the request body to the file
	if _, err := io.Copy(newFile, r.Body); err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Error writing object data")
		return
	}

	// Store object metadata
	if err := services.WriteObjectInfo(r, directoryPath, bucketName, objectKey); err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Error writing object info: "+err.Error())
		return
	}

	// Respond with success
	WriteXMLResponse(w, http.StatusOK, "Object created successfully")
}

func HandlerDeleteObject(w http.ResponseWriter, r *http.Request, directoryPath string, bucketName, objectKey string) {
	// Prevent deletion of the objects.csv metadata file
	if objectKey == "objects.csv" {
		WriteXMLResponse(w, http.StatusForbidden, "Cannot delete metadata file")
		return
	}

	bucketPath := directoryPath + bucketName

	// Check if the bucket exists
	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		WriteXMLResponse(w, http.StatusNotFound, "Bucket does not exist")
		return
	}

	// Check if the object exists
	objectPath := bucketPath + "/" + objectKey
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		WriteXMLResponse(w, http.StatusNotFound, "Object does not exist")
		return
	}

	// Open the objects CSV file
	file, err := os.Open(bucketPath + "/objects.csv")
	if err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Cannot open metadata file")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Cannot read metadata file")
		return
	}

	// Filter out the object to delete
	var newRecords [][]string
	for _, v := range records {
		decodedName, err := base64.StdEncoding.DecodeString(v[0])
		if err != nil {
			WriteXMLResponse(w, http.StatusInternalServerError, "Error decoding object name")
			return
		}
		if string(decodedName) != objectKey {
			newRecords = append(newRecords, v)
		} else {
			// Delete the actual object file
			if err := os.Remove(objectPath); err != nil {
				WriteXMLResponse(w, http.StatusInternalServerError, "Error deleting object")
				return
			}
		}
	}

	// Write the updated records back to the objects CSV file
	file, err = os.Create(bucketPath + "/objects.csv")
	if err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Error writing metadata file")
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.WriteAll(newRecords); err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Error writing to CSV file")
		return
	}
	writer.Flush()

	// Respond with success
	WriteXMLResponse(w, http.StatusOK, "Object successfully deleted")
}

// HandlerGetObject handles retrieving an object.
func HandlerGetObject(w http.ResponseWriter, r *http.Request, directoryPath string, bucketName, objectKey string) {
	bucketPath := directoryPath + bucketName
	// Check if bucket exists
	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		WriteXMLResponse(w, http.StatusNotFound, "Bucket does not exist")
		return
	}
	contType := r.Header.Get("Content-Type")
	// Check if the object exists
	objectPath := bucketPath + "/" + objectKey
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		WriteXMLResponse(w, http.StatusNotFound, "Object does not exist")
		return
	}

	// Open the objects CSV file
	file, err := os.Open(bucketPath + "/objects.csv")
	if err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Cannot open objects metadata file")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var localObject models.Object
	records, err := reader.ReadAll()
	if err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Cannot read metadata file")
		return
	}

	// Find the object metadata
	for _, v := range records {
		name, err := base64.StdEncoding.DecodeString(v[0])
		if err != nil {
			WriteXMLResponse(w, http.StatusInternalServerError, "Error decoding object name")
			return
		}
		if string(name) == objectKey {
			localObject.ObjectKey = objectKey
			size, _ := strconv.Atoi(v[1])
			localObject.Size = int64(size)
			localObject.ContentType = contType
			modTime, _ := base64.StdEncoding.DecodeString(v[3])
			localObject.LastModifiedTime = string(modTime)
		}
	}

	// Marshal the object info to XML
	x, err := xml.MarshalIndent(localObject, "", " ")
	if err != nil {
		WriteXMLResponse(w, http.StatusInternalServerError, "Error generating XML")
		return
	}

	// Set content type and write the response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}
