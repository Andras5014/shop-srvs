package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"google.golang.org/grpc"

	grpc_health_v1_impl "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	_ "gorm.io/driver/mysql"
	"io"
	"log"
	"net"
)

func genMd5(code string) string {
	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum([]byte{}))
}
func main() {
	//dsn := "root:123456@tcp(127.0.0.1:3307)/shop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"
	//newLogger := logger.New(
	//	log.New(os.Stdout, "\r\n", log.LstdFlags),
	//	logger.Config{
	//		SlowThreshold:             200 * time.Millisecond,
	//		LogLevel:                  logger.Info,
	//		IgnoreRecordNotFoundError: true,
	//		Colorful:                  true,
	//	},
	//)
	//
	//// 全局模式
	//db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	//	Logger: newLogger,
	//})
	//if err != nil {
	//	panic("failed to connect database")
	//}
	//db.AutoMigrate(&model.User{})
	//options := &password.Options{
	//	SaltLen:      16,
	//	Iterations:   1000,
	//	KeyLen:       32,
	//	HashFunction: sha512.New,
	//}
	//salt, encodedPwd := password.Encode("admin123", options)
	//newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	//fmt.Println(newPassword)
	//
	//for i := 0; i < 10; i++ {
	//	user := model.User{
	//		NickName: "user" + fmt.Sprintf("%d", i),
	//		Mobile:   "1380000000" + fmt.Sprintf("%d", i),
	//		Password: newPassword,
	//	}
	//	global.DB.Save(&user)
	//}

	IP := "0.0.0.0"
	Port := 50051

	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, grpc_health_v1_impl.NewServer())

	address := fmt.Sprintf("%s:%d", IP, Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	fmt.Printf("gRPC server is listening on %s\n", address)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
