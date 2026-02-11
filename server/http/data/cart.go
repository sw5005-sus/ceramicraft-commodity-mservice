package data

import "github.com/sw5005-sus/ceramicraft-commodity-mservice/server/types"

type CartItemBasicVO struct {
	ID        int  `json:"id"`
	UserID    int  `json:"user_id"`
	ProductID int  `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
	Selected  bool `json:"selected"`
}

const (
	CartItemStatus_Normal     = 1
	CartItemStatus_OutOfStock = 2
)

type CartItemDetailVO struct {
	ID          int                         `json:"id"`
	ProductInfo types.ProductSimplifiedInfo `json:"product_info"`
	Quantity    int                         `json:"quantity"`
	TotalPrice  int                         `json:"total_price"`
	Status      int                         `json:"status"` // 1: normal, 2: out of stock
	Selected    bool                        `json:"selected"`
}

type CartListVO struct {
	CartItems         []CartItemDetailVO `json:"cart_items"`
	SelectedItemCount int                `json:"selected_item_count"`
	SelectedPrice     int                `json:"selected_price"`
}

type CartPriceEstimateResult struct {
	ProductPrice  int `json:"product_price"`
	ShippingPrice int `json:"shipping_price"`
	Tax           int `json:"tax"`
	Total         int `json:"total"`
}
