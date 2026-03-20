package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/dao"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/dao/mocks"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/model"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mockgen -source=dao/productDao.go -destination=dao/mocks/productDao_mock.go -package=mocks

func TestProductServiceImpl_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()

	m := mocks.NewMockProductDao(ctrl)
	expectEditorId := 1
	productModel := &model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           0,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
		LatestEditorId:   expectEditorId,
	}

	m.EXPECT().CreateProduct(ctx, gomock.Eq(productModel)).Return(1, nil)

	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	productInfo := &types.ProductInfo{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           0,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}

	_, err := testProductServiceImpl.Create(ctx, productInfo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func init() {
	// 初始化测试用logger
	logger, _ := zap.NewDevelopment()
	log.Logger = logger.Sugar()
}

func TestProductServiceImpl_GetProductByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)

	m.EXPECT().GetProductByID(ctx, 1).Return(&model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           0,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}, nil)

	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	productInfo, err := testProductServiceImpl.GetProductByID(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedProductInfo := &types.ProductInfo{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           0,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}

	if !reflect.DeepEqual(productInfo, expectedProductInfo) {
		t.Errorf("Product info mismatch:\ngot: %+v\nwant: %+v", productInfo, expectedProductInfo)
	}

	m.EXPECT().GetProductByID(ctx, 2).Return(nil, errors.New("product not found"))

	_, err = testProductServiceImpl.GetProductByID(ctx, 2)
	if err == nil {
		t.Errorf("Expected error when getting a non-existent product, got nil")
	}

	m.EXPECT().GetProductByID(ctx, 3).Return(nil, nil)

	productInfo, _ = testProductServiceImpl.GetProductByID(ctx, 3)
	if productInfo != nil {
		t.Errorf("Expected nil product info for non-existent product, got %+v", productInfo)
	}
}

func TestProductServiceImpl_PublishProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)
	p := &model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           ProductStatusUnderReview,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}
	m.EXPECT().GetProductByID(ctx, 1).Return(p, nil)

	m.EXPECT().UpdateProductStatus(ctx, 1, ProductStatusUnderReview, p).Return(nil)

	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	err := testProductServiceImpl.PublishProduct(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	p = &model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           ProductStatusPublished,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}
	m.EXPECT().GetProductByID(ctx, 2).Return(p, nil)

	err = testProductServiceImpl.PublishProduct(ctx, 2)
	if err == nil {
		t.Errorf("Expected error when publishing an already published product, got nil")
	}

	m.EXPECT().GetProductByID(ctx, 3).Return(nil, errors.New("product not found"))

	err = testProductServiceImpl.PublishProduct(ctx, 3)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	m.EXPECT().GetProductByID(ctx, 4).Return(nil, nil)
	err = testProductServiceImpl.PublishProduct(ctx, 4)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	p = &model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           ProductStatusUnderReview,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}
	m.EXPECT().GetProductByID(ctx, 5).Return(p, nil)

	m.EXPECT().UpdateProductStatus(ctx, 5, ProductStatusUnderReview, p).Return(errors.New("database error"))
	err = testProductServiceImpl.PublishProduct(ctx, 5)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// editor and reviwer should not be the same person
	p.LatestEditorId = 1
	m.EXPECT().GetProductByID(ctx, 6).Return(p, nil)
	err = testProductServiceImpl.PublishProduct(ctx, 6)
	if err == nil {
		t.Errorf("Expected error when editor and reviewer are the same, got nil")
	}
}

func TestProductServiceImpl_UnpublishProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)
	p := &model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           ProductStatusPublished,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}
	m.EXPECT().GetProductByID(ctx, 1).Return(p, nil)
	m.EXPECT().UpdateProductStatus(ctx, 1, ProductStatusPublished, p).Return(nil)

	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	err := testProductServiceImpl.UnpublishProduct(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	m.EXPECT().GetProductByID(ctx, 2).Return(&model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           0,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}, nil)

	err = testProductServiceImpl.UnpublishProduct(ctx, 2)
	if err == nil {
		t.Errorf("Expected error when unpublishing an already unpublished product, got nil")
	}

	m.EXPECT().GetProductByID(ctx, 3).Return(nil, errors.New("product not found"))

	err = testProductServiceImpl.UnpublishProduct(ctx, 3)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	m.EXPECT().GetProductByID(ctx, 4).Return(nil, nil)
	err = testProductServiceImpl.UnpublishProduct(ctx, 4)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	p = &model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           ProductStatusPublished,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}
	m.EXPECT().GetProductByID(ctx, 5).Return(p, nil)

	m.EXPECT().UpdateProductStatus(ctx, 5, ProductStatusPublished, p).Return(errors.New("database error"))

	err = testProductServiceImpl.UnpublishProduct(ctx, 5)
	if err == nil {
		t.Errorf("Expected database error, got nil")
	}
}

