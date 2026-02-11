package api

import (
	"net/http"
	"strconv"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/service"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/types"
	"github.com/gin-gonic/gin"
)

// AddProduct godoc
// @Summary 添加商品
// @Description 新增一个商品
// @Tags 商品
// @Accept json
// @Produce json
// @Param product body types.ProductInfo true "商品信息"
// @Success 200 {object} data.BaseResponse
// @Failure 400 {object} data.BaseResponse
// @Router /merchant/products [post]
func AddProduct(c *gin.Context) {
	var req types.ProductInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger.Errorf("AddProduct: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed(err.Error()))
		return
	}
	productId, err := service.GetProductServiceInstance().Create(c.Request.Context(), &req)
	if err != nil {
		log.Logger.Errorf("AddProduct: Failed to create product: %v", err)
		c.JSON(http.StatusInternalServerError, data.ResponseFailed("Failed to create product"))
		return
	}
	c.JSON(http.StatusOK, data.ResponseSuccess(productId))
}

// GetProductMerchant godoc
// @Summary 获取商品详情
// @Description 根据商品ID获取商品详细信息
// @Tags 商品
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Success 200 {object} data.BaseResponse{data=types.ProductInfo} "成功"
// @Failure 400 {object} data.BaseResponse "请求参数错误"
// @Failure 404 {object} data.BaseResponse "商品不存在"
// @Failure 500 {object} data.BaseResponse "服务器内部错误"
// @Router /merchant/product/{id} [get]
func GetProductMerchant(c *gin.Context) {
	// 解析路径参数
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Logger.Errorf("GetProduct: Invalid product ID: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid product ID"))
		return
	}

	// 调用 service 层获取商品信息
	product, err := service.GetProductServiceInstance().GetProductByID(c.Request.Context(), id)
	if err != nil {
		log.Logger.Errorf("GetProduct: Failed to get product details: %v", err)
		c.JSON(http.StatusInternalServerError, data.ResponseFailed("Failed to get product details"))
		return
	}

	// 如果没找到商品
	if product == nil {
		c.JSON(http.StatusNotFound, data.ResponseFailed("Product not found"))
		return
	}

	// 返回商品信息
	c.JSON(http.StatusOK, data.ResponseSuccess(product))
}

// PublishProduct godoc
// @Summary 上架商品
// @Description 将商品状态更改为上架状态
// @Tags 商品
// @Accept json
// @Produce json
// @Param request body types.UpdateProductStatusRequest true "商品上架请求"
// @Success 200 {object} data.BaseResponse "上架成功"
// @Failure 400 {object} data.BaseResponse "请求参数错误"
// @Failure 404 {object} data.BaseResponse "商品不存在"
// @Failure 500 {object} data.BaseResponse "服务器内部错误"
// @Router /merchant/products/:id/status [patch]
func UpdateProductStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Logger.Errorf("GetProduct: Invalid product ID: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid product ID"))
		return
	}

	var req types.UpdateProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger.Errorf("PublishProduct: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed(err.Error()))
		return
	}

	if req.Status == 1 {
		err := service.GetProductServiceInstance().PublishProduct(c.Request.Context(), id)
		if err != nil {
			log.Logger.Errorf("UpdateProductStatus: Failed to publish product: %v", err)
			c.JSON(http.StatusOK, data.ResponseFailed(err.Error()))
			return
		}
		c.JSON(http.StatusOK, data.ResponseSuccess("publish product success"))
	} else {
		err := service.GetProductServiceInstance().UnpublishProduct(c.Request.Context(), id)
		if err != nil {
			log.Logger.Errorf("UpdateProductStatus: Failed to unpublish product: %v", err)
			c.JSON(http.StatusOK, data.ResponseFailed(err.Error()))
			return
		}
		c.JSON(http.StatusOK, data.ResponseSuccess("unpublish product success"))
	}
}

