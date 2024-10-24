package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"triple-s/internal/handlers"
)

var (
	portNumber    string
	directoryPath string
)

func main() {
	parseFlags()
	if directoryPath[len(directoryPath)-1] != '/' {
		directoryPath += "/"
	}
	mux := http.NewServeMux()
	args := os.Args[1:]
	for _, v := range args {
		if v == "--help" || v == "--h" {
			flag.Usage()
			return
		}
	}
	// Handle root requests for bucket actions
	mux.HandleFunc("/", rootHandler)

	// Start server on the configured port
	correctPort, _ := strconv.Atoi(portNumber)
	if !(correctPort >= 1024 && correctPort <= 49151) {
		fmt.Println("Incorrect port number")
		portNumber = "8080"
	}
	log.Printf("Server running on port %s...\n", portNumber)
	log.Fatal(http.ListenAndServe(":"+portNumber, mux))
}

var helpUsage string = `Simple Storage Service.

**Usage:**
    triple-s [-port <N>] [-dir <S>]  
    triple-s --help

**Options:**
- --help     Show this screen.
- --port N   Port number
- --dir S    Path to the directory`

// parseFlags reads command-line flags for configuration
func parseFlags() {
	flag.StringVar(&portNumber, "port", "8080", "Port number for the server")
	flag.StringVar(&directoryPath, "dir", "data/", "Directory path to store bucket data")
	flag.Usage = func() {
		fmt.Println(helpUsage)
	}
	flag.Parse()

	log.Println(directoryPath, portNumber)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	pathComponents := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathComponents) == 1 && pathComponents[0] != "" {
		bucketHandler(w, r)
	} else if len(pathComponents) == 2 {
		objectHandler(w, r, pathComponents[0], pathComponents[1])
	} else {
		if r.URL.Path == "/" {
			switch r.Method {
			case http.MethodGet:
				handlers.HandleGetBuckets(w, r, directoryPath)
			default:
				http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Invalid path", http.StatusBadRequest)
		}
	}
}

// bucketHandler handles actions related to the bucket
func bucketHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.HandleGetBuckets(w, r, directoryPath)
	case http.MethodPut:
		handlers.HandlePutBuckets(w, r, directoryPath)
	case http.MethodDelete:
		handlers.HandleDeleteBuckets(w, r, directoryPath)
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}

// objectHandler handles actions related to an object inside a bucket
func objectHandler(w http.ResponseWriter, r *http.Request, bucketName, objectKey string) {
	switch r.Method {
	case http.MethodGet:
		handlers.HandlerGetObject(w, r, directoryPath, bucketName, objectKey)
	case http.MethodPut:
		handlers.HandlerPutObject(w, r, directoryPath, bucketName, objectKey)
	case http.MethodDelete:
		handlers.HandlerDeleteObject(w, r, directoryPath, bucketName, objectKey)
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}
