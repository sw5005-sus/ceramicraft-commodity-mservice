package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/dao/mocks"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/model"
	"github.com/golang/mock/gomock"
	"gorm.io/gorm"
)

func TestGetCartService(t *testing.T) {
	t.Run("GetCartService returns the same instance on multiple calls", func(t *testing.T) {
		service1 := GetCartService()
		service2 := GetCartService()
		if service1 != service2 {
			t.Errorf("Expected the same instance, got different instances")
		}
	})

	t.Run("GetCartService initializes the CartServiceImpl", func(t *testing.T) {
		service := GetCartService()
		_, ok := service.(*CartServiceImpl)
		if !ok {
			t.Errorf("Expected CartServiceImpl, got %T", service)
		}
	})
}

func TestAddItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	productId := uint(1)
	t.Run("AddItem adds a new item to the cart", func(t *testing.T) {
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		item := &data.CartItemBasicVO{
			UserID:    1,
			ProductID: int(productId),
			Quantity:  2,
		}
		ctx := context.Background()
		cartItemDao.EXPECT().QueryItems(ctx, gomock.Any()).Return([]*model.ShoppingCartItem{}, nil)
		product := &model.Product{
			Model:  gorm.Model{ID: uint(item.ProductID)},
			Stock:  10,
			Status: ProductStatu_Online,
		}
		productDao.EXPECT().GetProductByID(ctx, 1).Return(product, nil)
		cartItemDao.EXPECT().CreateItem(ctx, gomock.Any()).Return(1, nil).Times(1)

		err := cartService.AddItem(ctx, item)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("AddItem updates existing item quantity", func(t *testing.T) {
		ctx := context.Background()
		item := &data.CartItemBasicVO{
			UserID:    1,
			ProductID: 1,
			Quantity:  3,
		}
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}

		cartItem := &model.ShoppingCartItem{
			ID:        1,
			UserID:    1,
			ProductID: item.ProductID,
		}
		cartItemDao.EXPECT().QueryItems(ctx, gomock.Eq(&model.ShoppingCartItem{
			UserID:    item.UserID,
			ProductID: item.ProductID,
		})).Return([]*model.ShoppingCartItem{cartItem}, nil)
		product := &model.Product{
			Model:  gorm.Model{ID: uint(item.ProductID)},
			Stock:  10,
			Status: ProductStatu_Online,
		}
		cartItemDao.EXPECT().GetItemById(ctx, cartItem.ID).Return(cartItem, nil).Times(1)
		productDao.EXPECT().GetProductByID(ctx, item.ProductID).Return(product, nil)
		cartItemDao.EXPECT().UpdateItem(ctx, gomock.AssignableToTypeOf(cartItem)).Return(nil).Times(1)
		err := cartService.AddItem(ctx, item)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if cartItem.Quantity != item.Quantity {
			t.Errorf("Expected quantity %d, got %d", item.Quantity, cartItem.Quantity)
		}
		if cartItem.SelectStatus != model.CartItemStatusSelected {
			t.Errorf("Expected SelectStatus %d, got %d", model.CartItemStatusSelected, cartItem.SelectStatus)
		}
	})

	t.Run("AddItem returns error for non-existent product", func(t *testing.T) {
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		ctx := context.Background()
		item := &data.CartItemBasicVO{
			UserID:    1,
			ProductID: 9999, // Assuming this product does not exist
			Quantity:  1,
		}
		cartItemDao.EXPECT().QueryItems(ctx, gomock.Any()).Return([]*model.ShoppingCartItem{}, nil)
		product := &model.Product{
			Model: gorm.Model{ID: uint(item.ProductID)},
			Stock: int64(item.Quantity) + 10,
		}
		productDao.EXPECT().GetProductByID(ctx, item.ProductID).Return(product, nil).Times(1)
		err := cartService.AddItem(ctx, item)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("AddItem returns error for insufficient stock", func(t *testing.T) {
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		ctx := context.Background()
		item := &data.CartItemBasicVO{
			UserID:    1,
			ProductID: 1,   // Assuming this product exists but has insufficient stock
			Quantity:  100, // Assuming the stock is less than 100
		}

		cartItemDao.EXPECT().QueryItems(ctx, gomock.Any()).Return([]*model.ShoppingCartItem{}, nil)
		product := &model.Product{
			Model:  gorm.Model{ID: uint(item.ProductID)},
			Stock:  int64(item.Quantity) - 10,
			Status: ProductStatu_Online,
		}
		productDao.EXPECT().GetProductByID(ctx, item.ProductID).Return(product, nil)
		err := cartService.AddItem(ctx, item)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})
}

func TestDeleteItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("DeleteItem successfully deletes an item from the cart", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
		}
		itemId := 1
		userId := 1

		cartItemDao.EXPECT().DeleteItemById(ctx, itemId, userId).Return(nil).Times(1)

		err := cartService.DeleteItem(ctx, itemId, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("DeleteItem logs error when deletion fails", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
		}
		itemId := 1
		userId := 1

		cartItemDao.EXPECT().DeleteItemById(ctx, itemId, userId).Return(errors.New("database error")).Times(1)

		err := cartService.DeleteItem(ctx, itemId, userId)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})
}
func TestDeleteItemByProductIds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("DeleteItemByProductIds successfully deletes items", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
		}
		userId := 1
		productIds := []int{1, 2, 3}

		cartItemDao.EXPECT().DeleteByProductIds(ctx, userId, productIds).Return(nil).Times(1)

		err := cartService.DeleteItemByProductIds(ctx, userId, productIds)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("DeleteItemByProductIds returns error when deletion fails", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
		}
		userId := 1
		productIds := []int{1, 2, 3}

		cartItemDao.EXPECT().DeleteByProductIds(ctx, userId, productIds).Return(errors.New("database error")).Times(1)

		err := cartService.DeleteItemByProductIds(ctx, userId, productIds)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("DeleteItemByProductIds does nothing when no product IDs are provided", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
		}
		userId := 1
		productIds := []int{}

		err := cartService.DeleteItemByProductIds(ctx, userId, productIds)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestGetCartItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("GetCartItems returns cart items for a user", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		userId := 1

		cartItems := []*model.ShoppingCartItem{
			{ID: 10, UserID: userId, ProductID: 1, Quantity: 2, SelectStatus: model.CartItemStatusSelected},
			{ID: 2, UserID: userId, ProductID: 2, Quantity: 1, SelectStatus: model.CartItemStatusUnselected},
		}
		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{UserID: userId}).Return(cartItems, nil)

		products := []*model.Product{
			{Model: gorm.Model{ID: 1}, Name: "Product 1", Price: 100, Stock: 10, Status: ProductStatu_Online},
			{Model: gorm.Model{ID: 2}, Name: "Product 2", Price: 200, Stock: 5, Status: ProductStatu_Online},
		}
		productDao.EXPECT().GetProductByIDs(ctx, []int{1, 2}).Return(products, nil)

		result, err := cartService.GetCartItems(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result.CartItems) != 2 {
			t.Errorf("Expected 2 cart items, got %d", len(result.CartItems))
		}
		if result.SelectedItemCount != 1 {
			t.Errorf("Expected 1 selected item, got %d", result.SelectedItemCount)
		}
		if result.SelectedPrice != 200 { // 2*100 + 0*200
			t.Errorf("Expected total price 200, got %d", result.SelectedItemCount)
		}
		if result.SelectedItemCount != 1 {
			t.Errorf("Expected 1 selected item, got %d", result.SelectedItemCount)
		}
		if result.CartItems[0].ID != 2 || result.CartItems[1].ID != 10 {
			t.Errorf("Cart item ordinal do not match expected values")
		}
	})

	t.Run("GetCartItems returns error when querying items fails", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		userId := 1

		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{UserID: userId}).Return(nil, errors.New("database error"))

		_, err := cartService.GetCartItems(ctx, userId)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("GetCartItems returns empty list when no items found", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		userId := 1

		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{UserID: userId}).Return([]*model.ShoppingCartItem{}, nil)

		result, err := cartService.GetCartItems(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result.CartItems) != 0 {
			t.Errorf("Expected 0 cart items, got %d", len(result.CartItems))
		}
	})

	t.Run("GetCartItems deletes invalid products from cart", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		userId := 1

		cartItems := []*model.ShoppingCartItem{
			{ID: 1, UserID: userId, ProductID: 1, Quantity: 2, SelectStatus: model.CartItemStatusSelected},
		}
		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{UserID: userId}).Return(cartItems, nil)

		products := []*model.Product{
			{Model: gorm.Model{ID: 1}, Name: "Product 1", Price: 100, Stock: 0, Status: -1}, // Invalid product
		}
		productDao.EXPECT().GetProductByIDs(ctx, []int{1}).Return(products, nil)
		var wg sync.WaitGroup
		wg.Add(1)
		cartItemDao.EXPECT().DeleteByProductIds(gomock.Any(), userId, []int{1}).DoAndReturn(func(ctx context.Context, userId int, productIds []int) error {
			defer wg.Done() // Signal that the goroutine is done
			return nil
		})

		result, err := cartService.GetCartItems(ctx, userId)
		wg.Wait()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result.CartItems) != 0 {
			t.Errorf("Expected 0 cart items, got %d", len(result.CartItems))
		}
	})
}
func TestGetCartSelectedItemCnt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("GetCartSelectedItemCnt returns correct count of selected items", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
		}
		userId := 1

		cartItems := []*model.ShoppingCartItem{
			{UserID: userId, SelectStatus: model.CartItemStatusSelected},
			{UserID: userId, SelectStatus: model.CartItemStatusUnselected},
			{UserID: userId, SelectStatus: model.CartItemStatusSelected},
		}
		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{UserID: userId}).Return(cartItems, nil)

		count, err := cartService.GetCartSelectedItemCnt(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("Expected selected item count 2, got %d", count)
		}
	})

	t.Run("GetCartSelectedItemCnt returns error when querying items fails", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
		}
		userId := 1

		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{UserID: userId}).Return(nil, errors.New("database error"))

		count, err := cartService.GetCartSelectedItemCnt(ctx, userId)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}
	})
}
func TestUpdateItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("UpdateItem successfully updates an existing item", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		item := &data.CartItemBasicVO{
			ID:        1,
			UserID:    1,
			ProductID: 1,
			Quantity:  2,
			Selected:  true,
		}
		existingItem := &model.ShoppingCartItem{
			ID:           1,
			UserID:       1,
			ProductID:    1,
			Quantity:     1,
			SelectStatus: model.CartItemStatusUnselected,
		}
		cartItemDao.EXPECT().GetItemById(ctx, item.ID).Return(existingItem, nil).Times(1)
		product := &model.Product{
			Model:  gorm.Model{ID: uint(item.ProductID)},
			Stock:  10,
			Status: ProductStatu_Online,
		}
		productDao.EXPECT().GetProductByID(ctx, item.ProductID).Return(product, nil).Times(1)
		cartItemDao.EXPECT().UpdateItem(ctx, gomock.Any()).Return(nil).Times(1)

		err := cartService.UpdateItem(ctx, item)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if existingItem.Quantity != item.Quantity {
			t.Errorf("Expected quantity %d, got %d", item.Quantity, existingItem.Quantity)
		}
		if existingItem.SelectStatus != model.CartItemStatusSelected {
			t.Errorf("Expected SelectStatus %d, got %d", model.CartItemStatusSelected, existingItem.SelectStatus)
		}
	})

	t.Run("UpdateItem returns error for non-existent item", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		item := &data.CartItemBasicVO{
			ID:        1,
			UserID:    1,
			ProductID: 1,
			Quantity:  2,
			Selected:  true,
		}
		cartItemDao.EXPECT().GetItemById(ctx, item.ID).Return(nil, nil).Times(1)

		err := cartService.UpdateItem(ctx, item)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("UpdateItem returns error when product is not found", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		item := &data.CartItemBasicVO{
			ID:        1,
			UserID:    1,
			ProductID: 1,
			Quantity:  2,
			Selected:  true,
		}
		existingItem := &model.ShoppingCartItem{
			ID:           1,
			UserID:       1,
			ProductID:    1,
			Quantity:     1,
			SelectStatus: model.CartItemStatusUnselected,
		}
		cartItemDao.EXPECT().GetItemById(ctx, item.ID).Return(existingItem, nil).Times(1)
		productDao.EXPECT().GetProductByID(ctx, item.ProductID).Return(nil, errors.New("product not found")).Times(1)

		err := cartService.UpdateItem(ctx, item)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("UpdateItem returns error when product stock is not enough", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		item := &data.CartItemBasicVO{
			ID:        1,
			UserID:    1,
			ProductID: 1,
			Quantity:  20,
			Selected:  true,
		}
		existingItem := &model.ShoppingCartItem{
			ID:           1,
			UserID:       1,
			ProductID:    1,
			Quantity:     1,
			SelectStatus: model.CartItemStatusUnselected,
		}
		product := &model.Product{
			Model: gorm.Model{ID: uint(item.ProductID)},
			Stock: 10,
		}
		cartItemDao.EXPECT().GetItemById(ctx, item.ID).Return(existingItem, nil).Times(1)
		productDao.EXPECT().GetProductByID(ctx, item.ProductID).Return(product, nil).Times(1)

		err := cartService.UpdateItem(ctx, item)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("UpdateItem returns error when updating fails", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		item := &data.CartItemBasicVO{
			ID:        1,
			UserID:    1,
			ProductID: 1,
			Quantity:  2,
			Selected:  true,
		}
		existingItem := &model.ShoppingCartItem{
			ID:           1,
			UserID:       1,
			ProductID:    1,
			Quantity:     1,
			SelectStatus: model.CartItemStatusUnselected,
		}
		cartItemDao.EXPECT().GetItemById(ctx, item.ID).Return(existingItem, nil).Times(1)
		product := &model.Product{
			Model:  gorm.Model{ID: uint(item.ProductID)},
			Stock:  10,
			Status: ProductStatu_Online,
		}
		productDao.EXPECT().GetProductByID(ctx, item.ProductID).Return(product, nil).Times(1)
		cartItemDao.EXPECT().UpdateItem(ctx, gomock.Any()).Return(errors.New("update failed")).Times(1)

		err := cartService.UpdateItem(ctx, item)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})
}
func TestEstimatePrice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("EstimatePrice returns correct total price for selected items", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		userId := 1

		cartItems := []*model.ShoppingCartItem{
			{ID: 1, UserID: userId, ProductID: 1, Quantity: 2, SelectStatus: model.CartItemStatusSelected},
			{ID: 2, UserID: userId, ProductID: 2, Quantity: 1, SelectStatus: model.CartItemStatusSelected},
		}
		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{
			UserID:       userId,
			SelectStatus: model.CartItemStatusSelected,
		}).Return(cartItems, nil)

		products := []*model.Product{
			{Model: gorm.Model{ID: 1}, Price: 100, Status: ProductStatu_Online},
			{Model: gorm.Model{ID: 2}, Price: 200, Status: ProductStatu_Online},
		}
		productDao.EXPECT().GetProductByIDs(ctx, []int{1, 2}).Return(products, nil)

		result, err := cartService.EstimatePrice(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		expectedTotal := 100*2 + 200*1 + getTaxPrice(100*2+200*1) + getShipmentPrice(100*2+200*1)
		if result.Total != expectedTotal {
			t.Errorf("Expected total %d, got %d", expectedTotal, result.Total)
		}
	})

	t.Run("EstimatePrice returns zero when no selected items", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		userId := 1

		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{
			UserID:       userId,
			SelectStatus: model.CartItemStatusSelected,
		}).Return([]*model.ShoppingCartItem{}, nil)

		result, err := cartService.EstimatePrice(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.Total != 0 {
			t.Errorf("Expected total 0, got %d", result.Total)
		}
	})

	t.Run("EstimatePrice returns error when querying items fails", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		userId := 1

		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{
			UserID:       userId,
			SelectStatus: model.CartItemStatusSelected,
		}).Return(nil, errors.New("database error"))

		_, err := cartService.EstimatePrice(ctx, userId)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("EstimatePrice ignores unavailable products", func(t *testing.T) {
		ctx := context.Background()
		cartItemDao := mocks.NewMockShoppingCartItemDao(ctrl)
		productDao := mocks.NewMockProductDao(ctrl)
		cartService := &CartServiceImpl{
			cartItemDao: cartItemDao,
			productDao:  productDao,
		}
		userId := 1

		cartItems := []*model.ShoppingCartItem{
			{ID: 1, UserID: userId, ProductID: 1, Quantity: 2, SelectStatus: model.CartItemStatusSelected},
			{ID: 2, UserID: userId, ProductID: 2, Quantity: 1, SelectStatus: model.CartItemStatusSelected},
		}
		cartItemDao.EXPECT().QueryItems(ctx, &model.ShoppingCartItem{
			UserID:       userId,
			SelectStatus: model.CartItemStatusSelected,
		}).Return(cartItems, nil)

		products := []*model.Product{
			{Model: gorm.Model{ID: 1}, Price: 100, Status: ProductStatu_Online},
			{Model: gorm.Model{ID: 2}, Price: 200, Status: -1}, // Unavailable product
		}
		productDao.EXPECT().GetProductByIDs(ctx, []int{1, 2}).Return(products, nil)

		result, err := cartService.EstimatePrice(ctx, userId)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		expectedTotal := 100*2 + getTaxPrice(100*2) + getShipmentPrice(100*2)
		if result.Total != expectedTotal {
			t.Errorf("Expected total %d, got %d", expectedTotal, result.Total)
		}
	})
}
