package database

import (
	"fmt"
	"log"
	"os"
	_ "github.com/leebrouse/ems/backend/common/config"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// initPostgres 从 viper 配置中初始化 PostgreSQL 数据库并自动迁移模型
// configPrefix 指定在 viper 中的配置前缀，例如 "services.logistics.postgres"
func initPostgres(configPrefix string, models ...any) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString(configPrefix+".host"),
		viper.GetString(configPrefix+".port"),
		viper.GetString(configPrefix+".user"),
		viper.GetString(configPrefix+".password"),
		viper.GetString(configPrefix+".database"),
		viper.GetString(configPrefix+".sslmode"), // 推荐：disable / require
	)

	log.Println("PostgreSQL DSN:", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL database at %s: %v\n", configPrefix, err)
		return nil, err
	}

	if len(models) > 0 {
		err = db.AutoMigrate(models...)
		if err != nil {
			log.Printf("Failed to auto-migrate for %s: %v\n", configPrefix, err)
			return nil, err
		}
	}

	return db, nil
}

// MustInitPostgres 初始化 PostgreSQL，如果失败则直接退出程序
func Connect(configPrefix string, models ...interface{}) *gorm.DB {
	db, err := initPostgres(configPrefix, models...)
	if err != nil {
		os.Exit(1)
	}
	return db
}
