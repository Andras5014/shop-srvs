package model

type Category struct {
	BaseModel
	Name             string      `gorm:"type:varchar(20);not null" json:"name"`
	ParentCategoryID int32       `gorm:"type:int;not null;default:0" json:"parent_category_id"`
	ParentCategory   *Category   `json:"-"`
	SubCategory      []*Category `gorm:"foreignKey:ParentCategoryID;references:ID" json:"sub_category"`
	Level            int32       `gorm:"type:int;not null;default:1" json:"level"`
	IsTab            bool        `gorm:"type:bool;not null;default:false" json:"is_tab"`
}

type Brands struct {
	BaseModel
	Name string `gorm:"type:varchar(20);not null" json:"name"`
	Logo string `gorm:"type:varchar(200);not null;default:''" json:"logo"`
}

type GoodsCategoryBrand struct {
	BaseModel
	CategoryID int32 `gorm:"type:int;index:idx_category_brand,unique;not null" json:"category_id"`
	Category   Category
	BrandsID   int32 `gorm:"type:int;index:idx_category_brand,unique;not null" json:"brands_id"`
	Brands     Brands
}

type Banner struct {
	BaseModel
	Image string `gorm:"type:varchar(200);not null" json:"image"`
	Url   string `gorm:"type:varchar(200);not null" json:"url"`
	Index int32  `gorm:"type:int;not null;default:1" json:"index"`
}

type Goods struct {
	BaseModel
	CategoryID int32 `gorm:"type:int;not null" json:"category_id"`
	Category   Category

	BrandsID int32 `gorm:"type:int;not null" json:"brands_id"`
	Brands   Brands

	OnSale          bool     `gorm:"type:bool;not null;default:false" json:"on_sale	"`
	ShipFree        bool     `gorm:"type:bool;not null;default:false" json:"ship_free"` //运费
	IsNew           bool     `gorm:"type:bool;not null;default:false" json:"is_new"`    //上新
	IsHot           bool     `gorm:"type:bool;not null;default:false" json:"is_hot"`    //热点
	Name            string   `gorm:"type:varchar(50);not null" json:"name"`
	GoodsSn         string   `gorm:"type:varchar(50);not null" json:"goods_sn"`    //订单编号
	ClickNum        int32    `gorm:"type:int;not null;default:0" json:"click_num"` //点击
	FavNum          int32    `gorm:"type:int;not null;default:0" json:"fav_num"`   //收藏
	MarketPrice     float32  `gorm:"type:decimal(10,2);not null;default:0" json:"market_price"`
	ShopPrice       float32  `gorm:"type:decimal(10,2);not null;default:0" json:"shop_price"`
	GoodsBrief      string   `gorm:"type:varchar(100);not null" json:"goods_brief"` //描述
	Images          GormList `gorm:"type:varchar(1000);not null;default:''" json:"images"`
	DescImages      GormList `gorm:"type:varchar(1000);not null;default:''" json:"desc_images"`
	GoodsFrontImage string   `gorm:"type:varchar(200);not null;default:''" json:"goods_front_image"`
}
