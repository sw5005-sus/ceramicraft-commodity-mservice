package service

import (
	"context"
	"fmt"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/dao"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/model"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/types"
)

type ProductService interface {
	Create(ctx context.Context, product *types.ProductInfo) (productId int, err error)
	GetProductByID(ctx context.Context, id int) (productInfo *types.ProductInfo, err error)
	PublishProduct(ctx context.Context, id int) error
	UnpublishProduct(ctx context.Context, id int) error
	ReviewSubmit(ctx context.Context, id int) error
	ReviewReject(ctx context.Context, id int) error

	// 商家后台更新商品库存
	UpdateProductStock(ctx context.Context, id int, newStock int) error
	GetProductList(ctx context.Context, req types.GetProductListQuery) (list []*types.ProductInfo, count int, err error)

	UpdateStockWithCAS(ctx context.Context, id int, deta int) error
	UpdateProductInfo(ctx context.Context, req *types.UpdateProductInfoRequest) error
}

type ProductServiceImpl struct {
	productDao dao.ProductDao
}

func GetProductServiceInstance() *ProductServiceImpl {
	return &ProductServiceImpl{
		productDao: dao.GetProductDao(),
	}
}

func (p *ProductServiceImpl) Create(ctx context.Context, product *types.ProductInfo) (productId int, err error) {
	productModel := product.ToProductModel()
	productModel.LatestEditorId = getUserId(ctx)
	id, err := p.productDao.CreateProduct(ctx, productModel)
	if err != nil {
		log.Logger.Errorf("ProductService: Failed to create product: %v", err)
		return -1, err
	}
	return id, nil
}

// GetProductByID 根据ID获取产品信息 (商家侧，无论是否上架都可以看到)
func (p *ProductServiceImpl) GetProductByID(ctx context.Context, id int) (productInfo *types.ProductInfo, err error) {
	product, err := p.productDao.GetProductByID(ctx, id)
	if err != nil {
		log.Logger.Errorf("ProductService: Failed to get product by ID: %v", err)
		return nil, err
	}
	if product == nil {
		return nil, nil
	}
	return types.NewProductInfo(product), nil
}

const (
	ProductStatusUnpublished = 0 // 下架状态
	ProductStatusPublished   = 1 // 上架状态
	ProductStatusUnderReview = 2 // 审核中状态
)

// GetProductByID 根据ID获取产品信息 (用户侧， 只有上架的商品才能查看详情页)
func (p *ProductServiceImpl) GetPublishedProductByID(ctx context.Context, id int) (productInfo *types.ProductInfo, err error) {
	product, err := p.productDao.GetProductByID(ctx, id)
	if err != nil {
		log.Logger.Errorf("ProductService: Failed to get product by ID: %v", err)
		return nil, err
	}
	if product == nil || product.Status == ProductStatusUnpublished {
		return nil, nil
	}
	return types.NewProductInfo(product), nil
}

// ReviewSubmit 提交审核
func (p *ProductServiceImpl) ReviewSubmit(ctx context.Context, id int) error {
	// 获取商品当前信息
	product, err := p.productDao.GetProductByID(ctx, id)
	if err != nil {
		log.Logger.Errorf("PublishProduct: Failed to get product by ID: %v", err)
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found with ID: %d", id)
	}

	if product.Status != ProductStatusUnpublished {
		return fmt.Errorf("only unpublished products can be submitted for review, product ID: %d", id)
	}
	oldStatus := int(product.Status)
	product.Status = ProductStatusUnderReview
	err = p.productDao.UpdateProductStatus(ctx, id, oldStatus, product)
	if err != nil {
		log.Logger.Errorf("ReviewSubmit: Failed to update product status: %v", err)
		return err
	}
	log.Logger.Infof("Product (ID: %d) submitted for review successfully", id)
	return nil
}

