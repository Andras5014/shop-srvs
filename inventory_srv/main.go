package main

import (
	"flag"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"os"
	"os/signal"
	"shop_srvs/inventory_srv/global"
	"shop_srvs/inventory_srv/handler"
	"shop_srvs/inventory_srv/initialize"
	"shop_srvs/inventory_srv/proto"
	"shop_srvs/inventory_srv/utils"
	"shop_srvs/inventory_srv/utils/register/consul"
	"syscall"
	"time"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50051, "端口号")
	// 初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.InitRedSync()
	flag.Parse()
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Infof("ip:%s,port:%d", *IP, *Port)

	server := grpc.NewServer()
	proto.RegisterInventoryServer(server, &handler.InventoryServer{})

	// 注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic(err)
	}
	//
	//go testRedSync()
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
	zap.S().Debugf("启动服务注册中心成功, 端口: %d", *Port)

	//// 监听库存归还topic
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"127.0.0.1"}),
		consumer.WithGroupName("shop-inventory"),
	)
	if err := c.Subscribe("order_reback", consumer.MessageSelector{}, handler.AutoReback); err != nil {
		fmt.Println("读取消息失败", err)
	}
	_ = c.Start()

	//接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = c.Shutdown()
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Info("注销失败:", err.Error())
	} else {
		zap.S().Info("注销成功:")
	}
}

func testRedSync() {
	log.Println("enter testRed")
	mu := global.Redsync.NewMutex("test2")
	if err := mu.Lock(); err != nil {
		log.Println("lock test2 failed", err)
	} else {
		log.Println("lock test2 success")
		defer mu.Unlock()
	}

	time.Sleep(time.Second)

	mu1 := global.Redsync.NewMutex("test2")
	if err := mu1.Lock(); err != nil {
		log.Println("relock test2 failed", err)
	} else {
		log.Println("relock test2 success")
		defer mu1.Unlock()
	}
	log.Println("end testRed")
}