func TestProductServiceImpl_GetPublishedProductByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)
	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	// 测试获取已上架商品
	publishedProduct := &model.Product{
		Name:             "已上架商品",
		Category:         "茶具",
		Price:            10000,
		Desc:             "精美陶瓷茶具",
		Stock:            100,
		Status:           1, // 已上架
		PicInfo:          "pic1.jpg",
		Dimensions:       "10x10x10",
		Material:         "陶瓷",
		Weight:           "1kg",
		Capacity:         "500ml",
		CareInstructions: "小心轻放",
	}
	m.EXPECT().GetProductByID(ctx, 1).Return(publishedProduct, nil)

	productInfo, err := testProductServiceImpl.GetPublishedProductByID(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if productInfo == nil {
		t.Error("Expected product info, got nil")
	} else {
		// 验证返回的商品信息是否正确
		if productInfo.Name != publishedProduct.Name {
			t.Errorf("Expected name %s, got %s", publishedProduct.Name, productInfo.Name)
		}
		if productInfo.Status != publishedProduct.Status {
			t.Errorf("Expected status %d, got %d", publishedProduct.Status, productInfo.Status)
		}
	}

	// 测试获取未上架商品（应返回nil）
	unpublishedProduct := &model.Product{
		Name:   "未上架商品",
		Status: 0, // 未上架
	}
	m.EXPECT().GetProductByID(ctx, 2).Return(unpublishedProduct, nil)

	productInfo, err = testProductServiceImpl.GetPublishedProductByID(ctx, 2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if productInfo != nil {
		t.Error("Expected nil for unpublished product, got product info")
	}

	// 测试获取不存在的商品
	m.EXPECT().GetProductByID(ctx, 3).Return(nil, nil)

	productInfo, err = testProductServiceImpl.GetPublishedProductByID(ctx, 3)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if productInfo != nil {
		t.Error("Expected nil for non-existent product, got product info")
	}

	// 测试数据库错误的情况
	m.EXPECT().GetProductByID(ctx, 4).Return(nil, errors.New("database error"))

	productInfo, err = testProductServiceImpl.GetPublishedProductByID(ctx, 4)
	if err == nil {
		t.Error("Expected database error, got nil")
	}
	if productInfo != nil {
		t.Error("Expected nil product info when database error occurs")
	}
}

func TestProductServiceImpl_GetProductList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockProductDao(ctrl)

	// 准备测试数据
	mockProducts := []*model.Product{
		{
			Name:             "陶瓷茶具1",
			Category:         "茶具",
			Price:            10000,
			Desc:             "精美陶瓷茶具",
			Stock:            100,
			Status:           1, // 已上架
			PicInfo:          "pic1.jpg",
			Dimensions:       "10x10x10",
			Material:         "陶瓷",
			Weight:           "1kg",
			Capacity:         "500ml",
			CareInstructions: "小心轻放",
		},
		{
			Name:             "陶瓷花瓶",
			Category:         "装饰品",
			Price:            20000,
			Desc:             "中式陶瓷花瓶",
			Stock:            50,
			Status:           0, // 未上架
			PicInfo:          "pic2.jpg",
			Dimensions:       "20x20x30",
			Material:         "陶瓷",
			Weight:           "2kg",
			Capacity:         "2L",
			CareInstructions: "防摔",
		},
	}

	testCases := []struct {
		name        string
		query       types.GetProductListQuery
		mockResult  []*model.Product
		mockCount   int
		mockError   error
		expectCount int
		expectLen   int
		expectError bool
	}{
		{
			name: "成功获取商家端全部商品列表",
			query: types.GetProductListQuery{
				Offset:     0,
				Limit:      10,
				IsCustomer: false,
				OrderBy:    0,
			},
			mockResult:  mockProducts,
			mockCount:   2,
			mockError:   nil,
			expectCount: 2,
			expectLen:   2,
			expectError: false,
		},
		{
			name: "成功获取用户端商品列表(只显示已上架)",
			query: types.GetProductListQuery{
				Offset:     0,
				Limit:      10,
				IsCustomer: true,
				OrderBy:    0,
			},
			mockResult:  mockProducts[:1], // 只返回已上架的商品
			mockCount:   1,
			mockError:   nil,
			expectCount: 1,
			expectLen:   1,
			expectError: false,
		},
		{
			name: "按关键词搜索",
			query: types.GetProductListQuery{
				Keyword:    "茶具",
				Offset:     0,
				Limit:      10,
				IsCustomer: true,
				OrderBy:    0,
			},
			mockResult:  mockProducts[:1],
			mockCount:   1,
			mockError:   nil,
			expectCount: 1,
			expectLen:   1,
			expectError: false,
		},
		{
			name: "按分类筛选",
			query: types.GetProductListQuery{
				Category:   "茶具",
				Offset:     0,
				Limit:      10,
				IsCustomer: true,
				OrderBy:    0,
			},
			mockResult:  mockProducts[:1],
			mockCount:   1,
			mockError:   nil,
			expectCount: 1,
			expectLen:   1,
			expectError: false,
		},
		{
			name: "数据库错误",
			query: types.GetProductListQuery{
				Offset:     0,
				Limit:      10,
				IsCustomer: true,
				OrderBy:    0,
			},
			mockResult:  nil,
			mockCount:   0,
			mockError:   errors.New("database error"),
			expectCount: -1,
			expectLen:   0,
			expectError: true,
		},
		{
			name: "空结果",
			query: types.GetProductListQuery{
				Keyword:    "不存在的商品",
				Offset:     0,
				Limit:      10,
				IsCustomer: true,
				OrderBy:    0,
			},
			mockResult:  []*model.Product{},
			mockCount:   0,
			mockError:   nil,
			expectCount: 0,
			expectLen:   0,
			expectError: false,
		},
		{
			name: "按更新时间升序",
			query: types.GetProductListQuery{
				Offset:     0,
				Limit:      10,
				IsCustomer: true,
				OrderBy:    1, // 升序
			},
			mockResult:  mockProducts[:1],
			mockCount:   1,
			mockError:   nil,
			expectCount: 1,
			expectLen:   1,
			expectError: false,
		},
	}

	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := createContext()
			// 设置Mock期望
			m.EXPECT().ListProduct(gomock.Any(), dao.ListProductQuery{
				Keyword:    tc.query.Keyword,
				Category:   tc.query.Category,
				Offset:     tc.query.Offset,
				Limit:      tc.query.Limit,
				IsCustomer: tc.query.IsCustomer,
				OrderBy:    tc.query.OrderBy,
			}).Return(tc.mockResult, tc.mockCount, tc.mockError)

			// 调用被测试的方法
			products, count, err := testProductServiceImpl.GetProductList(ctx, tc.query)

			// 验证结果
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if count != tc.expectCount {
				t.Errorf("Expected count %d but got %d", tc.expectCount, count)
			}

			if len(products) != tc.expectLen {
				t.Errorf("Expected %d products but got %d", tc.expectLen, len(products))
			}

			// 如果有返回结果，验证字段映射是否正确
			if len(products) > 0 {
				for i, p := range products {
					if p.Name != tc.mockResult[i].Name {
						t.Errorf("Product name mismatch at index %d: expected %s but got %s", i, tc.mockResult[i].Name, p.Name)
					}
					if p.Category != tc.mockResult[i].Category {
						t.Errorf("Product category mismatch at index %d: expected %s but got %s", i, tc.mockResult[i].Category, p.Category)
					}
					if p.Price != tc.mockResult[i].Price {
						t.Errorf("Product price mismatch at index %d: expected %d but got %d", i, tc.mockResult[i].Price, p.Price)
					}
					if p.Stock != tc.mockResult[i].Stock {
						t.Errorf("Product stock mismatch at index %d: expected %d but got %d", i, tc.mockResult[i].Stock, p.Stock)
					}
					if p.Status != tc.mockResult[i].Status {
						t.Errorf("Product status mismatch at index %d: expected %d but got %d", i, tc.mockResult[i].Status, p.Status)
					}
				}
			}
		})
	}
}

