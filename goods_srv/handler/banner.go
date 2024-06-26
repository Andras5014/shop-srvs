package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"shop_srvs/goods_srv/global"
	"shop_srvs/goods_srv/model"
	"shop_srvs/goods_srv/proto"
)

func (g *GoodsServer) BannerList(ctx context.Context, empty *emptypb.Empty) (*proto.BannerListResponse, error) {
	bannerListResponse := &proto.BannerListResponse{}
	var banners []model.Banner
	result := global.DB.Find(&banners)
	if result.Error != nil {
		return nil, result.Error
	}
	bannerListResponse.Total = int32(result.RowsAffected)
	var bannerResponses []*proto.BannerResponse
	for _, banner := range banners {
		bannerResponses = append(bannerResponses, &proto.BannerResponse{
			Id:    banner.ID,
			Image: banner.Image,
			Url:   banner.Url,
			Index: banner.Index,
		})
	}
	bannerListResponse.Data = bannerResponses
	return bannerListResponse, nil

}

func (g *GoodsServer) CreateBanner(ctx context.Context, request *proto.BannerRequest) (*proto.BannerResponse, error) {
	banner := model.Banner{
		Image: request.Image,
		Url:   request.Url,
		Index: request.Index,
	}
	global.DB.Save(&banner)
	return &proto.BannerResponse{
		Id: banner.ID,
	}, nil
}

func (g *GoodsServer) DeleteBanner(ctx context.Context, request *proto.BannerRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Banner{}, request.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "Banner Not Found")
	}
	return &emptypb.Empty{}, nil
}

func (g *GoodsServer) UpdateBanner(ctx context.Context, request *proto.BannerRequest) (*emptypb.Empty, error) {

	var banner model.Banner
	if result := global.DB.First(&banner, request.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "Banner Not Found")
	}
	if request.Image != "" {
		banner.Image = request.Image
	}
	if request.Url != "" {
		banner.Url = request.Url
	}
	if request.Index != 0 {
		banner.Index = request.Index
	}
	global.DB.Save(&banner)
	return &emptypb.Empty{}, nil
}
