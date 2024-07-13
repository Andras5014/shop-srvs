package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

//func (i *InventoryServer) Sell(ctx context.Context, in *proto.SellInfo) (*emptypb.Empty, error) {
//	tx := global.DB.Begin()
//	tx.Set("gorm:query_option", "FOR UPDATE") // 悲观锁
//	for _, goodInfo := range in.GoodsInfo {
//		var inv model.Inventory
//		if err := tx.Where("goods = ?", goodInfo.GoodsId).First(&inv).Error; err != nil {
//			tx.Rollback()
//			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
//		}
//
//		if inv.Stocks < goodInfo.Num {
//			tx.Rollback()
//			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
//		}
//
//		inv.Stocks -= goodInfo.Num
//		if err := tx.Save(&inv).Error; err != nil {
//			tx.Rollback()
//			return nil, status.Errorf(codes.Internal, "更新库存失败")
//		}
//	}
//
//	if err := tx.Commit().Error; err != nil {
//		tx.Rollback()
//		return nil, status.Errorf(codes.Internal, "事务提交失败: %v", err)
//	}
//	return &emptypb.Empty{}, nil
//}

// 可重复读 全部商品加锁
//func (*InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
//	// 初始化一个用于存放所有商品锁的slice
//	var mutexes []*redsync.Mutex
//
//	// 首先获取所有商品的分布式锁
//	for _, goodInfo := range req.GoodsInfo {
//		mutex := global.Redsync.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
//		if err := mutex.Lock(); err != nil {
//			// 如果获取锁失败，则释放之前获取的所有锁
//			for _, m := range mutexes {
//				m.Unlock()
//			}
//			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常: %v", err)
//		}
//		// 将锁添加到列表中
//		mutexes = append(mutexes, mutex)
//	}
//
//	// 开始数据库事务
//	tx := global.DB.Begin()
//
//	sellDetail := model.StockSellDetail{
//		OrderSn: req.OrderSn,
//		Status:  1,
//	}
//	var details []model.GoodsDetail
//
//	// 处理每个商品的库存扣减
//	for _, goodInfo := range req.GoodsInfo {
//		details = append(details, model.GoodsDetail{
//			Goods: goodInfo.GoodsId,
//			Num:   goodInfo.Num,
//		})
//
//		var inv model.Inventory
//		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
//			tx.Rollback()
//			// 在返回前释放所有锁
//			for _, m := range mutexes {
//				m.Unlock()
//			}
//			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
//		}
//		if inv.Stocks < goodInfo.Num {
//			tx.Rollback()
//			for _, m := range mutexes {
//				m.Unlock()
//			}
//			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
//		}
//		inv.Stocks -= goodInfo.Num
//		tx.Save(&inv)
//	}
//
//	sellDetail.Detail = details
//	if result := tx.Create(&sellDetail); result.RowsAffected == 0 {
//		tx.Rollback()
//		for _, m := range mutexes {
//			m.Unlock()
//		}
//		return nil, status.Errorf(codes.Internal, "保存库存扣减历史失败")
//	}
//
//	// 提交事务
//	if result := tx.Commit(); result.Error != nil {
//		tx.Rollback()
//		for _, m := range mutexes {
//			m.Unlock()
//		}
//		return nil, status.Errorf(codes.Internal, "事务提交失败: %v", result.Error)
//	}
//
//	// 释放所有锁
//	for _, m := range mutexes {
//		m.Unlock()
//	}
//
//	return &emptypb.Empty{}, nil
//}