func TestProductServiceImpl_UpdateStockWithCAS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)
	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	// 测试成功增加库存
	m.EXPECT().GetProductByID(ctx, 1).Return(&model.Product{
		Model: gorm.Model{
			ID: 1,
		},
		Name:    "Test Product",
		Stock:   50,
		Version: 1,
	}, nil)
	m.EXPECT().UpdateStockWithCAS(ctx, 1, 1, 60, 1).Return(nil)

	err := testProductServiceImpl.UpdateStockWithCAS(ctx, 1, 10)
	if err != nil {
		t.Errorf("Expected no error when increasing stock, got %v", err)
	}

	// 测试成功减少库存
	m.EXPECT().GetProductByID(ctx, 2).Return(&model.Product{
		Model: gorm.Model{
			ID: 2,
		},
		Name:    "Test Product",
		Stock:   50,
		Version: 1,
	}, nil)
	m.EXPECT().UpdateStockWithCAS(ctx, 2, 1, 40, 1).Return(nil)

	err = testProductServiceImpl.UpdateStockWithCAS(ctx, 2, -10)
	if err != nil {
		t.Errorf("Expected no error when decreasing stock, got %v", err)
	}

	// 测试库存不足的情况
	m.EXPECT().GetProductByID(ctx, 3).Return(&model.Product{
		Model: gorm.Model{
			ID: 3,
		},
		Name:    "Test Product",
		Stock:   5,
		Version: 1,
	}, nil)

	err = testProductServiceImpl.UpdateStockWithCAS(ctx, 3, -10)
	if err == nil {
		t.Error("Expected error when stock is insufficient, got nil")
	}

	// 测试获取商品信息失败
	m.EXPECT().GetProductByID(ctx, 4).Return(nil, fmt.Errorf("database error"))

	err = testProductServiceImpl.UpdateStockWithCAS(ctx, 4, 10)
	if err == nil {
		t.Error("Expected error when getting product fails, got nil")
	}

	// 测试CAS更新失败（版本号不匹配）
	m.EXPECT().GetProductByID(ctx, 5).Return(&model.Product{
		Model: gorm.Model{
			ID: 5,
		},
		Name:    "Test Product",
		Stock:   50,
		Version: 1,
	}, nil)
	m.EXPECT().UpdateStockWithCAS(ctx, 5, 1, 60, 1).Return(fmt.Errorf("version conflict"))

	err = testProductServiceImpl.UpdateStockWithCAS(ctx, 5, 10)
	if err == nil {
		t.Error("Expected error when CAS update fails, got nil")
	}
}

