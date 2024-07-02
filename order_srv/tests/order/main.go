package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"shop_srvs/order_srv/proto"
)

var orderClient proto.OrderClient
var conn *grpc.ClientConn

func TestCreateCartItem(userId int32, goodsId int32, nums int32) {
	rsp, err := orderClient.CreateCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  userId,
		GoodsId: goodsId,
		Nums:    nums,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("CreateCartItem成功", rsp)
}
func TestCartItemList(id int32) {
	rsp, err := orderClient.CartItemList(context.Background(), &proto.UserInfo{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	for _, item := range rsp.Data {
		fmt.Println(item)
	}
}
func TestUpdateCartItem(id int32) {
	_, err := orderClient.UpdateCartItem(context.Background(), &proto.CartItemRequest{
		Id:      id,
		Checked: true,
	})
	if err != nil {
		panic(err)
	}
}

func TestCreateOrder() {
	_, err := orderClient.CreateOrder(context.Background(), &proto.OrderRequest{
		UserId:  4,
		Address: "四川省成都市xxxx",
		Name:    "andras",
		Mobile:  "17628311111",
		Post:    "请尽快发货",
	})
	if err != nil {
		panic(err)
	}
}

func TestGetOrderDetail(orderId int32) {
	rsp, err := orderClient.OrderDetail(context.Background(), &proto.OrderRequest{
		Id: orderId,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.OrderInfo.OrderSn)
	for _, good := range rsp.Goods {
		fmt.Println(good.GoodsName)
	}

}

func TestOrderList() {
	rsp, err := orderClient.OrderList(context.Background(), &proto.OrderFilterRequest{
		UserId: 1,
	})
	if err != nil {
		panic(err)
	}

	for _, order := range rsp.Data {
		fmt.Println(order.OrderSn)
	}
}
func Init() {
	var err error
	conn, err = grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	orderClient = proto.NewOrderClient(conn)
}

func main() {
	Init()
	//TestCreateCartItem(1, 421, 1)
	//TestCartItemList(1)
	TestCreateOrder()
	//TestGetOrderDetail(10)
	conn.Close()

}
