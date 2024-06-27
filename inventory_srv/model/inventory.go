package model

//	type Stock struct {
//		BaseModel
//		Name string `gorm:"type:varchar(20);not null" json:"name"`
//		Address string `gorm:"type:varchar(200);not null;default:''" json:"address"`
//	}
type Inventory struct {
	BaseModel
	Goods   int32 `gorm:"type:int;index;not null;default:0" json:"goods"`
	Stocks  int32 `gorm:"type:int;not null;default:0" json:"stocks"`  // 库存数
	Version int32 `gorm:"type:int;not null;default:0" json:"version"` // 乐观锁
}

//type InventoryHistory struct {
//	user int32
//	goods int32
//	nums int32
//	order int32
//	status int32 //1:库存预扣 2:支付成功
//}
