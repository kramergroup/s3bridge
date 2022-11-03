package bridge

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	log "github.com/sirupsen/logrus"
)

type S3Bridge struct {
	Bucket      string        // bucket name
	S3SecretKey string        // s3 secret access key
	S3AccessKey string        // s3 access key
	Endpoint    string        // custom s3 endpoint URL
	Region      string        // aws region
	ExpiryTime  time.Duration // Exiry time for the pre-signed URL
}

func (b S3Bridge) GetRequestURL(assett string) (*url.URL, error) {

	// Remove any leading lash from key
	assett = strings.TrimLeft(assett, "/")

	// Prepare the S3 request so a signature can be generated
	opts := s3.Options{
		UsePathStyle: true,
		Credentials:  credentials.NewStaticCredentialsProvider(b.S3AccessKey, b.S3SecretKey, ""),
		Region:       b.Region,
	}

	if b.Endpoint != "" {
		opts.EndpointResolver = s3.EndpointResolverFromURL(b.Endpoint)
	}

	client := s3.New(opts)

	presignClient := s3.NewPresignClient(client)

	presignParams := &s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(assett),
	}

	// Apply an expiration via an option function
	presignDuration := func(po *s3.PresignOptions) {
		po.Expires = b.ExpiryTime
	}

	presignResult, err := presignClient.PresignGetObject(context.TODO(), presignParams, presignDuration)
	if err != nil {
		return nil, err
	}

	return url.Parse(presignResult.URL)
}

func (b S3Bridge) PutRequestURL(assett string) (*url.URL, error) {

	return nil, errors.New("not yet implemented [S3Bridge.PutRequestURL]")

}

// Validate validates the configuration of the S3Bridge. The function throws an error
// if the configuration is invalid
func (b S3Bridge) Validate() error {

	// Critical errors
	if b.Bucket == "" {
		return errors.New("no bucket specified")
	}

	if b.S3AccessKey == "" {
		return errors.New("no AWS access key id specifiec")
	}

	if b.S3SecretKey == "" {
		return errors.New("no AWS secret key specified")
	}

	if b.ExpiryTime <= 0 {
		return errors.New("expiry time must be larger than zero")
	}

	// Log warnings for potentially unexpected behaviour
	if b.Endpoint == "" {
		log.Debug("standard ASW endpoint will be used")
	} else {
		log.Debug("endpoint %s will be used", b.Endpoint)
	}

	if b.Region == "" {
		log.Warn("aws region is empty")
	}

	return nil
}
