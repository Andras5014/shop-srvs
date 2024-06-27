package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"shop_srvs/inventory_srv/global"
	"shop_srvs/inventory_srv/model"
	"shop_srvs/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

func (i *InventoryServer) SetInv(ctx context.Context, in *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	// 设置库存
	var inv model.Inventory
	global.DB.Where("goods = ?", in.GoodsId).First(&inv)
	inv.Goods = in.GoodsId
	inv.Stocks = in.Num

	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

func (i *InventoryServer) InvDetail(ctx context.Context, in *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv model.Inventory
	if result := global.DB.Where("goods = ?", in.GoodsId).First(&inv); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
	}
	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

func (i *InventoryServer) Sell(ctx context.Context, in *proto.SellInfo) (*emptypb.Empty, error) {
	// 保证全部成功 事务
	tx := global.DB.Begin()
	for _, goodInfo := range in.GoodsInfo {
		var inv model.Inventory
		if result := tx.Where("goods = ?", goodInfo.GoodsId).First(&inv); result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		}
		// 库存判断是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback()
			return nil, status.Errorf(codes.OutOfRange, "库存不足")
		}
		// 库存扣减
		inv.Stocks -= goodInfo.Num
		tx.Save(&inv)
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}

func (i *InventoryServer) Reback(ctx context.Context, in *proto.SellInfo) (*emptypb.Empty, error) {
	// 库存归还 1、订单超时 2、订单创建失败取消
	tx := global.DB.Begin()
	for _, goodInfo := range in.GoodsInfo {
		var inv model.Inventory
		if result := tx.Where("goods = ?", goodInfo.GoodsId).First(&inv); result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		}
		// 扣减会出现数据不一致的问题 锁库存
		inv.Stocks += goodInfo.Num
		tx.Save(&inv)
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}
func (i *InventoryServer) mustEmbedUnimplementedInventoryServer() {

}
