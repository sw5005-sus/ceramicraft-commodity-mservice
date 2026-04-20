package dao

import (
	"context"
	"fmt"
	"sync"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/model"

	"gorm.io/gorm"
)

// This file defines the ProductDao interface and its GORM-based implementation.
// The following go:generate directive can be used to generate a gomock-based mock
// implementation of ProductDao for use in unit tests:
//
//go:generate mockgen -destination=./mocks/productDao_mock.go -package=mocks . ProductDao
type ProductDao interface {
	CreateProduct(ctx context.Context, product *model.Product) (productId int, err error)
	UpdateProduct(ctx context.Context, product *model.Product) error
	UpdateStockWithCAS(ctx context.Context, id int, version int, newStock int, editorId int) error
	GetProductByID(ctx context.Context, id int) (*model.Product, error)
	GetProductByIDs(ctx context.Context, ids []int) ([]*model.Product, error)
	UpdateProductStatus(ctx context.Context, id, fromStatus int, product *model.Product) error
	UpdateProductStock(ctx context.Context, id int, stock int, editorId int) error
	ListProduct(ctx context.Context, q ListProductQuery) ([]*model.Product, int, error)
}

type ProductDaoImpl struct {
	db *gorm.DB
}

var (
	productOnce sync.Once
	productDao  *ProductDaoImpl
)

func GetProductDao() *ProductDaoImpl {
	productOnce.Do(func() {
		if productDao == nil {
			productDao = &ProductDaoImpl{db: repository.DB}
		}
	})
	return productDao
}

// CreateProduct 创建产品并返回ID
func (p *ProductDaoImpl) CreateProduct(ctx context.Context, product *model.Product) (int, error) {
	result := p.db.WithContext(ctx).Create(product)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("Failed to create product: %v", result.Error)
		return 0, result.Error
	}
	return int(product.ID), nil
}

// UpdateProduct 更新产品信息
func (p *ProductDaoImpl) UpdateProduct(ctx context.Context, product *model.Product) error {
	result := p.db.WithContext(ctx).Model(&model.Product{}).Where("id = ?", product.ID).Updates(product)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("Failed to update product ID %d: %v", product.ID, result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		err := fmt.Errorf("product not found with ID: %d", product.ID)
		log.WithContext(ctx).Error(err)
		return err
	}
	return nil
}

// UpdateStockWithCAS
func (p *ProductDaoImpl) UpdateStockWithCAS(ctx context.Context, id, version, newStock int, editorId int) error {
	ret := p.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ? AND version = ?", id, version).
		Updates(map[string]interface{}{
			"stock":            newStock,
			"latest_editor_id": editorId,
			"version":          gorm.Expr("version + 1"),
		})
	if ret.Error != nil {
		log.WithContext(ctx).Errorf("Failed to update product ID %d: %v", id, ret.Error)
		return ret.Error
	}
	return nil
}

// GetProductByID 根据ID获取产品信息
func (p *ProductDaoImpl) GetProductByID(ctx context.Context, id int) (*model.Product, error) {
	var product model.Product
	result := p.db.WithContext(ctx).Where("id = ?", id).First(&product)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("Failed to get product by ID %d: %v", id, result.Error)
		return nil, result.Error
	}
	return &product, nil
}

func (p *ProductDaoImpl) GetProductByIDs(ctx context.Context, ids []int) ([]*model.Product, error) {
	var products []*model.Product
	result := p.db.WithContext(ctx).Where("id IN ?", ids).Find(&products)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("Failed to get products by IDs %v: %v", ids, result.Error)
		return nil, result.Error
	}
	return products, nil
}

// UpdateProductStock 更新商品库存
func (p *ProductDaoImpl) UpdateProductStock(ctx context.Context, id int, stock int, editorId int) error {
	result := p.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"stock":            stock,
			"latest_editor_id": editorId,
		})
	if result.Error != nil {
		log.WithContext(ctx).Errorf("Failed to update product stock, ID: %d, stock: %d, error: %v", id, stock, result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		err := fmt.Errorf("product not found with ID: %d", id)
		log.WithContext(ctx).Error(err)
		return err
	}
	return nil
}

// ListProduct 查询商品列表
func (p *ProductDaoImpl) ListProduct(ctx context.Context, q ListProductQuery) ([]*model.Product, int, error) {
	var products []*model.Product
	var total int64

	query := p.db.WithContext(ctx).Model(&model.Product{})

	if q.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+q.Keyword+"%")
	}

	if q.Category != "" {
		query = query.Where("category = ?", q.Category)
	}

	// 用户侧只能看到上架的商品
	if q.IsCustomer {
		query = query.Where("status = ?", 1)
	}

	if q.OrderBy == 0 {
		query = query.Order("updated_at DESC")
	} else {
		query = query.Order("updated_at")
	}

	if q.Limit == 0 {
		q.Limit = 10
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		log.WithContext(ctx).Errorf("Failed to count products: %v", err)
		return nil, 0, err
	}

	err = query.Offset(q.Offset).Limit(q.Limit).Find(&products).Error
	if err != nil {
		log.WithContext(ctx).Errorf("Failed to get products ordered by time: %v", err)
		return nil, 0, err
	}

	return products, int(total), nil
}

// UpdateProductStatus 更新商品状态
func (p *ProductDaoImpl) UpdateProductStatus(ctx context.Context, id, fromStatus int, product *model.Product) error {
	result := p.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ? and status=?", id, fromStatus).
		Select("status", "latest_reviewer_id").
		Updates(product)
	if result.Error != nil {
		log.WithContext(ctx).Errorf("Failed to update product status, ID: %d, status: %d, error: %v", id, product.Status, result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		err := fmt.Errorf("product not found with ID: %d", id)
		log.WithContext(ctx).Error(err)
		return err
	}
	return nil
}
