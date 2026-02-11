package service

import (
	"context"
	"testing"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/proxy/mocks"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func initEnv() {
	config.Config = &config.Conf{
		S3Config:  &config.S3Config{},
		LogConfig: &config.LogConfig{Level: "debug"},
	}
}
func TestGenUploadURL(t *testing.T) {
	initEnv()
	ctx := context.Background()
	s3Proxy := new(mocks.S3Proxy)
	imageService := &ImageServiceImpl{
		s3Proxy: s3Proxy,
	}

	tests := []struct {
		imageType string
		expectErr bool
	}{
		{"jpg", false},
		{"png", false},
		{"jpeg", false},
		{"", true},
	}
	s3Proxy.On("GenPutPresignRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&v4.PresignedHTTPRequest{URL: "https://signedurl.com"}, nil)
	for _, tt := range tests {
		t.Run(tt.imageType, func(t *testing.T) {
			resp, err := imageService.GenUploadURL(ctx, tt.imageType)
			if (err != nil) != tt.expectErr {
				t.Errorf("GenUploadURL() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && resp == nil {
				t.Errorf("GenUploadURL() response = %v, want non-nil", resp)
			}
			if err == nil {
				assert.NotEmpty(t, resp.UploadURL)
				assert.NotEmpty(t, resp.ImageId)
			}
		})
	}
}
func TestGetImageService(t *testing.T) {
	initEnv()

	// First call to GetImageService
	service1 := GetImageService()

	// Second call to GetImageService
	service2 := GetImageService()

	// Ensure that both calls return the same instance
	if service1 != service2 {
		t.Errorf("GetImageService() returned different instances")
	}

	// Ensure that the instance is not nil
	if service1 == nil {
		t.Errorf("GetImageService() returned nil instance")
	}
}
