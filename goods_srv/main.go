package main

import (
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
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
	"syscall"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50051, "端口号")
	// 初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	flag.Parse()
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Infof("ip:%s,port:%d", *IP, *Port)

	// 服务注册
	cfg := api.DefaultConfig()
	consulInfo := global.ServerConfig.ConsulInfo
	cfg.Address = fmt.Sprintf("%s:%d", consulInfo.Host, consulInfo.Port)
	var client *api.Client
	var err error
	client, err = api.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	// 生成对应检查对象
	check := &api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("host.docker.internal:%d", *Port),
		Timeout:                        "1s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "1m",
	}
	// 生成注册对象
	serviceID := fmt.Sprintf("%s", uuid.New())
	registration := &api.AgentServiceRegistration{
		ID:   serviceID,
		Name: global.ServerConfig.Name,
		Port: *Port,
		Tags: global.ServerConfig.Tags,
		//Address: "host.docker.internal",
		Address: global.ServerConfig.Host,
		Check:   check,
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}

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

	// 接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err := client.Agent().ServiceDeregister(serviceID); err != nil {
		zap.S().Errorw("注销失败", "err", err.Error())
	}
}
