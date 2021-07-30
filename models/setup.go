package models

import (
	"fmt"
	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"net/url"
	"os"
	"time"
)

var DB *gorm.DB
var DBAutoMigrate []interface{}

func ConnectDatabase() {
	// set default parameters.
	cfg, err := config.LoadConf()
	if err != nil {
		logx.LogError.Error("DB ERR Load yaml config file error: '%v'", err)
		return
	}

	//SQL string connection
	sqlConn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.Charset)

	timezone := cfg.DB.Timezone
	if timezone != "" {
		sqlConn = sqlConn + "&loc=" + url.QueryEscape(timezone)
	}

	defer logx.LogAccess.Info("DB Init Connection : ", sqlConn)

	//GORM Logger config
	dbLogger := gormLogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      gormLogger.Warn,
		Colorful:      true,
	})
	if cfg.Core.Mode == "debug" {
		dbLogger = dbLogger.LogMode(gormLogger.Info)
	}

	//GORM Connection
	sqlDB, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       sqlConn, // data source name
		DefaultStringSize:         256,     // default size for string fields
		DisableDatetimePrecision:  true,    // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,    // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,    // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,   // auto configure based on currently MySQL version
	}), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		logx.LogError.Error(err.Error())
		panic("DB ERR failed to connect database")
	}

	// Migrate the schema
	if err := sqlDB.AutoMigrate(DBAutoMigrate...); err != nil {
		logx.LogError.Error(err.Error())
		panic("DB ERR failed to migrate database")
	} else {
		defer logx.LogAccess.Info("DB migration completed")
	}

	// SQL set debug mode
	if cfg.Core.Mode == "debug" {
		sqlDB = sqlDB.Debug()
		defer logx.LogAccess.Info("DB Debug mode enabled")
	}

	DB = sqlDB
}
