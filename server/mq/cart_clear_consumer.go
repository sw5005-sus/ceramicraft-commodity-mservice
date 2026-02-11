package mq

import (
	"context"
	"encoding/json"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/service"
)

type OrderItem struct {
	ProductID int `json:"product_id"`
}
type OrderCreatedMessage struct {
	UserID        int          `json:"user_id"`
	OrderItemList []*OrderItem `json:"order_item_list"`
}

func clearCartProcess(msg []byte) error {
	var orderCreatedMessage OrderCreatedMessage
	err := json.Unmarshal(msg, &orderCreatedMessage)
	if err != nil {
		log.Logger.Warnf("Failed to unmarshal order created message: %s", string(msg))
		return nil
	}
	if orderCreatedMessage.UserID == 0 || len(orderCreatedMessage.OrderItemList) == 0 {
		log.Logger.Warnf("Invalid order created message: %+v", orderCreatedMessage)
		return nil
	}
	productIds := make([]int, 0)
	for _, item := range orderCreatedMessage.OrderItemList {
		productIds = append(productIds, item.ProductID)
	}
	err = service.GetCartService().DeleteItemByProductIds(context.Background(), orderCreatedMessage.UserID, productIds)
	if err != nil {
		log.Logger.Errorf("Failed to delete user_cart for user ID %d: %v", orderCreatedMessage.UserID, err)
		return err
	}
	log.Logger.Infof("User cart created for user ID %d, order Items %+v", orderCreatedMessage.UserID, orderCreatedMessage.OrderItemList)
	return nil
}
