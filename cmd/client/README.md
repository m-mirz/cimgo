# CIMgo CLI Client Usage

This document explains how to use the command-line interface client to interact with the CIMgo webserver. The client provides a more convenient way to upload CIM files, trigger processing, and retrieve merged data than using `curl` directly.

## Running the Client

Navigate to the project's root directory and use `go run` to execute the client.

```bash
go run ./cmd/cim-client [OPTIONS] <command> [arguments...]
```

## Common Options

* `-id <model_id>`: **Required**. Specifies a unique identifier for your CIM model instance on the server. All operations (upload, get) are performed in the context of this ID.

## Commands

### `upload <file_path...>`

Uploads one or more CIM files that are compressed into a ZIP archive to the server, associating them with the specified model ID. These files are stored temporarily on the server awaiting processing.

* **Arguments**: One or more paths to the CIM files you want to upload.

**Example:**

First, ensure your webserver is running (`go run ./cmd/webserver`).

Now, upload a CIM ZIP archive using the client:

```bash
go run ./cmd/client/main.go -id my_model_1 upload equipment.zip
```

### 3. `get`

Retrieves the merged CIM specification data for the specified model ID from the server. The data is returned in JSON format.

* **Arguments**: None.

**Example:**

```bash
go run ./cmd/cim-client -id my_model_1 get
```
