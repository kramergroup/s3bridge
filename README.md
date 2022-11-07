# s3bridge

This project provides a simple S3 bridge service that allows to access objects in a backing S3 bucket via HTML requests.

The service provides two access modes:

1. **Proxy Mode**: In proxy mode, the service acts as man in the middle. It reads the content of objects from the backing s3 service and passes them on as responses to HTTP requests. This has the advantage that the backing s3 does not need to be exposed to clients, but it can put resource pressure on the service, because all data is read into memory before passing on to the client. Proxy mode does support range requests, which should go a long way reducing memory pressure on the service.

2. **Pre-Signed URL Mode**: In pre-signed URL mode, the service does not proxy the content of s3 objects. Rather, it generates pre-signed URLs to directly access the objects from the backing s3 server and returns HTML redirect responses pointing towards these. This has the advantage of not passing actual objects through the service, but requires exposing the backing s3 instance.

Note that pre-signed URL mode does not work with media assetts served as HLS, because this requires accessing a number of files defined in a .m3u8 file. This breaks the pre-signing, because the files are directly accessed on the s3 endpoint, but without pre-signing.

## Compilation

```linux
go mod tidy
go build -o server cmd/server/main.go
```

## Usage

1. Issue a `GET` request against the service, e.g. `http://localhost:8080/README.md` to retrieve the object with key `README.md`
2. The service will return a *Temporarily Moved (302)* response providing a pre-signed URL to obtain the object from the backing S3 service.
3. Most clients will automatically follow this redirect to access the object. If not, the provided URL needs o be called separately.

## Configuration

| Option                  | Description                                   | Default       |
| ----------------------- | --------------------------------------------- | ------------- |
| `bucket`                | Name of the S3 bucket backing the bridge      | none/required |
| `aws-access-key-id`     | The AWS access key id                         | none/required |
| `aws-secret-access-key` | The AWS secret key                            | none/required |
| `aws-region`            | The AWS region                                | us-west-2     |
| `endpoint`              | The S3 endpoint for the bucket                | none          |
| `expiry-time`           | The expiry time of the issued pre-signed URLs | 2h            |
| `proxy-port`            | The port to listen to for proxy requests      | 80            |
| `presign-port`          | The port to listen to for pre-sign requests   | 8080          |
| `loglevel`              | The log-level (ERROR, WARN, INFO, DEBUG)      | INFO          |

Settign either `port` or `presign-port` to zero disables the mode, respectively.
