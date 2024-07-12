package main

import (
	"flag"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"os"
	"os/signal"
	"shop_srvs/order_srv/global"
	"shop_srvs/order_srv/handler"
	"shop_srvs/order_srv/initialize"
	"shop_srvs/order_srv/proto"
	"shop_srvs/order_srv/utils"
	"shop_srvs/order_srv/utils/otgrpc"
	"shop_srvs/order_srv/utils/register/consul"
	"syscall"
	"time"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 0, "端口号")
	// 初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.InitSrvsConn()
	flag.Parse()
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Infof("ip:%s,port:%d", *IP, *Port)

	// 初始化jaegercfg
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: fmt.Sprintf("%s:%d", global.ServerConfig.JaegerInfo.Host, global.ServerConfig.JaegerInfo.Port),
		},
		ServiceName: global.ServerConfig.JaegerInfo.Name,
	}
	tracer, closer, err := cfg.NewTracer(jaegercfg.Logger(jaeger.StdLogger))
	if err != nil {
		panic(err)
	}
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	server := grpc.NewServer(grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
	proto.RegisterOrderServer(server, &handler.OrderServer{})

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
	zap.S().Debugf("启动服务注册中心成功, 端口: %d", *Port)

	// 监听订单超时topic
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"127.0.0.1:9876"}),
		consumer.WithGroupName("shop-order"),
	)
	if err := c.Subscribe("order_timeout", consumer.MessageSelector{}, handler.OrderTimeout); err != nil {
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
