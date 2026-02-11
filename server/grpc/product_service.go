package grpc

import (
	"context"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/common/productpb"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/service"
)

type ProductService struct {
	productpb.UnimplementedProductServiceServer
}

func (p *ProductService) UpdateStockWithCAS(ctx context.Context, req *productpb.UpdateStockWithCASRequest) (*productpb.UpdateStockWithCASResponse, error) {
	// execute
	err := service.GetProductServiceInstance().UpdateStockWithCAS(ctx, int(req.Id), int(req.Deta))
	
	// failed
	if err != nil {
		return &productpb.UpdateStockWithCASResponse{
			Base: &productpb.BaseResponse{
				Code: int32(productpb.ResponseCode_INTERNAL_ERROR),
				Msg: productpb.ResponseCode_name[int32(productpb.ResponseCode_INTERNAL_ERROR)],
			},
		}, err
	}

	// success
	return &productpb.UpdateStockWithCASResponse{
		Base: &productpb.BaseResponse{
			Code: int32(productpb.ResponseCode_SUCCESS),
			Msg: productpb.ResponseCode_name[int32(productpb.ResponseCode_SUCCESS)],
		},
	}, nil
}

func (p *ProductService) GetProductList(ctx context.Context, req *productpb.GetProductListRequest) (*productpb.GetProductListResponse, error) {
	productList := make([]*productpb.Product, 0)
	for _, id := range req.Ids {
		productRaw, err := service.GetProductServiceInstance().GetProductByID(ctx, int(id))
		if err != nil {
			continue
		}

		productList = append(productList, &productpb.Product{
			Id: id,
			Name: productRaw.Name,
			Stock: productRaw.Stock,
			Price: productRaw.Price,
			Status: productRaw.Status,
		})
	}

	return &productpb.GetProductListResponse{
		Base: &productpb.BaseResponse{
			Code: int32(productpb.ResponseCode_SUCCESS),
			Msg: productpb.ResponseCode_name[int32(productpb.ResponseCode_SUCCESS)],
		},
		Products: productList,
	}, nil
}
