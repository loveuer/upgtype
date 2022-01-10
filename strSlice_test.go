package upgtype

import (
	"fmt"
	"log"
	"testing"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

var (
	db     *gorm.DB
	config = &DBConfig{}
)

func initDB() {
	var (
		err error
	)

	viper.SetConfigType("json")
	viper.SetConfigFile("./config.json")

	if err = viper.ReadInConfig(); err != nil {
		log.Fatalf("read config err: %v", err)
	}

	if err = viper.Unmarshal(config); err != nil {
		log.Fatalf("unmarshal config err: %v", err)
	}

	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			config.Host,
			config.User,
			config.Password,
			config.DBName,
			config.Port),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		log.Fatalf("init pg client err: %v", err)
	}

	db = db.Debug()
}

type User struct {
	ID    int      `json:"id" gorm:"column:id"`
	Names StrSlice `json:"names" gorm:"column:names;type:varchar(32)[]"`
}

func TestScan(t *testing.T) {
	var (
		err error
		su  = &User{ID: 8}
	)

	initDB()

	if err = db.Model(&User{}).Where("id = ?", su.ID).First(su).Error; err != nil {
		t.Errorf("gorm first err: %v", err)
	}

	log.Printf("scan user: %+v", *su)
}
