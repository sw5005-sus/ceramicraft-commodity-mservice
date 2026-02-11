package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/dao"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/model"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/types"
	"gorm.io/gorm"
)

type CartService interface {
	AddItem(ctx context.Context, item *data.CartItemBasicVO) *types.BizError
	UpdateItem(ctx context.Context, item *data.CartItemBasicVO) *types.BizError
	DeleteItem(ctx context.Context, itemId int, userId int) error
	GetCartSelectedItemCnt(ctx context.Context, userId int) (int, error)
	GetCartItems(ctx context.Context, userId int) (*data.CartListVO, error)
	DeleteItemByProductIds(ctx context.Context, userId int, productIds []int) error
	EstimatePrice(ctx context.Context, userId int) (*data.CartPriceEstimateResult, error)
}

var (
	cartServiceInstance CartService
	cartServiceSyncOnce sync.Once
)

func GetCartService() CartService {
	cartServiceSyncOnce.Do(func() {
		cartServiceInstance = &CartServiceImpl{
			cartItemDao: dao.GetShoppingCartItemDao(),
			productDao:  dao.GetProductDao(),
		}
	})
	return cartServiceInstance
}

type CartServiceImpl struct {
	cartItemDao dao.ShoppingCartItemDao
	productDao  dao.ProductDao
}

const (
	ProductStatu_Online                  = 1
	ProductCheckStatus_NotExist          = -1
	ProductCheckStatus_InsufficientStock = -2
	ProductCheckStatus_DBError           = -3

	CartItemStatus_NotExist = -10
)

