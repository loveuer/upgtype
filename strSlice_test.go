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
		su  = &User{ID: 10}

		originData = &User{Names: []string{
			"123",
			`aaa"bbb`,
			`aa""bn`,
			`cc\"dd`,
			`eee'ff`,
			`ggg''hh`,
			`iii\'jj`,
			`ii\\\jj`,
			`{dadkaj}`,
			`dbdb}`,
			"{hdjad",
			`\,"'}{`,
		}}
	)

	initDB()

	if err = db.Model(&User{}).Where("id = ?", su.ID).First(su).Error; err != nil {
		t.Errorf("gorm first err: %v", err)
	}

	log.Printf("scan user: %+v", *su)

	compairOriAndGot(originData.Names, su.Names)
}

func TestValue(t *testing.T) {
	var (
		err error
	)

	initDB()

	nu := &User{Names: []string{
		"123",
		`aaa"bbb`,
		`aa""bn`,
		`cc\"dd`,
		`eee'ff`,
		`ggg''hh`,
		`iii\'jj`,
		`ii\\\jj`,
		`{dadkaj}`,
		`dbdb}`,
		"{hdjad",
		`\,"'}{`,
	}}

	if err = db.Create(nu).Error; err != nil {
		t.Errorf("gorm create err: %v", nu)
	}
}

func compairOriAndGot(ori, got []string) {
	var (
		length = len(ori)
	)

	if len(ori) != len(got) {
		log.Printf("x => 长度不一致")
		if len(ori) < len(got) {
			length = len(ori)
		} else {
			length = len(got)
		}
	}

	for idx := 0; idx < length; idx++ {
		log.Printf("O => 原始: %s, 现在: %s", ori[idx], got[idx])
	}
}
