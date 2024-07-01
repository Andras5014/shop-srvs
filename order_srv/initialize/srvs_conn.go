package initialize

import (
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"shop_srvs/order_srv/global"
	"shop_srvs/order_srv/proto"
)

func InitSrvsConn() {
	// 初始化第三方服务连接
	consulInfo := global.ServerConfig.ConsulInfo

	goodsConn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", consulInfo.Host, consulInfo.Port, global.ServerConfig.GoodsServerInfo.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Errorw("[InitSrvConn] 连接Goods服务失败", "msg", err.Error())
	}
	inventoryConn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", consulInfo.Host, consulInfo.Port, global.ServerConfig.InventoryServerInfo.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Errorw("[InitSrvConn] 连接Inventory服务失败", "msg", err.Error())
	}

	goodsSrvClient := proto.NewGoodsClient(goodsConn)
	inventoryClient := proto.NewInventoryClient(inventoryConn)
	global.GoodsSrvClient = goodsSrvClient
	global.InventorySrvClient = inventoryClient
}
