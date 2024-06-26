package handler

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"shop_srvs/user_srv/global"
	"shop_srvs/user_srv/model"
	"shop_srvs/user_srv/proto"
	"strings"
	"time"
)

type UserServer struct {
	proto.UnimplementedUserServer
}

func ModelToResponse(user model.User) *proto.UserInfoResponse {
	//
	userInfoResponse := &proto.UserInfoResponse{
		Id:       user.ID,
		NickName: user.NickName,
		Mobile:   user.Mobile,
		Password: user.Password,
		Role:     int32(user.Role),
	}
	if user.Birthday != nil {
		userInfoResponse.Birthday = user.Birthday.Unix()
	}
	return userInfoResponse
}
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
func (u *UserServer) GetUserList(ctx context.Context, info *proto.PageInfo) (*proto.UserListResponse, error) {
	// 获取用户列表
	var users []model.User
	result := global.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	fmt.Println("users:")
	res := &proto.UserListResponse{}
	res.Total = int32(result.RowsAffected)

	global.DB.Scopes(Paginate(int(info.Pn), int(info.PSize))).Find(&users)

	for _, user := range users {
		userInfoRes := ModelToResponse(user)
		res.Data = append(res.Data, userInfoRes)
	}
	return res, nil
}

func (u *UserServer) GetUserByMobile(ctx context.Context, request *proto.MobileRequest) (*proto.UserInfoResponse, error) {

	var user model.User
	result := global.DB.Where(&model.User{Mobile: request.Mobile}).First(&user)

	if result.RowsAffected == 0 {
		// 如果没有找到任何结果，返回 nil 表示没有用户
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	if result.Error != nil {
		// 如果发生了错误，返回错误
		return nil, result.Error
	}
	return ModelToResponse(user), nil
}

func (u *UserServer) GetUserById(ctx context.Context, request *proto.IdRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	result := global.DB.First(&user, request.Id)

	if result.RowsAffected == 0 {
		// 如果没有找到任何结果，返回 nil 表示没有用户
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	if result.Error != nil {
		// 如果发生了错误，返回错误
		return nil, result.Error
	}
	return ModelToResponse(user), nil
}

func (u *UserServer) CreateUser(ctx context.Context, info *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	// 新建用户
	var user model.User
	result := global.DB.Where(&model.User{Mobile: info.Mobile}).First(&user)
	if result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "user already exists")
	}
	user = model.User{
		NickName: info.NickName,
		Mobile:   info.Mobile,
	}
	options := &password.Options{
		SaltLen:      16,
		Iterations:   1000,
		KeyLen:       32,
		HashFunction: sha512.New,
	}
	salt, encodedPwd := password.Encode(info.Password, options)
	user.Password = fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)

	result = global.DB.Create(&user)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "create user failed")
	}

	return ModelToResponse(user), nil

}
func timePtr(t time.Time) *time.Time {
	return &t
}
func (u *UserServer) UpdateUser(ctx context.Context, info *proto.UpdateUserInfo) (*emptypb.Empty, error) {

	var user model.User
	result := global.DB.First(&user, info.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	user.NickName = info.NickName
	user.Birthday = timePtr(time.Unix(info.Birthday, 0))
	user.Gender = info.Gender
	result = global.DB.Save(&user)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "update user failed")
	}
	return &emptypb.Empty{}, nil
}

func (u *UserServer) CheckPassword(ctx context.Context, info *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {

	// 校验密码
	options := &password.Options{
		SaltLen:      16,
		Iterations:   1000,
		KeyLen:       32,
		HashFunction: sha512.New,
	}
	passwordInfo := strings.Split(info.EncryptedPassword, "$")
	check := password.Verify(info.Password, passwordInfo[2], passwordInfo[3], options)
	return &proto.CheckResponse{Success: check}, nil
}
