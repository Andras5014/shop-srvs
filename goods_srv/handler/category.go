package handler

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"shop_srvs/goods_srv/global"
	"shop_srvs/goods_srv/model"
	"shop_srvs/goods_srv/proto"
)

func (g *GoodsServer) GetAllCategorysList(ctx context.Context, empty *emptypb.Empty) (*proto.CategoryListResponse, error) {

	var categorys []model.Category
	global.DB.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categorys)
	b, _ := json.Marshal(&categorys)
	return &proto.CategoryListResponse{
		Total:    int32(len(categorys)),
		JsonData: string(b),
	}, nil

}

func (g *GoodsServer) GetSubCategory(ctx context.Context, request *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	categoryListResponse := &proto.SubCategoryListResponse{}

	var category model.Category
	if result := global.DB.First(&category, request.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}
	categoryListResponse.Info = &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		Level:          category.Level,
		IsTab:          category.IsTab,
		ParentCategory: category.ParentCategoryID,
	}
	var subCategorys []model.Category
	//preloads := "SubCategory"
	//if category.Level == 1 {
	//	preloads = "SubCategory.SubCategory"
	//}
	global.DB.Where(&model.Category{ParentCategoryID: request.Id}).Find(&subCategorys)
	for _, subCategory := range subCategorys {
		categoryListResponse.SubCategorys = append(categoryListResponse.SubCategorys, &proto.CategoryInfoResponse{
			Id:             subCategory.ID,
			Name:           subCategory.Name,
			Level:          subCategory.Level,
			IsTab:          subCategory.IsTab,
			ParentCategory: subCategory.ParentCategoryID,
		})
	}
	return categoryListResponse, nil
}

func (g *GoodsServer) CreateCategory(ctx context.Context, request *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	category := model.Category{}
	cMap := map[string]interface{}{
		"name":   request.Name,
		"level":  request.Level,
		"is_tab": request.IsTab,
	}
	if request.Level != 1 {
		cMap["parent_category_id"] = request.ParentCategory
	}
	global.DB.Model(&model.Category{}).Create(cMap)
	return &proto.CategoryInfoResponse{Id: category.ID}, nil

}

func (g *GoodsServer) DeleteCategory(ctx context.Context, request *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Category{}, request.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}
	return &emptypb.Empty{}, nil
}

func (g *GoodsServer) UpdateCategory(ctx context.Context, request *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	var category model.Category
	if result := global.DB.First(&category, request.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}
	if request.Name != "" {
		category.Name = request.Name
	}

	category.IsTab = request.IsTab

	cMap := map[string]interface{}{
		"name":   category.Name,
		"is_tab": category.IsTab,
	}
	res := global.DB.Model(&model.Category{}).Where("id = ?", request.Id).Updates(cMap)
	if res.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新失败")
	}
	return &emptypb.Empty{}, nil
}