// UpdateProductStock godoc
// @Summary 商家端更新商品库存
// @Description 只有当商品处于下架状态时，才能更改商品库存
// @Tags 商品
// @Accept json
// @Produce json
// @Param request body types.UpdateProductStockRequest true "更新商品库存请求"
// @Success 200 {object} data.BaseResponse "更新成功"
// @Failure 400 {object} data.BaseResponse "请求参数错误"
// @Router /merchant/products/:id/stock [patch]
func UpdateProductStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Logger.Errorf("GetProduct: Invalid product ID: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid product ID"))
		return
	}

	var req types.UpdateProductStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger.Errorf("UpdateProductStock: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed(err.Error()))
		return
	}

	err = service.GetProductServiceInstance().UpdateProductStock(c.Request.Context(), id, req.Stock)
	if err != nil {
		log.Logger.Errorf("UpdateProductStock: Failed to update product stock: %v", err)
		c.JSON(http.StatusOK, data.ResponseFailed(err.Error()))
		return
	}

	c.JSON(http.StatusOK, data.ResponseSuccess(nil))
}

// GetCustomerProductList godoc
// @Summary 用户端获取商品列表
// @Description 支持按关键词搜索、分类筛选、分页，并按更新时间排序
// @Tags 商品
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param category query string false "商品分类"
// @Param offset query int false "偏移量，默认0"
// @Param order_by query int false "排序方式：0-按更新时间降序，1-按更新时间升序，默认0"
// @Success 200 {object} data.BaseResponse
// @Failure 400 {object} data.BaseResponse
// @Failure 500 {object} data.BaseResponse
// @Router /customer/products [get]
func GetCustomerProductList(c *gin.Context) {
	var req types.GetProductListRequest

	// 获取查询参数
	req.Keyword = c.Query("keyword")
	req.Category = c.Query("category")
	offsetStr := c.Query("offset")
	orderByStr := c.DefaultQuery("order_by", "0")

	// 处理offset参数
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			log.Logger.Errorf("GetCustomerProductList: Invalid offset parameter: %v", err)
			c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid offset parameter"))
			return
		}
		req.Offset = offset
	}

	// 处理order_by参数
	orderBy, err := strconv.Atoi(orderByStr)
	if err != nil || (orderBy != 0 && orderBy != 1) {
		log.Logger.Errorf("GetCustomerProductList: Invalid order_by parameter: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid order_by parameter"))
		return
	}
	req.OrderBy = orderBy

	// 构造service层参数
	query := types.GetProductListQuery{
		Keyword:    req.Keyword,
		Limit:      10,
		Offset:     req.Offset,
		OrderBy:    req.OrderBy,
		Category:   req.Category,
		IsCustomer: true,
	}

	// 调用service层获取商品列表
	productList, total, err := service.GetProductServiceInstance().GetProductList(c.Request.Context(), query)
	if err != nil {
		log.Logger.Errorf("GetCustomerProductList: Failed to get product list: %v", err)
		c.JSON(http.StatusInternalServerError, data.ResponseFailed("Failed to get product list"))
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, data.ResponseSuccess(gin.H{
		"total": total,
		"list":  productList,
	}))
}