func TestProductServiceImpl_UpdateProductInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)
	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	// 测试成功更新商品信息
	existingProduct := &model.Product{
		Model: gorm.Model{
			ID: 1,
		},
		Name:             "Old Product Name",
		Category:         "Old Category",
		Price:            100,
		Desc:             "Old Description",
		Stock:            50,
		Status:           0, // 未上架
		PicInfo:          "old_pic.jpg",
		Dimensions:       "10x10x10",
		Material:         "Old Material",
		Weight:           "1kg",
		Capacity:         "500ml",
		CareInstructions: "Old instructions",
		Version:          1,
	}

	updateRequest := &types.UpdateProductInfoRequest{
		ID:               1,
		Name:             "Updated Product Name",
		Category:         "Updated Category",
		Price:            200,
		Desc:             "Updated Description",
		PicInfo:          "updated_pic.jpg",
		Dimensions:       "20x20x20",
		Material:         "Updated Material",
		Weight:           "2kg",
		Capacity:         "1L",
		CareInstructions: "Updated instructions",
	}

	expectedUpdatedProduct := &model.Product{
		Model:            existingProduct.Model,
		Name:             updateRequest.Name,
		Category:         updateRequest.Category,
		Price:            updateRequest.Price,
		Desc:             updateRequest.Desc,
		Stock:            existingProduct.Stock, // 保持原有库存
		PicInfo:          updateRequest.PicInfo,
		Dimensions:       updateRequest.Dimensions,
		Material:         updateRequest.Material,
		Weight:           updateRequest.Weight,
		Capacity:         updateRequest.Capacity,
		CareInstructions: updateRequest.CareInstructions,
		Status:           existingProduct.Status,  // 保持原有状态
		Version:          existingProduct.Version, // 保持原有版本
		LatestEditorId:   1,                       // 假设编辑者ID为1
	}

	m.EXPECT().GetProductByID(ctx, 1).Return(existingProduct, nil)
	m.EXPECT().UpdateProduct(ctx, expectedUpdatedProduct).Return(nil)

	err := testProductServiceImpl.UpdateProductInfo(ctx, updateRequest)
	if err != nil {
		t.Errorf("Expected no error when updating product info, got %v", err)
	}

	// 测试商品不存在的情况
	m.EXPECT().GetProductByID(ctx, 2).Return(nil, errors.New("product not found"))

	updateRequest2 := &types.UpdateProductInfoRequest{
		ID:   2,
		Name: "Test Product",
	}

	err = testProductServiceImpl.UpdateProductInfo(ctx, updateRequest2)
	if err == nil {
		t.Error("Expected error when product not found, got nil")
	}

	// 测试商品为nil的情况
	m.EXPECT().GetProductByID(ctx, 3).Return(nil, nil)

	updateRequest3 := &types.UpdateProductInfoRequest{
		ID:   3,
		Name: "Test Product",
	}

	err = testProductServiceImpl.UpdateProductInfo(ctx, updateRequest3)
	if err == nil {
		t.Error("Expected error when product is nil, got nil")
	}

	// 测试尝试更新已上架商品的情况
	publishedProduct := &model.Product{
		Model: gorm.Model{
			ID: 4,
		},
		Name:   "Published Product",
		Status: 1, // 已上架
	}

	m.EXPECT().GetProductByID(ctx, 4).Return(publishedProduct, nil)

	updateRequest4 := &types.UpdateProductInfoRequest{
		ID:   4,
		Name: "Updated Name",
	}

	err = testProductServiceImpl.UpdateProductInfo(ctx, updateRequest4)
	if err == nil {
		t.Error("Expected error when updating published product, got nil")
	}

	// 测试DAO更新失败的情况
	unpublishedProduct := &model.Product{
		Model: gorm.Model{
			ID: 5,
		},
		Name:    "Unpublished Product",
		Status:  0, // 未上架
		Stock:   30,
		Version: 2,
	}

	updateRequest5 := &types.UpdateProductInfoRequest{
		ID:   5,
		Name: "Updated Name",
	}

	expectedUpdatedProduct5 := &model.Product{
		Model:            unpublishedProduct.Model,
		Name:             updateRequest5.Name,
		Category:         updateRequest5.Category,
		Price:            updateRequest5.Price,
		Desc:             updateRequest5.Desc,
		Stock:            unpublishedProduct.Stock,
		PicInfo:          updateRequest5.PicInfo,
		Dimensions:       updateRequest5.Dimensions,
		Material:         updateRequest5.Material,
		Weight:           updateRequest5.Weight,
		Capacity:         updateRequest5.Capacity,
		CareInstructions: updateRequest5.CareInstructions,
		Status:           unpublishedProduct.Status,
		Version:          unpublishedProduct.Version,
		LatestEditorId:   1,
	}

	m.EXPECT().GetProductByID(ctx, 5).Return(unpublishedProduct, nil)
	m.EXPECT().UpdateProduct(ctx, expectedUpdatedProduct5).Return(errors.New("database error"))

	err = testProductServiceImpl.UpdateProductInfo(ctx, updateRequest5)
	if err == nil {
		t.Error("Expected error when DAO update fails, got nil")
	}
}