// PublishProduct 上架商品
func (p *ProductServiceImpl) PublishProduct(ctx context.Context, id int) error {
	// 获取商品当前信息
	product, err := p.productDao.GetProductByID(ctx, id)
	if err != nil {
		log.Logger.Errorf("PublishProduct: Failed to get product by ID: %v", err)
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found with ID: %d", id)
	}

	// 检查当前状态
	if product.Status != ProductStatusUnderReview {
		return fmt.Errorf("product (ID: %d) must be in review for publish", id)
	}
	userId := getUserId(ctx)
	// 审核人跟编辑者不能是同一人
	if product.LatestEditorId == userId {
		return fmt.Errorf("the reviewer cannot be the same as the latest editor for product (ID: %d)", id)
	}

	product.LatestReviewerId = userId
	oldStatus := int(product.Status)
	product.Status = ProductStatusPublished
	// 更新状态为已上架
	err = p.productDao.UpdateProductStatus(ctx, id, oldStatus, product)
	if err != nil {
		log.Logger.Errorf("PublishProduct: Failed to update product status: %v", err)
		return err
	}

	return nil
}

// RejectReview 驳回审核
func (p *ProductServiceImpl) ReviewReject(ctx context.Context, id int) error {
	product, err := p.productDao.GetProductByID(ctx, id)
	if err != nil {
		log.Logger.Errorf("UnpublishProduct: Failed to get product by ID: %v", err)
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found with ID: %d", id)
	}
	if product.Status != ProductStatusUnderReview {
		return fmt.Errorf("only products under review can be rejected, product ID: %d", id)
	}
	userId := getUserId(ctx)
	if product.LatestEditorId == userId {
		return fmt.Errorf("the reviewer cannot be the same as the latest editor for product (ID: %d)", id)
	}
	product.LatestReviewerId = userId
	oldStatus := int(product.Status)
	product.Status = ProductStatusUnpublished
	err = p.productDao.UpdateProductStatus(ctx, id, oldStatus, product)
	if err != nil {
		log.Logger.Errorf("RejectReview: Failed to update product status: %v", err)
		return err
	}
	log.Logger.Infof("Product (ID: %d) review rejected successfully", id)
	return nil
}

// UnpublishProduct 商品从上架状态变更为下架状态
func (p *ProductServiceImpl) UnpublishProduct(ctx context.Context, id int) error {
	// 获取商品当前信息
	product, err := p.productDao.GetProductByID(ctx, id)
	if err != nil {
		log.Logger.Errorf("UnpublishProduct: Failed to get product by ID: %v", err)
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found with ID: %d", id)
	}

	// 检查当前状态
	if product.Status == ProductStatusUnpublished {
		return fmt.Errorf("product (ID: %d) is already unpublished", id)
	}

	oldStatus := int(product.Status)
	product.Status = ProductStatusUnpublished
	// 更新状态为已下架
	err = p.productDao.UpdateProductStatus(ctx, id, oldStatus, product)
	if err != nil {
		log.Logger.Errorf("UnpublishProduct: Failed to update product status: %v", err)
		return err
	}

	return nil
}

// UpdateProductStock 更新商品库存
// 要求：
// 1. 商品必须存在
// 2. 商品必须处于下架状态
// 3. 新的库存不能小于0
func (p *ProductServiceImpl) UpdateProductStock(ctx context.Context, id int, newStock int) error {
	// 检查库存是否合法
	if newStock < 0 {
		return fmt.Errorf("invalid stock value: %d, stock cannot be negative", newStock)
	}

	// 获取商品信息
	product, err := p.productDao.GetProductByID(ctx, id)
	if err != nil {
		log.Logger.Errorf("UpdateProductStock: Failed to get product by ID: %v", err)
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found with ID: %d", id)
	}

	// 检查商品状态
	if product.Status != ProductStatusUnpublished {
		return fmt.Errorf("cannot update stock for published product (ID: %d)", id)
	}

	// 更新库存
	err = p.productDao.UpdateProductStock(ctx, id, newStock, getUserId(ctx))
	if err != nil {
		log.Logger.Errorf("UpdateProductStock: Failed to update stock: %v", err)
		return err
	}

	return nil
}

