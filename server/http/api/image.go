package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/service"
)

// GetImageUploadPresignURL godoc
// @Summary Get presigned URL for image upload
// @Description Get presigned URL for image upload
// @Tags Image
// @Accept json
// @Produce json
// @Param product body data.ImgUploadRequest true "image_type=(jpg|jpeg|png)"
// @Param   client  path   string  true  "client type" Enums(customer, merchant)
// @Success 200 {object} data.ImgUploadResponse
// @Failure 400 {object} data.BaseResponse
// @Router /{client}/images/upload-urls [post]
func GetImageUploadPresignURL(c *gin.Context) {
	var req data.ImgUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger.Errorf("GetImageUploadPresignURL: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed(err.Error()))
		return
	}
	ret, err := service.GetImageService().GenUploadURL(c.Request.Context(), req.ImageType)
	if err != nil {
		log.Logger.Errorf("GetImageUploadPresignURL: Failed to generate upload url: %v", err)
		c.JSON(http.StatusInternalServerError, data.ResponseFailed("Failed to generate image uplaod url"))
		return
	}
	c.JSON(http.StatusOK, ret)
}