func TestProductServiceImpl_UpdateProductStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)
	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	// 测试成功更新库存的情况
	m.EXPECT().GetProductByID(ctx, 1).Return(&model.Product{
		Name:   "Test Product",
		Stock:  50,
		Status: 0,
	}, nil)
	m.EXPECT().UpdateProductStock(ctx, 1, 60, 1).Return(nil)

	err := testProductServiceImpl.UpdateProductStock(ctx, 1, 60)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 测试产品不存在的情况
	m.EXPECT().GetProductByID(ctx, 2).Return(nil, errors.New("product not found"))

	err = testProductServiceImpl.UpdateProductStock(ctx, 2, 30)
	if err == nil {
		t.Errorf("Expected error for non-existent product, got nil")
	}

	// 测试产品为nil的情况
	m.EXPECT().GetProductByID(ctx, 3).Return(nil, nil)

	err = testProductServiceImpl.UpdateProductStock(ctx, 3, 30)
	if err == nil {
		t.Errorf("Expected error for nil product, got nil")
	}

	// 测试更新库存失败的情况
	m.EXPECT().GetProductByID(ctx, 4).Return(&model.Product{
		Name:   "Test Product",
		Stock:  50,
		Status: 0,
	}, nil)
	m.EXPECT().UpdateProductStock(ctx, 4, 70, 1).Return(errors.New("database error"))

	err = testProductServiceImpl.UpdateProductStock(ctx, 4, 70)
	if err == nil {
		t.Errorf("Expected database error, got nil")
	}

	// 测试更新负数库存的情况
	err = testProductServiceImpl.UpdateProductStock(ctx, 5, -10)
	if err == nil {
		t.Errorf("Expected error for negative stock, got nil")
	}

	// 测试更新库存为零的情况
	m.EXPECT().GetProductByID(ctx, 6).Return(&model.Product{
		Name:   "Test Product",
		Stock:  50,
		Status: 0,
	}, nil)
	m.EXPECT().UpdateProductStock(ctx, 6, 0, 1).Return(nil)

	err = testProductServiceImpl.UpdateProductStock(ctx, 6, 0)
	if err != nil {
		t.Errorf("Expected no error for zero stock, got %v", err)
	}

	// 测试更新已上架商品库存的情况
	m.EXPECT().GetProductByID(ctx, 7).Return(&model.Product{
		Name:   "Test Product",
		Stock:  50,
		Status: 1,
	}, nil)

	err = testProductServiceImpl.UpdateProductStock(ctx, 7, 60)
	if err == nil {
		t.Errorf("Expected error when updating stock for published product, got nil")
	}
}
func TestProductServiceImpl_ReviewReject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)

	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	product := &model.Product{
		Name:             "Test Product",
		Price:            200,
		Desc:             "This is a test product",
		Stock:            50,
		PicInfo:          "http://example.com/pic.jpg",
		Status:           ProductStatusUnderReview,
		Category:         "Test Category",
		Weight:           "1kg",
		Material:         "Plastic",
		Capacity:         "500ml",
		Dimensions:       "10x10x10cm",
		CareInstructions: "Handle with care",
	}

	m.EXPECT().GetProductByID(ctx, 1).Return(product, nil)
	m.EXPECT().UpdateProductStatus(ctx, 1, ProductStatusUnderReview, product).Return(nil)

	err := testProductServiceImpl.ReviewReject(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	m.EXPECT().GetProductByID(ctx, 2).Return(&model.Product{
		Name:   "Test Product",
		Status: ProductStatusUnpublished,
	}, nil)

	err = testProductServiceImpl.ReviewReject(ctx, 2)
	if err == nil {
		t.Errorf("Expected error when rejecting an already unpublished product, got nil")
	}

	m.EXPECT().GetProductByID(ctx, 3).Return(nil, errors.New("product not found"))

	err = testProductServiceImpl.ReviewReject(ctx, 3)
	if err == nil {
		t.Errorf("Expected error when product not found, got nil")
	}

	m.EXPECT().GetProductByID(ctx, 4).Return(nil, nil)

	err = testProductServiceImpl.ReviewReject(ctx, 4)
	if err == nil {
		t.Errorf("Expected error when product does not exist, got nil")
	}

	// editor and reviwer should not be the same person
	product.LatestEditorId = 1
	m.EXPECT().GetProductByID(ctx, 5).Return(product, nil)
	err = testProductServiceImpl.ReviewReject(ctx, 5)
	if err == nil {
		t.Errorf("Expected error when editor and reviewer are the same, got nil")
	}
}