func (p *ProductServiceImpl) GetProductList(ctx context.Context, req types.GetProductListQuery) (list []*types.ProductSimplifiedInfo, count int, err error) {
	listRaw, cnt, err := p.productDao.ListProduct(ctx, dao.ListProductQuery{
		Keyword:    req.Keyword,
		Category:   req.Category,
		Offset:     req.Offset,
		Limit:      req.Limit,
		IsCustomer: req.IsCustomer,
		OrderBy:    req.OrderBy,
	})
	if err != nil {
		log.Logger.Errorf("GetProductList: Failed to get product list, err: %v", err)
		return nil, -1, err
	}

	list = make([]*types.ProductSimplifiedInfo, len(listRaw))
	for k, listModel := range listRaw {
		list[k] = &types.ProductSimplifiedInfo{
			ID:       int(listModel.ID),
			Name:     listModel.Name,
			Category: listModel.Category,
			Price:    listModel.Price,
			Desc:     listModel.Desc,
			Stock:    listModel.Stock,
			PicInfo:  listModel.PicInfo,
			Status:   listModel.Status,
		}
	}

	return list, cnt, nil
}

func (p *ProductServiceImpl) UpdateStockWithCAS(ctx context.Context, id, deta int) error {
	pModel, err := p.productDao.GetProductByID(ctx, id)
	if err != nil {
		log.Logger.Errorf("UpdateStockWithCAS: get product failed, err: %s", err.Error())
		return err
	}

	if int(pModel.Stock)+deta < 0 {
		log.Logger.Errorf("UpdateStockWithCAS: do not have enough stock, product id: %d, current stock: %d", id, int(pModel.Stock))
		return fmt.Errorf("do not have enough stock, product id: %d, current stock: %d", id, int(pModel.Stock))
	}

	newStock := int(pModel.Stock) + deta
	err = p.productDao.UpdateStockWithCAS(ctx, id, int(pModel.Version), newStock, getUserId(ctx))
	if err != nil {
		log.Logger.Errorf("UpdateStockWithCAS: update failed, err:%s", err.Error())
		return err
	}

	return nil
}

// UpdateProductInfo 更新商品信息
// 要求：
// 1. 商品必须存在
// 2. 商品必须处于下架状态
func (p *ProductServiceImpl) UpdateProductInfo(ctx context.Context, req *types.UpdateProductInfoRequest) error {
	// 获取商品信息
	product, err := p.productDao.GetProductByID(ctx, req.ID)
	if err != nil {
		log.Logger.Errorf("UpdateProductInfo: Failed to get product by ID: %v", err)
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found with ID: %d", req.ID)
	}

	// 检查商品状态
	if product.Status != ProductStatusUnpublished {
		return fmt.Errorf("cannot update product info for published product (ID: %d)", req.ID)
	}

	// 构建更新的商品模型
	updatedProduct := &model.Product{
		Model:            product.Model, // 保持原有的ID、创建时间等
		Name:             req.Name,
		Category:         req.Category,
		Price:            req.Price,
		Desc:             req.Desc,
		Stock:            product.Stock, // 保持原有库存
		PicInfo:          req.PicInfo,
		Dimensions:       req.Dimensions,
		Material:         req.Material,
		Weight:           req.Weight,
		Capacity:         req.Capacity,
		CareInstructions: req.CareInstructions,
		Status:           product.Status,  // 保持原有状态
		Version:          product.Version, // 保持原有版本
		LatestEditorId:   getUserId(ctx),  // 更新编辑者为当前用户
	}

	// 调用DAO层更新商品信息
	err = p.productDao.UpdateProduct(ctx, updatedProduct)
	if err != nil {
		log.Logger.Errorf("UpdateProductInfo: Failed to update product: %v", err)
		return err
	}

	return nil
}

func getUserId(ctx context.Context) int {
	userId, ok := ctx.Value(types.UserIDKey).(int)
	if !ok {
		log.Logger.Warn("getUserId: user_id not found in context or invalid type")
		return 0
	}
	return userId
}
