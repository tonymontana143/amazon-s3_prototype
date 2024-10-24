# Triple-S (Simple Storage Service)

## Overview

The **Triple-S** project is a basic implementation of a RESTful service inspired by Amazon S3, designed to handle file storage (objects) in containers known as buckets. The project provides endpoints for managing both buckets and objects, with operations such as creating, listing, and deleting buckets, as well as uploading, retrieving, and deleting objects. All metadata related to these operations is stored in CSV files.

## Table of Contents

- [Features](#features)
- [Usage](#usage)
- [Installation](#installation)
- [API Endpoints](#api-endpoints)
- [Error Handling](#error-handling)
- [Directory Structure](#directory-structure)
- [Example Scenarios](#example-scenarios)
- [Contributions](#contributions)

## Features

1. **Basic HTTP Server**: A Go-based HTTP server that listens on a configurable port and responds to RESTful API requests.
2. **Bucket Management**:
    - Create, list, and delete buckets.
    - Conforms to Amazon S3's bucket naming conventions.
3. **Object Operations**:
    - Upload, retrieve, and delete objects (files).
    - Object metadata is stored in CSV files.
4. **Metadata Management**: Buckets and objects are tracked through CSV files storing creation time, modification time, and other necessary metadata.
5. **Error Handling**: Graceful error handling with meaningful HTTP status codes.
6. **No Authentication/Authorization**: The service operates without credentials for simplicity.

## Usage

The **Triple-S** application can be configured with a port number and a base directory where files will be stored.

```bash
$ ./triple-s [-port <N>] [-dir <S>]
$ ./triple-s --help
```

### Options:

- `--help`: Displays the help information for the program.
- `-port N`: Specifies the port number for the HTTP server. Defaults to 8080 if not provided.
- `-dir S`: Specifies the directory path where buckets and objects will be stored. Defaults to `./data` if not provided.

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/username/triple-s.git
    cd triple-s
    ```

2. Build the project:
    ```bash
    go build -o triple-s
    ```

3. Run the server:
    ```bash
    ./triple-s -port 8080 -dir /path/to/data
    ```

## API Endpoints

### Bucket Management

#### 1. Create a Bucket
- **Method**: `PUT`
- **Endpoint**: `/{BucketName}`
- **Response**:
    - Success: `200 OK`
    - Errors: `400 Bad Request` (Invalid bucket name), `409 Conflict` (Bucket already exists)

#### 2. List All Buckets
- **Method**: `GET`
- **Endpoint**: `/`
- **Response**: 
    - Success: `200 OK` with XML list of buckets.
    - Error: `500 Internal Server Error`

#### 3. Delete a Bucket
- **Method**: `DELETE`
- **Endpoint**: `/{BucketName}`
- **Response**:
    - Success: `204 No Content`
    - Errors: `404 Not Found` (Bucket does not exist), `409 Conflict` (Bucket not empty)

### Object Operations

#### 1. Upload a New Object
- **Method**: `PUT`
- **Endpoint**: `/{BucketName}/{ObjectKey}`
- **Request**: Binary data of the object in the request body.
- **Response**:
    - Success: `200 OK`
    - Errors: `404 Not Found` (Bucket does not exist)

#### 2. Retrieve an Object
- **Method**: `GET`
- **Endpoint**: `/{BucketName}/{ObjectKey}`
- **Response**:
    - Success: Returns the binary content of the object.
    - Errors: `404 Not Found` (Object or bucket does not exist)

#### 3. Delete an Object
- **Method**: `DELETE`
- **Endpoint**: `/{BucketName}/{ObjectKey}`
- **Response**:
    - Success: `204 No Content`
    - Errors: `404 Not Found` (Object does not exist)

## Error Handling

- **400 Bad Request**: Invalid bucket or object names.
- **404 Not Found**: Requested bucket or object does not exist.
- **409 Conflict**: Conflicting operations (e.g., attempting to delete a non-empty bucket).
- **500 Internal Server Error**: General server error for unexpected issues (e.g., file system access issues).

## Directory Structure

All data is stored in a base directory, with subdirectories for each bucket. Objects are stored as files in their respective bucket directories, and each bucket maintains a CSV file to track its objects.

```bash
data/
    └── {bucket-name}/
        ├── objects.csv
        ├── {object-key}
```

### Example Structure:
```bash
data/
    └── photos/
        ├── objects.csv
        ├── sunset.png
        └── beach.jpg
```

## Example Scenarios

### Bucket Creation

```bash
PUT /my-bucket
```
Response: `200 OK` if successful, or `400 Bad Request` for invalid names.

### List Buckets

```bash
GET /
```
Response: XML containing all buckets.

### Object Upload

```bash
PUT /my-bucket/photo.png
```
Uploads the object to `data/my-bucket/photo.png`.

## Contributions

Feel free to fork this repository and submit a pull request with your changes or improvements. Contributions are always welcome!
