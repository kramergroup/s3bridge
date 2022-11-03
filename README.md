# s3bridge

This project provides a simple S3 bridge service that allows to access objects in a backing S3 bucket via HTML requests.

The service is *no* proxy. Rather, it generates pre-signed URLs and returns HTML redirect responses pointing towards these.

## Compilation

```linux
go mod tidy
go build -o server cmd/server/main.go
```

## Configuration

| Option                  | Description                                   | Default       |
| ----------------------- | --------------------------------------------- | ------------- |
| `bucket`                | Name of the S3 bucket backing the bridge      | none/required |
| `aws-access-key-id`     | The AWS access key id                         | none/required |
| `aws-secret-access-key` | The AWS secret key                            | none/required |
| `aws-region`            | The AWS region                                | us-west-2     |
| `endpoint`              | The S3 endpoint for the bucket                | none          |
| `expiry-time`           | The expiry time of the issued pre-signed URLs | 2h            |
| `port`                  | The port to listen to                         | 8080          |
| `loglevel`              | The log-level (ERROR, WARN, INFO, DEBUG)      | INFO          |
