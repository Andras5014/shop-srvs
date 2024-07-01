package global

import (
	"github.com/go-redsync/redsync/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"shop_srvs/order_srv/proto"

	"shop_srvs/order_srv/config"
	"time"
)

var (
	DB           *gorm.DB
	Redsync      *redsync.Redsync
	ServerConfig *config.ServerConfig
	NacosConfig  *config.NacosConfig

	GoodsSrvClient     proto.GoodsClient
	InventorySrvClient proto.InventoryClient
)

func init() {
	dsn := "root:123456@tcp(127.0.0.1:3307)/shop_order_srv?charset=utf8mb4&parseTime=True&loc=Local"
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
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

}