// AddItem implements CartService.
func (c *CartServiceImpl) AddItem(ctx context.Context, item *data.CartItemBasicVO) *types.BizError {
	existingItems, err := c.cartItemDao.QueryItems(ctx, &model.ShoppingCartItem{
		UserID:    item.UserID,
		ProductID: item.ProductID,
	})
	if err != nil {
		log.Logger.Errorf("CartService: AddItem: Failed to query existing items: %v", err)
		return types.NewBizError(ProductCheckStatus_DBError, fmt.Sprintf("database error: %v", err))
	}
	if len(existingItems) > 0 {
		item.ID = existingItems[0].ID
		item.Quantity += existingItems[0].Quantity
		item.Selected = true
		return c.UpdateItem(ctx, item)
	}
	bizErr := c.checkProductWithItem(ctx, item)
	if bizErr != nil {
		log.Logger.Errorf("CartService: AddItem: Failed to check product with item: %v", err)
		return bizErr
	}
	itemId, err := c.cartItemDao.CreateItem(ctx, &model.ShoppingCartItem{
		UserID:       item.UserID,
		ProductID:    item.ProductID,
		Quantity:     item.Quantity,
		SelectStatus: model.CartItemStatusSelected,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		log.Logger.Errorf("CartService: AddItem: Failed to create cart item: %v", err)
		return types.NewBizError(ProductCheckStatus_DBError, fmt.Sprintf("database error: %v", err))
	}
	item.ID = itemId
	log.Logger.Infof("CartService: AddItem: Added item with ID %d to cart", itemId)
	return nil
}

// DeleteItem implements CartService.
func (c *CartServiceImpl) DeleteItem(ctx context.Context, itemId int, userId int) error {
	err := c.cartItemDao.DeleteItemById(ctx, itemId, userId)
	if err != nil {
		log.Logger.Errorf("CartService: DeleteItem: Failed to delete cart item: %v", err)
		return err
	}
	log.Logger.Infof("CartService: DeleteItem: Deleted item with ID %d from cart", itemId)
	return nil
}

// DeleteItemByProductIds implements CartService.
func (c *CartServiceImpl) DeleteItemByProductIds(ctx context.Context, userId int, productIds []int) error {
	if len(productIds) == 0 {
		log.Logger.Warnf("CartService: DeleteItemByProductIds: No product IDs provided for deletion")
		return nil
	}
	err := c.cartItemDao.DeleteByProductIds(ctx, userId, productIds)
	if err != nil {
		log.Logger.Errorf("CartService: DeleteItemByProductIds: Failed to delete cart items: %v", err)
		return err
	}
	log.Logger.Infof("CartService: DeleteItemByProductIds: Deleted items with product IDs %v from cart", productIds)
	return nil
}

// GetCartItems implements CartService.
func (c *CartServiceImpl) GetCartItems(ctx context.Context, userId int) (*data.CartListVO, error) {
	items, err := c.cartItemDao.QueryItems(ctx, &model.ShoppingCartItem{
		UserID: userId,
	})
	if err != nil {
		log.Logger.Errorf("CartService: GetCartItems: Failed to query cart items: %v", err)
		return nil, err
	}
	if len(items) == 0 {
		log.Logger.Infof("CartService: GetCartItems: No items found for user ID %d", userId)
		return &data.CartListVO{
			CartItems: []data.CartItemDetailVO{},
		}, nil
	}
	productIds := make([]int, 0, len(items))
	productId2Item := make(map[int]*model.ShoppingCartItem)
	for _, item := range items {
		productIds = append(productIds, item.ProductID)
		productId2Item[item.ProductID] = item
	}
	products, err := c.productDao.GetProductByIDs(ctx, productIds)
	if err != nil {
		log.Logger.Errorf("CartService: GetCartItems: Failed to get products by IDs: %v", err)
		return nil, err
	}
	ret := &data.CartListVO{
		CartItems: make([]data.CartItemDetailVO, 0),
	}
	toDeleteProductIds := make([]int, 0)
	for _, product := range products {
		if product.Status != ProductStatu_Online {
			toDeleteProductIds = append(toDeleteProductIds, int(product.ID))
			continue
		}
		if item, exists := productId2Item[int(product.ID)]; exists {
			cartItemDetail := buildCartItemDetail(product, item)
			ret.CartItems = append(ret.CartItems, cartItemDetail)
			if item.SelectStatus == model.CartItemStatusSelected {
				ret.SelectedItemCount += 1
				ret.SelectedPrice += cartItemDetail.TotalPrice
			}
		}
	}
	if len(toDeleteProductIds) > 0 {
		go func() {
			err := c.cartItemDao.DeleteByProductIds(context.Background(), userId, toDeleteProductIds)
			log.Logger.Infof("CartService: GetCartItems: Deleted cart items with invalid products for user ID %d, err: %v", userId, err)
		}()
	}
	sort.Slice(ret.CartItems, func(i, j int) bool {
		return ret.CartItems[i].ID < ret.CartItems[j].ID
	})
	return ret, nil
}

func buildCartItemDetail(product *model.Product, item *model.ShoppingCartItem) data.CartItemDetailVO {
	ret := data.CartItemDetailVO{
		ID: item.ID,
		ProductInfo: types.ProductSimplifiedInfo{
			ID:       int(product.ID),
			Name:     product.Name,
			Category: product.Category,
			Price:    product.Price,
			Stock:    product.Stock,
			PicInfo:  product.PicInfo,
		},
		Quantity:   item.Quantity,
		TotalPrice: int(product.Price) * item.Quantity,
		Selected:   item.SelectStatus == model.CartItemStatusSelected,
	}
	ret.Status = data.CartItemStatus_Normal
	if item.Quantity > int(product.Stock) {
		ret.Status = data.CartItemStatus_OutOfStock
	}
	return ret
}

// GetCartSelectedItemCnt implements CartService.
func (c *CartServiceImpl) GetCartSelectedItemCnt(ctx context.Context, userId int) (int, error) {
	items, err := c.cartItemDao.QueryItems(ctx, &model.ShoppingCartItem{UserID: userId})
	if err != nil {
		log.Logger.Errorf("CartService: GetCartSelectedItemCnt: Failed to query cart items: %v", err)
		return 0, err
	}
	ret := 0
	for _, it := range items {
		if it.SelectStatus == model.CartItemStatusSelected {
			ret += 1
		}
	}
	return ret, nil
}

// UpdateItem implements CartService.
func (c *CartServiceImpl) UpdateItem(ctx context.Context, item *data.CartItemBasicVO) *types.BizError {
	existingItems, err := c.cartItemDao.GetItemById(ctx, item.ID)
	if err != nil {
		log.Logger.Errorf("CartService: UpdateItem: Failed to get item by ID: %v", err)
		return types.NewBizError(ProductCheckStatus_DBError, fmt.Sprintf("database error: %v", err))
	}
	if existingItems == nil || existingItems.UserID != item.UserID {
		log.Logger.Errorf("CartService: UpdateItem: Item not found or does not belong to user")
		return types.NewBizError(CartItemStatus_NotExist, "cart item not found or does not belong to user")
	}
	bizErr := c.checkProductWithItem(ctx, item)
	if bizErr != nil {
		return bizErr
	}
	existingItems.Quantity = item.Quantity
	if item.Selected {
		existingItems.SelectStatus = model.CartItemStatusSelected
	} else {
		existingItems.SelectStatus = model.CartItemStatusUnselected
	}
	existingItems.UpdatedAt = time.Now()
	err = c.cartItemDao.UpdateItem(ctx, existingItems)
	if err != nil {
		log.Logger.Errorf("CartService: UpdateItem: Failed to update cart item: %v", err)
		return types.NewBizError(ProductCheckStatus_DBError, fmt.Sprintf("database error: %v", err))
	}
	log.Logger.Infof("CartService: UpdateItem: Updated item with ID %d in cart", item.ID)
	return nil
}

func (c *CartServiceImpl) checkProductWithItem(ctx context.Context, item *data.CartItemBasicVO) *types.BizError {
	product, err := c.productDao.GetProductByID(ctx, item.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewBizError(ProductCheckStatus_NotExist, fmt.Sprintf("product not found with ID: %d", item.ProductID))
		}
		log.Logger.Errorf("CartService: UpdateItem: Failed to get product by ID: %v", err)
		return types.NewBizError(ProductCheckStatus_DBError, fmt.Sprintf("database error: %v", err))
	}
	if product == nil || product.Status != ProductStatu_Online {
		return types.NewBizError(ProductCheckStatus_NotExist, "product not found or not available")
	}
	if product.Stock < int64(item.Quantity) {
		return types.NewBizError(ProductCheckStatus_InsufficientStock, fmt.Sprintf("insufficient stock for product ID %d", item.ProductID))
	}
	return nil
}