func (i *InventoryServer) FastSell(ctx context.Context, in *proto.SellInfo) (*emptypb.Empty, error) {
	if len(in.GoodsInfo) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "秒杀活动只能涉及一件商品")
	}
	goodInfo := in.GoodsInfo[0]

	// 使用 singleflight 防止多次请求同时处理同一商品
	key := fmt.Sprintf("goods_%d", goodInfo.GoodsId)
	_, err, _ := global.Sf.Do(key, func() (interface{}, error) {
		// 获取分布式锁
		mutex := global.Redsync.NewMutex(fmt.Sprintf("lock_goods_%d", goodInfo.GoodsId))
		if err := mutex.LockContext(ctx); err != nil {
			return nil, status.Errorf(codes.Internal, "无法获取商品ID %d 的锁: %v", goodInfo.GoodsId, err)
		}
		defer mutex.Unlock()

		// 执行更新库存的 SQL 语句
		result := global.DB.Exec("UPDATE inventories SET stocks = stocks - ? WHERE goods = ? AND stocks - ?>0", goodInfo.Num, goodInfo.GoodsId, goodInfo.Num)
		if result.Error != nil {
			return nil, status.Errorf(codes.Internal, "更新库存失败: %v", result.Error)
		}
		// 检查是否有行受影响，如果没有则说明库存不足
		if result.RowsAffected == 0 {
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足，商品ID: %d", goodInfo.GoodsId)
		}

		return &emptypb.Empty{}, nil
	})

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// 全交给数据库
func (i *InventoryServer) Sell(ctx context.Context, in *proto.SellInfo) (*emptypb.Empty, error) {
	// 开始一个事务

	tx := global.DB.Begin()
	sellDetail := model.StockSellDetail{
		OrderSn: in.OrderSn,
		Status:  1,
	}
	var details []model.GoodsDetail
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "无法开始事务: %v", tx.Error)
	}

	for _, goodInfo := range in.GoodsInfo {
		// 执行更新库存的 SQL 语句
		details = append(details, model.GoodsDetail{
			Goods: goodInfo.GoodsId,
			Num:   goodInfo.Num,
		})
		result := tx.Exec("UPDATE inventories SET stocks = stocks - ? WHERE goods = ? AND stocks - ?>0", goodInfo.Num, goodInfo.GoodsId, goodInfo.Num)
		if result.Error != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "更新库存失败: %v", result.Error)
		}
		// 检查是否有行受影响，如果没有则说明库存不足
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足，商品ID: %d", goodInfo.GoodsId)
		}
	}
	sellDetail.Detail = details
	if result := tx.Create(&sellDetail); result.RowsAffected == 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "保存库存扣减历史失败")
	}
	// 尝试提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "事务提交失败: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (i *InventoryServer) Reback(ctx context.Context, in *proto.SellInfo) (*emptypb.Empty, error) {
	// 库存归还 1、订单超时 2、订单创建失败取消
	tx := global.DB.Begin()
	for _, goodInfo := range in.GoodsInfo {
		var inv model.Inventory
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("goods = ?", goodInfo.GoodsId).First(&inv); result.RowsAffected == 0 {
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

func AutoReback(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	type OrderInfo struct {
		OrderSn string
	}
	for i := range msgs {

		// 新建一张表， 这张表记录了详细的订单扣减细节，以及归还细节
		var orderInfo OrderInfo
		err := json.Unmarshal(msgs[i].Body, &orderInfo)
		if err != nil {
			zap.S().Errorf("解析json失败： %v\n", msgs[i].Body)
			return consumer.ConsumeSuccess, nil
		}

		//去将inv的库存加回去 将selldetail的status设置为2 要在事务中进行
		tx := global.DB.Begin()
		var sellDetail model.StockSellDetail
		if result := tx.Model(&model.StockSellDetail{}).Where(&model.StockSellDetail{OrderSn: orderInfo.OrderSn, Status: 1}).First(&sellDetail); result.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}
		//那么逐个归还库存
		for _, orderGood := range sellDetail.Detail {
			// update语句的 update xx set stocks=stocks+2
			if result := tx.Model(&model.Inventory{}).Where(&model.Inventory{Goods: orderGood.Goods}).Update("stocks", gorm.Expr("stocks+?", orderGood.Num)); result.RowsAffected == 0 {
				tx.Rollback()
				return consumer.ConsumeRetryLater, nil
			}
		}

		if result := tx.Model(&model.StockSellDetail{}).Where(&model.StockSellDetail{OrderSn: orderInfo.OrderSn}).Update("status", 2); result.RowsAffected == 0 {
			tx.Rollback()
			return consumer.ConsumeRetryLater, nil
		}
		tx.Commit()
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}
func (i *InventoryServer) mustEmbedUnimplementedInventoryServer() {

}
