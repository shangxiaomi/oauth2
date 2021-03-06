package common

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"oauth2/config"
	mylog "oauth2/log"
	"oauth2/model"
	"time"
)

var dB *gorm.DB

func InitDB() *gorm.DB {
	cfg := config.Get()
	databaseConfig := mysql.New(mysql.Config{
		DriverName: cfg.Db.Default.DriveName,
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
			cfg.Db.Default.User,
			cfg.Db.Default.Password,
			cfg.Db.Default.Host,
			cfg.Db.Default.Port,
			cfg.Db.Default.DbName),
	})
	newLogger := logger.New(
		log.New(mylog.GetLogFile(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
		},
	)
	db, err := gorm.Open(databaseConfig, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		//log.Println(err)
		panic(err)
	}
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		//log.Println("用户表创建失败")
		panic("自动创建用户表失败，请进行排查")
	}
	dB = db
	mylog.Info.Println("database init success")
	return dB
}

func GetDB() *gorm.DB {
	if dB == nil {
		//log.Println("DB没有进行初始化，请调用common.InitDB()函数进行初始化")
		panic("DB没有进行初始化，请调用common.InitDB()函数进行初始化")
	}
	return dB
}