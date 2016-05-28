package usth

import (
	"database/sql"
	"fmt"
	"github.com/EyciaZhou/configparser"
	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/EyciaZhou/msghub-http/M/HeadStorer"
)

type Error struct {
	_time string
	_err  string
}

func (m *Error) Error() string {
	return m._time + " : " + m._err
}

func newError(_time string, _err string) *Error {
	return &Error{_time, _err}
}

func newErrorByError(_time string, _err error) *Error {
	return &Error{_time, _err.Error()}
}

type Config struct {
	DBAddress  string `default:"127.0.0.1"`
	DBPort     string `default:"3306"`
	DBName     string `default:"usth"`
	DBUsername string `default:"root"`
	DBPassword string `default:"fmttm233"`

	QiniuAccessKey string`default:"fake"`
	QiniuSecretKey string`default:"fake"`
	QiniuBucket string `default:"usth-head"`
	QiniuDownloadUrl string `default:"http://o7rtp39nn.bkt.clouddn.com/"`
	QiniuCallbackUrl string `default:"https://usth.eycia.me/head/callback"`
}

var config Config
var (
	db *sql.DB
	HeadStore HeadStorer.HeadStorer
)

func init() {
	configparser.AutoLoadConfig("M.reply", &config)

	var err error
	log.Info("M.msghub Start Connect mysql")
	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?collation=utf8mb4_general_ci", config.DBUsername, config.DBPassword, config.DBAddress, config.DBPort, config.DBName)
	db, err = sql.Open("mysql", url)
	if err != nil {
		log.Panic("M.msghub Can't Connect DB REASON : " + err.Error())
		return
	}
	err = db.Ping()
	if err != nil {
		log.Panic("M.msghub Can't Connect DB REASON : " + err.Error())
		return
	}
	log.Info("M.msghub connected")

	HeadStore = HeadStorer.NewQiniuHeadStorer(&HeadStorer.QiniuHeadStorerConfig{
		config.QiniuAccessKey,
		config.QiniuSecretKey,
		config.QiniuBucket,
		config.QiniuDownloadUrl,
		config.QiniuCallbackUrl,
	}, DBInfo)
}
