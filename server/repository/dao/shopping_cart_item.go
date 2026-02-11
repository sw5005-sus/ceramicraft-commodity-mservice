package dao

import (
	"context"
	"errors"
	"sync"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/repository/model"
	"gorm.io/gorm"
)

type ShoppingCartItemDao interface {
	CreateItem(ctx context.Context, item *model.ShoppingCartItem) (itemId int, err error)
	UpdateItem(ctx context.Context, item *model.ShoppingCartItem) error
	DeleteItemById(ctx context.Context, id int, userId int) error
	DeleteByProductIds(ctx context.Context, userId int, productIds []int) error
	GetItemById(ctx context.Context, id int) (item *model.ShoppingCartItem, err error)
	QueryItems(ctx context.Context, query *model.ShoppingCartItem) (item []*model.ShoppingCartItem, err error)
}

var (
	shoppingCartItemDaoInstance ShoppingCartItemDao
	shoppingCartItemDaoSyncOnce sync.Once
)

func GetShoppingCartItemDao() ShoppingCartItemDao {
	shoppingCartItemDaoSyncOnce.Do(func() {
		shoppingCartItemDaoInstance = &ShoppingCartItemDaoImpl{
			db: repository.DB,
		}
	})
	return shoppingCartItemDaoInstance
}

type ShoppingCartItemDaoImpl struct {
	db *gorm.DB
}

// DeleteByUserAndProductIds implements ShoppingCartItemDao.
func (s *ShoppingCartItemDaoImpl) DeleteByProductIds(ctx context.Context, userId int, productIds []int) error {
	ret := s.db.WithContext(ctx).Where("user_id = ? AND product_id IN ?", userId, productIds).Delete(&model.ShoppingCartItem{})
	if ret.Error != nil {
		log.Logger.Errorf("ShoppingCartItemDao: DeleteByProductIds: Failed to delete items: %v", ret.Error)
		return ret.Error
	}
	log.Logger.Infof("ShoppingCartItemDao: DeleteByProductIds: Deleted %d items for user ID %d", ret.RowsAffected, userId)
	return nil
}

// CreateItem implements ShoppingCartItemDao.
func (s *ShoppingCartItemDaoImpl) CreateItem(ctx context.Context, item *model.ShoppingCartItem) (itemId int, err error) {
	ret := s.db.WithContext(ctx).Create(item)
	if ret.Error != nil {
		log.Logger.Errorf("ShoppingCartItemDao: CreateItem: Failed to create item: %v", ret.Error)
		return -1, ret.Error
	}
	log.Logger.Infof("ShoppingCartItemDao: CreateItem: Created item with ID %d", item.ID)
	return item.ID, nil
}

// DeleteItem implements ShoppingCartItemDao.
func (s *ShoppingCartItemDaoImpl) DeleteItemById(ctx context.Context, id int, userId int) error {
	ret := s.db.WithContext(ctx).Delete(&model.ShoppingCartItem{ID: id, UserID: userId})
	if ret.Error != nil {
		log.Logger.Errorf("ShoppingCartItemDao: DeleteItem: Failed to delete item: %v", ret.Error)
		return ret.Error
	}
	if ret.RowsAffected == 0 {
		log.Logger.Warnf("ShoppingCartItemDao: DeleteItem: No item found with ID %d to delete", id)
		return nil
	}
	log.Logger.Infof("ShoppingCartItemDao: DeleteItem: Deleted item with ID %d", id)
	return nil
}

// GetItemById implements ShoppingCartItemDao.
func (s *ShoppingCartItemDaoImpl) GetItemById(ctx context.Context, id int) (item *model.ShoppingCartItem, err error) {
	ret := s.db.WithContext(ctx).First(&item, id)
	if ret.Error != nil {
		if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
			log.Logger.Warnf("ShoppingCartItemDao: GetItemById: No item found with ID %d", id)
			return nil, nil
		} else {
			log.Logger.Errorf("ShoppingCartItemDao: GetItemById: Failed to get item: %v", ret.Error)
			return nil, ret.Error
		}
	}
	return item, nil
}

// GetItemByUserId implements ShoppingCartItemDao.
func (s *ShoppingCartItemDaoImpl) QueryItems(ctx context.Context, query *model.ShoppingCartItem) (item []*model.ShoppingCartItem, err error) {
	ret := s.db.WithContext(ctx).Where(query).Find(&item)
	if ret.Error != nil {
		log.Logger.Errorf("ShoppingCartItemDao: GetItemByUserId: Failed to get items: %v", ret.Error)
		return nil, ret.Error
	}
	return item, nil
}

// UpdateItem implements ShoppingCartItemDao.
func (s *ShoppingCartItemDaoImpl) UpdateItem(ctx context.Context, item *model.ShoppingCartItem) error {
	ret := s.db.WithContext(ctx).Save(item)
	if ret.Error != nil {
		log.Logger.Errorf("ShoppingCartItemDao: UpdateItem: Failed to update item: %v", ret.Error)
		return ret.Error
	}
	if ret.RowsAffected == 0 {
		log.Logger.Warnf("ShoppingCartItemDao: UpdateItem: No item found with ID %d to update", item.ID)
		return nil
	}
	return nil
}