func TestProductServiceImpl_ReviewSubmit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := createContext()
	m := mocks.NewMockProductDao(ctrl)

	testProductServiceImpl := &ProductServiceImpl{
		productDao: m,
	}

	product := &model.Product{
		Status:           ProductStatusUnpublished,
		LatestEditorId:   1,
		LatestReviewerId: 0,
	}

	m.EXPECT().GetProductByID(ctx, 1).Return(product, nil)
	m.EXPECT().UpdateProductStatus(ctx, 1, ProductStatusUnpublished, product).Return(nil)

	err := testProductServiceImpl.ReviewSubmit(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test case: product already under review
	product.Status = ProductStatusUnderReview
	m.EXPECT().GetProductByID(ctx, 2).Return(product, nil)

	err = testProductServiceImpl.ReviewSubmit(ctx, 2)
	if err == nil {
		t.Errorf("Expected error when submitting review for product already under review, got nil")
	}

	// Test case: product not found
	product.Status = ProductStatusUnpublished
	m.EXPECT().GetProductByID(ctx, 3).Return(nil, nil)
	err = testProductServiceImpl.ReviewSubmit(ctx, 3)
	if err == nil {
		t.Errorf("Expected error when product not found, got nil")
	}

	// Test case: database error
	m.EXPECT().GetProductByID(ctx, 4).Return(product, nil)
	m.EXPECT().UpdateProductStatus(ctx, 4, ProductStatusUnpublished, product).Return(errors.New("database error"))

	err = testProductServiceImpl.ReviewSubmit(ctx, 4)
	if err == nil {
		t.Errorf("Expected database error, got nil")
	}
}

func createContext() context.Context {
	parent := context.Background()
	return context.WithValue(parent, types.UserIDKey, 1)
}
