module github.com/kramergroup/s3bridge

go 1.19

require github.com/sirupsen/logrus v1.9.0

require (
	github.com/aws/aws-sdk-go-v2 v1.17.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.9 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.19 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.29.1 // indirect
	github.com/aws/smithy-go v1.13.4 // indirect
	golang.org/x/sys v0.1.0 // indirect
)

require internal/bridge v1.0.0

require (
	golang.org/x/exp v0.0.0-20221031165847-c99f073a8326
	internal/util v1.0.0
)

replace internal/bridge => ./internal/bridge

replace internal/util => ./internal/util