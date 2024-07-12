package main

import (
	"flag"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"os"
	"os/signal"
	"shop_srvs/goods_srv/global"
	"shop_srvs/goods_srv/handler"
	"shop_srvs/goods_srv/initialize"
	"shop_srvs/goods_srv/proto"
	"shop_srvs/goods_srv/utils"
	"shop_srvs/goods_srv/utils/register/consul"
	"syscall"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 0, "端口号")
	// 初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.InitEs()
	flag.Parse()
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Infof("ip:%s,port:%d", *IP, *Port)

	zap.S().Debugf("启动服务注册中心成功, 端口: %d", *Port)

	server := grpc.NewServer()
	proto.RegisterGoodsServer(server, &handler.GoodsServer{})

	// 注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic(err)
	}

	// 启动 gRPC 服务器
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	//服务注册
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.New())
	err = register_client.Register(global.ServerConfig.Name, serviceId, global.ServerConfig.Host, *Port,
		global.ServerConfig.Tags)
	if err != nil {
		zap.S().Panic("服务注册失败:", err.Error())
	}

	//接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Info("注销失败:", err.Error())
	} else {
		zap.S().Info("注销成功:")
	}
}
