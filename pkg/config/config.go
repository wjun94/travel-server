// Package config 负责从环境变量加载所有配置项
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config 应用配置结构体
type Config struct {
	// MySQL 配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// 服务配置
	ServerPort         string // 监听端口
	SnowflakeMachineID int64  // 雪花算法机器 ID
	Env                string // 运行环境：development / production
	APIVersionPrefix   string // API 版本前缀（如 /api/v1、/api/testv1）

	// Redis 配置
	REDIS_HOST     string
	REDIS_PORT     string
	REDIS_PASSWORD string
	REDIS_DB       string

	// 七牛云配置（可选）
	QiniuAccessKey     string
	QiniuSecretKey     string
	QiniuBucket        string
	QiniuDomain        string
	QiniuZone          string
	UploadPrefix       string
	UploadPrefixAdmin  string
	UploadPrefixQrcode string

	// 微信小程序配置
	AppId     string
	AppSecret string

	// 微信支付配置（可选）
	WechatPayMchID          string
	WechatPayApiV3Key       string
	WechatPaySerialNo       string
	WechatPayPrivateKeyPath string
	WechatPayNotifyURL      string
	PublicKeyPath           string
	PublicKeySerialNo       string

	// 第三方服务 Key
	AmapKey        string // 高德地图 Web API Key
	QWeatherKey    string // 和风天气PEM
	QWeatherKID    string // 和风天气凭证ID
	QWeatherPID    string // 和风天气项目ID
	DeepSeekApiKey string // DeepSeek API Key
}

// AppConfig 全局配置实例，在 main 启动时初始化
var AppConfig *Config

// LoadConfig 从环境变量加载配置，未设置的变量使用默认值
func LoadConfig() {
	AppConfig = &Config{
		DBHost:           getEnv("DB_HOST", "127.0.0.1"),
		DBPort:           getEnv("DB_PORT", "3306"),
		DBUser:           getEnv("DB_USER", "root"),
		DBPassword:       getEnv("DB_PASSWORD", "123456"),
		DBName:           getEnv("DB_NAME", "travel"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		APIVersionPrefix: getEnv("API_VERSION_PREFIX", "/api/v1"),

		REDIS_HOST:     getEnv("REDIS_HOST", "127.0.0.1"),
		REDIS_PORT:     getEnv("REDIS_PORT", "6379"),
		REDIS_PASSWORD: getEnv("REDIS_PASSWORD", ""),
		REDIS_DB:       getEnv("REDIS_DB", "0"),

		QiniuAccessKey:     getEnv("QINIU_ACCESS_KEY", ""),
		QiniuSecretKey:     getEnv("QINIU_SECRET_KEY", ""),
		QiniuBucket:        getEnv("QINIU_BUCKET", ""),
		QiniuDomain:        getEnv("QINIU_DOMAIN", ""),
		QiniuZone:          getEnv("QINIU_ZONE", "z0"),
		UploadPrefix:       getEnv("UPLOAD_PREFIX", ""),
		UploadPrefixAdmin:  getEnv("UPLOAD_PREFIX_ADMIN", ""),
		UploadPrefixQrcode: getEnv("UPLOAD_PREFIX_QRCODE", ""),

		AppId:     getEnv("APPID", ""),
		AppSecret: getEnv("APPSECRET", ""),

		WechatPayMchID:          getEnv("WECHATPAYMCHID", ""),
		WechatPayApiV3Key:       getEnv("WECHATAPIV3", ""),
		WechatPayPrivateKeyPath: getEnv("WECHATPAYPRIVATEKEYPATH", ""),
		WechatPaySerialNo:       getEnv("WECHATPAYPRIVATECERT", ""),
		PublicKeyPath:           getEnv("PUBLICKEYPATH", ""),
		PublicKeySerialNo:       getEnv("PUBLICKEYSERIALNO", ""),
		WechatPayNotifyURL:      getEnv("WECHATPAYNOTIFYURL", ""),

		AmapKey:        getEnv("AMAP_KEY", ""),
		QWeatherKey:    getEnv("QWEATHER_KEY", ""),
		QWeatherKID:    getEnv("QWEATHER_KID", ""),
		QWeatherPID:    getEnv("QWEATHER_PID", ""),
		DeepSeekApiKey: getEnv("DEEPSEEK_API_KEY", ""),
	}
	// 如果 QWeatherKey 看起来像一个文件路径且文件存在，优先从文件读取密钥内容
	if AppConfig.QWeatherKey != "" {
		if _, err := os.Stat(AppConfig.QWeatherKey); err == nil {
			if b, err := os.ReadFile(AppConfig.QWeatherKey); err == nil {
				AppConfig.QWeatherKey = strings.TrimSpace(string(b))
			}
		}
	}
	AppConfig.SnowflakeMachineID = getEnvInt64("SNOWFLAKE_MACHINE_ID", 1)
	AppConfig.Env = getEnv("ENV", "development")
}

// getEnv 读取环境变量，不存在时返回默认值
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt64 读取 int64 类型的环境变量
func getEnvInt64(key string, fallback int64) int64 {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

// getEnvFloat64 读取 float64 类型的环境变量（备用）
func getEnvFloat64(key string, fallback float64) float64 {
	if val, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}