// GetMerchantProductList godoc
// @Summary 商家端获取商品列表
// @Description 支持按关键词搜索、分类筛选、分页，并按更新时间排序
// @Tags 商品
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param category query string false "商品分类"
// @Param offset query int false "偏移量，默认0"
// @Param order_by query int false "排序方式：0-按更新时间降序，1-按更新时间升序，默认0"
// @Success 200 {object} data.BaseResponse
// @Failure 400 {object} data.BaseResponse
// @Failure 500 {object} data.BaseResponse
// @Router /merchant/products [get]
func GetMerchantProductList(c *gin.Context) {
	var req types.GetProductListRequest

	// 获取查询参数
	req.Keyword = c.Query("keyword")
	req.Category = c.Query("category")
	offsetStr := c.Query("offset")
	orderByStr := c.DefaultQuery("order_by", "0")

	// 处理offset参数
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			log.Logger.Errorf("GetMerchantProductList: Invalid offset parameter: %v", err)
			c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid offset parameter"))
			return
		}
		req.Offset = offset
	}

	// 处理order_by参数
	orderBy, err := strconv.Atoi(orderByStr)
	if err != nil || (orderBy != 0 && orderBy != 1) {
		log.Logger.Errorf("GetMerchantProductList: Invalid order_by parameter: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid order_by parameter"))
		return
	}
	req.OrderBy = orderBy

	// 构造service层参数
	query := types.GetProductListQuery{
		Keyword:    req.Keyword,
		Limit:      10,
		Offset:     req.Offset,
		OrderBy:    req.OrderBy,
		Category:   req.Category,
		IsCustomer: false,
	}

	// 调用service层获取商品列表
	productList, total, err := service.GetProductServiceInstance().GetProductList(c.Request.Context(), query)
	if err != nil {
		log.Logger.Errorf("GetMerchantProductList: Failed to get product list: %v", err)
		c.JSON(http.StatusInternalServerError, data.ResponseFailed("Failed to get product list"))
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, data.ResponseSuccess(gin.H{
		"total": total,
		"list":  productList,
	}))
}

// EditProductInfo godoc
// @Summary 编辑商品信息
// @Description 根据商品ID更新商品详细信息
// @Tags 商品
// @Accept json
// @Produce json
// @Param request body types.UpdateProductInfoRequest true "编辑商品请求"
// @Success 200 {object} data.BaseResponse "编辑成功"
// @Failure 400 {object} data.BaseResponse "请求参数错误"
// @Failure 404 {object} data.BaseResponse "商品不存在"
// @Failure 500 {object} data.BaseResponse "服务器内部错误"
// @Router /merchant/products/{id} [put]
func EditProductInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Logger.Errorf("GetProduct: Invalid product ID: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid product ID"))
		return
	}

	var req types.UpdateProductInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger.Errorf("EditProductInfo: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed(err.Error()))
		return
	}

	req.ID = id

	// 调用 service 层更新商品信息
	err = service.GetProductServiceInstance().UpdateProductInfo(c.Request.Context(), &req)
	if err != nil {
		log.Logger.Errorf("EditProductInfo: Failed to update product info: %v", err)
		c.JSON(http.StatusInternalServerError, data.ResponseFailed("Failed to update product info"))
		return
	}

	c.JSON(http.StatusOK, data.ResponseSuccess(nil))
}

// GetProductCustomer godoc
// @Summary 获取商品详情(用户侧)
// @Description 根据商品ID获取商品详细信息
// @Tags 商品
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Success 200 {object} data.BaseResponse{data=types.ProductInfo} "成功"
// @Failure 400 {object} data.BaseResponse "请求参数错误"
// @Failure 404 {object} data.BaseResponse "商品不存在"
// @Failure 500 {object} data.BaseResponse "服务器内部错误"
// @Router /customer/product/{id} [get]
func GetProductCustomer(c *gin.Context) {
	// 解析路径参数
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Logger.Errorf("GetProduct: Invalid product ID: %v", err)
		c.JSON(http.StatusBadRequest, data.ResponseFailed("Invalid product ID"))
		return
	}

	// 调用 service 层获取商品信息
	product, err := service.GetProductServiceInstance().GetPublishedProductByID(c.Request.Context(), id)
	if err != nil {
		log.Logger.Errorf("GetProduct: Failed to get product details: %v", err)
		c.JSON(http.StatusInternalServerError, data.ResponseFailed("Failed to get product details"))
		return
	}

	// 如果没找到商品
	if product == nil {
		c.JSON(http.StatusNotFound, data.ResponseFailed("Product not found"))
		return
	}

	// 返回商品信息
	c.JSON(http.StatusOK, data.ResponseSuccess(product))
}
