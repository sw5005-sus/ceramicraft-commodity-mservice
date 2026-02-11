package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/proxy"
)

type ImageService interface {
	GenUploadURL(ctx context.Context, imageType string) (*data.ImgUploadResponse, error)
}

type ImageServiceImpl struct {
	s3Proxy proxy.S3Proxy
}

var (
	imageServiceInst ImageService
	imageOnce        sync.Once
)

func GetImageService() ImageService {
	imageOnce.Do(func() {
		imageServiceInst = &ImageServiceImpl{
			s3Proxy: proxy.GetPresigner(),
		}
	})
	return imageServiceInst
}

const (
	lifetimeSecs = 15 * 60 // 15 minutes
)

var (
	supportedImageTypes = map[string]bool{
		"jpg":  true,
		"png":  true,
		"jpeg": true,
	}
)

func (i *ImageServiceImpl) GenUploadURL(ctx context.Context, imageType string) (*data.ImgUploadResponse, error) {
	if _, exist := supportedImageTypes[imageType]; !exist {
		return nil, fmt.Errorf("unsupported image type: %s", imageType)
	}
	timeStamp := time.Now().UnixNano()
	objectKey := fmt.Sprintf("%x.%s", timeStamp, imageType)
	s3PresignReq, err := i.s3Proxy.GenPutPresignRequest(
		ctx, config.Config.S3Config.BucketName, objectKey, lifetimeSecs)
	if err != nil {
		log.Logger.Errorf("Failed to generate presign URL for object %s: %v", objectKey, err)
		return nil, err
	}
	return &data.ImgUploadResponse{UploadURL: s3PresignReq.URL, ImageId: objectKey}, nil
}
