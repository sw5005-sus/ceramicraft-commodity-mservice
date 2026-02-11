package proxy

import (
	"context"
	"sync"
	"time"

	local_config "github.com/sw5005-sus/ceramicraft-commodity-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Client       *s3.Client
	S3ProxyInst    S3Proxy
	s3InitSyncOnce sync.Once
)

func InitS3Client() {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Logger.Errorf("Couldn't load default configuration. err: %v", err)
		return
	}
	sdkConfig.Region = local_config.Config.S3Config.Region
	s3Client = s3.NewFromConfig(sdkConfig)

}

type S3Proxy interface {
	GenPutPresignRequest(ctx context.Context, bucketName string, objectKey string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error)
}

type S3ProxyImpl struct {
	presignClient *s3.PresignClient
}

func GetPresigner() S3Proxy {
	s3InitSyncOnce.Do(func() {
		if s3Client == nil {
			InitS3Client()
		}
		S3ProxyInst = &S3ProxyImpl{
			presignClient: s3.NewPresignClient(s3Client),
		}
	})
	return S3ProxyInst
}

// PutObject makes a presigned request that can be used to put an object in a bucket.
// The presigned request is valid for the specified number of seconds.
func (s S3ProxyImpl) GenPutPresignRequest(
	ctx context.Context, bucketName string, objectKey string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error) {
	request, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		log.Logger.Errorf("Couldn't get a presigned request to put %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request, err
}