func (c *CartServiceImpl) EstimatePrice(ctx context.Context, userId int) (*data.CartPriceEstimateResult, error) {
	items, err := c.cartItemDao.QueryItems(ctx, &model.ShoppingCartItem{
		UserID:       userId,
		SelectStatus: model.CartItemStatusSelected,
	})
	if err != nil {
		log.Logger.Errorf("CartService: EstimatePrice: Failed to query cart items: %v", err)
		return nil, err
	}
	ret := &data.CartPriceEstimateResult{}
	if len(items) == 0 {
		log.Logger.Infof("CartService: EstimatePrice: No items found for user ID %d", userId)
		return ret, nil
	}
	productIds := make([]int, 0, len(items))
	productId2Item := make(map[int]*model.ShoppingCartItem)
	for _, item := range items {
		productIds = append(productIds, item.ProductID)
		productId2Item[item.ProductID] = item
	}
	products, err := c.productDao.GetProductByIDs(ctx, productIds)
	if err != nil {
		log.Logger.Errorf("CartService: EstimatePrice: Failed to get products by IDs: %v", err)
		return nil, err
	}
	for _, product := range products {
		if product.Status != ProductStatu_Online {
			continue
		}
		if item, exists := productId2Item[int(product.ID)]; exists {
			if item.SelectStatus == model.CartItemStatusSelected {
				ret.ProductPrice += int(product.Price) * item.Quantity
			}
		}
	}
	ret.ShippingPrice = getShipmentPrice(ret.ProductPrice)
	ret.Tax = getTaxPrice(ret.ProductPrice)
	ret.Total = ret.ShippingPrice + ret.Tax + ret.ProductPrice
	return ret, nil
}

const (
	freeShipPrice = 30000 // $300.00 in cents
	shipPrice     = 800    // $8.00 in cents
)

func getTaxPrice(productPrice int) int {
	return productPrice * 9 / 100
}

func getShipmentPrice(productPrice int) int {
	if productPrice > freeShipPrice {
		return 0
	} else {
		return shipPrice
	}
}
