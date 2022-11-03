############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Git is required for fetching the dependencies.
RUN apk update
RUN apk add --no-cache git

WORKDIR $GOPATH/src/github.com/kramergroup/s3bridge

COPY . .

# Fetch dependencies.# Using go get.
RUN go mod download
RUN go build -o /server cmd/server/main.go
RUN chmod a+x /server
RUN ls -la /

# STEP 2 build a small image
############################
FROM alpine

# Copy our static executable.
COPY --from=builder /server /server

# Run the hello binary.
ENTRYPOINT ["/server"]