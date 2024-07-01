package model

import (
	"database/sql/driver"
	"encoding/json"
)

//	type Stock struct {
//		BaseModel
//		Name string `gorm:"type:varchar(20);not null" json:"name"`
//		Address string `gorm:"type:varchar(200);not null;default:''" json:"address"`
//	}
//type Inventory struct {
//	BaseModel
//	Goods   int32 `gorm:"type:int;index;not null;default:0" json:"goods"`
//	Stocks  int32 `gorm:"type:int;not null;default:0" json:"stocks"`  // 库存数
//	Version int32 `gorm:"type:int;not null;default:0" json:"version"` // 乐观锁
//}

//type InventoryHistory struct {
//	user int32
//	goods int32
//	nums int32
//	order int32
//	status int32 //1:库存预扣 2:支付成功
//}

//type Stock struct {
//  BaseModel
//  Name string
//  Address string
//}

type GoodsDetail struct {
	Goods int32
	Num   int32
}
type GoodsDetailList []GoodsDetail

func (g GoodsDetailList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (g *GoodsDetailList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)
}

type Inventory struct {
	BaseModel
	Goods   int32 `gorm:"type:int;index"`
	Stocks  int32 `gorm:"type:int"`
	Version int32 `gorm:"type:int"` //分布式锁的乐观锁
}

type InventoryNew struct {
	BaseModel
	Goods   int32 `gorm:"type:int;index"`
	Stocks  int32 `gorm:"type:int"`
	Version int32 `gorm:"type:int"` //分布式锁的乐观锁
	Freeze  int32 `gorm:"type:int"` //冻结库存
}

type Delivery struct {
	Goods   int32  `gorm:"type:int;index"`
	Nums    int32  `gorm:"type:int"`
	OrderSn string `gorm:"type:varchar(200)"`
	Status  string `gorm:"type:varchar(200)"` //1. 表示等待支付 2. 表示支付成功 3. 失败
}

type StockSellDetail struct {
	OrderSn string          `gorm:"type:varchar(200);index:idx_order_sn,unique;"`
	Status  int32           `gorm:"type:varchar(200)"` //1 表示已扣减 2. 表示已归还
	Detail  GoodsDetailList `gorm:"type:varchar(200)"`
}

func (StockSellDetail) TableName() string {
	return "stockselldetail"
}
