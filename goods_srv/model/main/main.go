package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"log"
	"math/rand"
	"os"
	"shop_srvs/goods_srv/model"
	"time"
)

func genMd5(code string) string {
	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum([]byte{}))
}
func main() {
	dsn := "root:123456@tcp(127.0.0.1:3307)/shop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// 全局模式
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&model.Category{}, &model.Brands{}, &model.Goods{}, &model.GoodsCategoryBrand{}, &model.Banner{})
	insertBanners(db)
	insertCategories(db)
	insertBrands(db)
	insertGoodsCategoryBrands(db)
	insertGoods(db)

}

// 假设有一个初始化数据库和 logger 的函数
func insertCategories(db *gorm.DB) {
	for i := 1; i <= 100; i++ {
		category := model.Category{
			Name:             fmt.Sprintf("Category %d", i),
			ParentCategoryID: 0, // 假设这些分类没有父分类
			Level:            1,
			IsTab:            false,
		}
		if err := db.Create(&category).Error; err != nil {
			log.Printf("Error inserting category: %s", err)
		}
	}
}
func insertBrands(db *gorm.DB) {
	for i := 1; i <= 100; i++ {
		brand := model.Brands{
			Name: fmt.Sprintf("Brand %d", i),
			Logo: fmt.Sprintf("https://example.com/logos/logo%d.png", i),
		}
		if err := db.Create(&brand).Error; err != nil {
			log.Printf("Error inserting brand: %s", err)
		}
	}
}
func insertGoodsCategoryBrands(db *gorm.DB) {
	for i := 1; i <= 100; i++ {
		gcb := model.GoodsCategoryBrand{
			CategoryID: int32(i), // 假设 CategoryID 和 i 匹配
			BrandsID:   int32(i), // 假设 BrandsID 和 i 匹配
		}
		if err := db.Create(&gcb).Error; err != nil {
			log.Printf("Error inserting GoodsCategoryBrand: %s", err)
		}
	}
}
func insertBanners(db *gorm.DB) {
	for i := 1; i <= 100; i++ {
		banner := model.Banner{
			Image: fmt.Sprintf("https://example.com/banners/banner%d.png", i),
			Url:   fmt.Sprintf("https://example.com/products/%d", i),
			Index: int32(i),
		}
		if err := db.Create(&banner).Error; err != nil {
			log.Printf("Error inserting banner: %s", err)
		}
	}
}
func insertGoods(db *gorm.DB) {
	for i := 1; i <= 100; i++ {
		goods := model.Goods{
			CategoryID:      int32(rand.Intn(100) + 1), // 随机分配 CategoryID
			BrandsID:        int32(rand.Intn(100) + 1), // 随机分配 BrandsID
			OnSale:          true,
			ShipFree:        false,
			IsNew:           true,
			IsHot:           false,
			Name:            fmt.Sprintf("Product %d", i),
			GoodsSn:         fmt.Sprintf("SN%d", i),
			ClickNum:        int32(rand.Intn(1000)),
			FavNum:          int32(rand.Intn(500)),
			MarketPrice:     float32(100.0 + float64(i)),
			ShopPrice:       float32(80.0 + float64(i)),
			GoodsBrief:      fmt.Sprintf("Brief description of product %d", i),
			Images:          model.GormList{fmt.Sprintf("https://example.com/images/img%d.png", i)},
			DescImages:      model.GormList{fmt.Sprintf("https://example.com/images/desc%d.png", i)},
			GoodsFrontImage: fmt.Sprintf("https://example.com/images/front%d.png", i),
		}
		if err := db.Create(&goods).Error; err != nil {
			log.Printf("Error inserting goods: %s", err)
		}
	}
}
