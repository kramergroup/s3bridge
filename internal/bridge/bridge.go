package bridge

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
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

type s3ReadSeeker struct {
	currentPos *int64
	size       *int64
	ctx        context.Context
	objParams  *s3.GetObjectInput
	client     *s3.Client
}

func (b S3Bridge) GetRequestURL(assett string) (*url.URL, error) {

	// Remove any leading lash from key
	assett = strings.TrimLeft(assett, "/")

	client := b.s3client()

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

func (b S3Bridge) StreamObject(assett string, ctx context.Context, writer io.WriterAt) error {

	// Remove any leading lash from key
	assett = strings.TrimLeft(assett, "/")

	streamParams := &s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(assett),
	}

	client := b.s3client()

	downloader := manager.NewDownloader(client)

	_, err := downloader.Download(ctx, writer, streamParams)
	return err
}

func (b S3Bridge) ReadSeeker(assett string, ctx context.Context) (io.ReadSeeker, error) {

	// Remove any leading lash from key
	assett = strings.TrimLeft(assett, "/")

	// Find the size of the object
	headObj := s3.HeadObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(assett),
	}

	client := b.s3client()

	result, err := client.HeadObject(ctx, &headObj)
	if err != nil {
		return nil, err
	}
	size := aws.Int64(result.ContentLength)

	// Prepare object get ReadSeeker
	objParams := &s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(assett),
	}

	pos := int64(0)
	return s3ReadSeeker{currentPos: &pos, size: size, objParams: objParams, ctx: ctx, client: b.s3client()}, nil

}

func (b S3Bridge) s3client() *s3.Client {

	// Prepare the S3 request so a signature can be generated
	opts := s3.Options{
		UsePathStyle: true,
		Credentials:  credentials.NewStaticCredentialsProvider(b.S3AccessKey, b.S3SecretKey, ""),
		Region:       b.Region,
	}

	if b.Endpoint != "" {
		opts.EndpointResolver = s3.EndpointResolverFromURL(b.Endpoint)
	}

	return s3.New(opts)

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

func (rs s3ReadSeeker) Read(p []byte) (int, error) {

	downloader := manager.NewDownloader(rs.client)

	chunk := int64(len(p))
	if chunk+(*rs.currentPos) > *rs.size {
		chunk = *rs.size - *rs.currentPos
	}

	begin := *rs.currentPos
	end := begin + chunk - 1

	rangeStr := "bytes=" + strconv.FormatInt(begin, 10) + "-" + strconv.FormatInt(end, 10)

	para := &s3.GetObjectInput{
		Bucket: aws.String(*rs.objParams.Bucket),
		Key:    aws.String(*rs.objParams.Key),
		Range:  aws.String(rangeStr),
	}

	buf := manager.NewWriteAtBuffer(p)
	nr, err := downloader.Download(rs.ctx, buf, para)

	// Shift the current position to the end of the current chunk
	*rs.currentPos += chunk

	return int(nr), err
}

func (rs s3ReadSeeker) Seek(offset int64, whence int) (int64, error) {

	switch whence {
	case io.SeekStart:
		if offset > *rs.size {
			return 0, errors.New("offset larger than object")
		}
		*rs.currentPos = offset
	case io.SeekCurrent:
		if *rs.currentPos+offset > *rs.size {
			return 0, errors.New("offset points beyond offset")
		}
		*rs.currentPos += offset
	case io.SeekEnd:
		if offset > *rs.size {
			return 0, errors.New("offset larger than object")
		}
		*rs.currentPos = *rs.size - offset
	}

	return *rs.currentPos, nil
}
