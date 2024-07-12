package global

import (
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
	"shop_srvs/goods_srv/config"
)

var (
	DB           *gorm.DB
	ServerConfig *config.ServerConfig
	NacosConfig  *config.NacosConfig
	EsClient     *elastic.Client
)

//func init() {
//	dsn := "root:123456@tcp(127.0.0.1:3307)/shop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
//	newLogger := logger.New(
//		log.New(os.Stdout, "\r\n", log.LstdFlags),
//		logger.Config{
//			SlowThreshold:             200 * time.Millisecond,
//			LogLevel:                  logger.Info,
//			IgnoreRecordNotFoundError: true,
//			Colorful:                  true,
//		},
//	)
//
//	// 全局模式
//	var err error
//	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//		Logger: newLogger,
//	})
//	if err != nil {
//		panic("failed to connect database")
//	}
//
//}
