package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"shop_srvs/user_srv/proto"
)

var userClient proto.UserClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())

	if err != nil {
		panic(err)
	}
	userClient = proto.NewUserClient(conn)

}
func TestGetUserList() {

	res, err := userClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    1,
		PSize: 2,
	})
	if err != nil {
		panic(err)
	}
	for _, user := range res.Data {
		fmt.Println(user)
		checkResp, err := userClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
			Password:          "admin123",
			EncryptedPassword: user.Password,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(checkResp.Success)
	}
}
func TestCreateUser() {
	user := &proto.CreateUserInfo{
		NickName: "andras",
		Password: "admin111",
		Mobile:   "12345678901",
	}
	res, err := userClient.CreateUser(context.Background(), user)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
func main() {
	Init()
	TestGetUserList()
	defer conn.Close()
}
