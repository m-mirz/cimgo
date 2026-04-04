# CIMgo Webserver

This document outlines how to interact with the CIMgo webserver using `curl`. The webserver provides endpoints for uploading CIM (Common Information Model) files, and retrieving the merged in-memory representation.

## Starting the Webserver

First, ensure the webserver is running. Navigate to the project root and execute:

```bash
go run ./cmd/server
```

## Endpoints

### Upload CIM Files (`POST /cim/{id}`)

This endpoint allows you to upload individual CIM files (e.g., RDF files) associated with a specific ID. The files will be stored in a temporary directory on the server.

* **Method**: `POST`
* **URL**: `http://localhost:8080/cim/{your_id}`
* **Content-Type**: `multipart/form-data`

**Example:**

Let's assume you have a file named `my_cim_archive.zip`.

Now, upload the files:

```bash
curl -X POST -F "file=@my_cim_archive.zip" http://localhost:8080/cim/test_model_1
```

### Retrieve Merged Data (`GET /cim/{id}`)

Once the files are processed, you can retrieve the merged CIM specification in JSON format using this endpoint.

*   **Method**: `GET`
*   **URL**: `http://localhost:8080/cim/{your_id}`

**Example:**

```bash
curl http://localhost:8080/cim/test_model_1
```
