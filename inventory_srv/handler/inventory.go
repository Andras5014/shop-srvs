package handler

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
//	for _, goodInfo := range in.GoodsInfo {
//		var inv model.Inventory
//		for {
//			if result := tx.Where("goods = ?", goodInfo.GoodsId).First(&inv); result.RowsAffected == 0 {
//				tx.Rollback()
//				return nil, status.Errorf(codes.NotFound, "库存信息不存在")
//			}
//			if inv.Stocks < goodInfo.Num {
//
//				tx.Rollback()
//				return nil, status.Errorf(codes.OutOfRange, "库存不足")
//			}
//
//			inv.Stocks -= goodInfo.Num
//			if result := tx.Model(&model.Inventory{}).Select("stocks", "version").
//				Where("goods = ? and version = ?", goodInfo.GoodsId, inv.Version).
//				Updates(&model.Inventory{Stocks: inv.Stocks, Version: inv.Version + 1}); result.RowsAffected == 0 {
//				zap.S().Infof("更新失败")
//			} else {
//				break
//			}
//
//		}
//		tx.Save(&inv)
//	}
//
//	if err := tx.Commit().Error; err != nil {
//		tx.Rollback()
//		return nil, status.Errorf(codes.Internal, "事务提交失败")
//	}
//
//	return &emptypb.Empty{}, nil
//}

func (*InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存， 本地事务 [1:10,  2:5, 3: 20]
	//数据库基本的一个应用场景：数据库事务
	//并发情况之下 可能会出现超卖 1
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)

	tx := global.DB.Begin()
	//m.Lock() //获取锁 这把锁有问题吗？  假设有10w的并发， 这里并不是请求的同一件商品  这个锁就没有问题了吗？

	//这个时候应该先查询表，然后确定这个订单是否已经扣减过库存了，已经扣减过了就别扣减了
	//并发时候会有漏洞， 同一个时刻发送了重复了多次， 使用锁，分布式锁
	//sellDetail := model.StockSellDetail{
	//	OrderSn: req.OrderSn,
	//	Status:  1,
	//}
	//var details []model.GoodsDetail
	for _, goodInfo := range req.GoodsInfo {
		//details = append(details, model.GoodsDetail{
		//	Goods: goodInfo.GoodsId,
		//	Num: goodInfo.Num,
		//})

		var inv model.Inventory
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods:goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//  tx.Rollback() //回滚之前的操作
		//  return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		//}

		//for {
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		inv.Stocks -= goodInfo.Num
		tx.Save(&inv)

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
		//update order set stocks = stocks-1, version=version+1 where goods=goods and version=version
		//这种写法有瑕疵，为什么？
		//零值 对于int类型来说 默认值是0 这种会被gorm给忽略掉
		//if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version= ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version+1}); result.RowsAffected == 0 {
		//  zap.S().Info("库存扣减失败")
		//}else{
		//  break
		//}
		//}
		//tx.Save(&inv)
	}
	//sellDetail.Detail = details
	////写selldetail表
	//if result := tx.Create(&sellDetail); result.RowsAffected == 0 {
	//	tx.Rollback()
	//	return nil, status.Errorf(codes.Internal, "保存库存扣减历史失败")
	//}
	tx.Commit() // 需要自己手动提交操作
	//m.Unlock() //释放锁
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
func (i *InventoryServer) mustEmbedUnimplementedInventoryServer() {

}
