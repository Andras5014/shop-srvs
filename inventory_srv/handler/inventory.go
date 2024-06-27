package handler

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"shop_srvs/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

func (i InventoryServer) SetInv(ctx context.Context, in *proto.GoodsInvInfo, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (i InventoryServer) InvDetail(ctx context.Context, in *proto.GoodsInvInfo, opts ...grpc.CallOption) (*proto.GoodsInvInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (i InventoryServer) Sell(ctx context.Context, in *proto.SellInfo, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (i InventoryServer) Reback(ctx context.Context, in *proto.SellInfo, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}
