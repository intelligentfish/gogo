package app_cfg

import (
	"github.com/intelligentfish/gogo/xstring"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"sync"
)

// 全局变量
var (
	once     sync.Once // 执行一次
	instance *AppCfg   // 单实例
)

// RoutinePoolCfg 协程池配置
type RoutinePoolCfg struct {
	MinRoutinePoolSize int `json:"minRoutinePoolSize" yaml:"minRoutinePoolSize"` // 协程池大小
}

// MongoDBCfg 数据库配置
type MongoDBCfg struct {
	MongoDBUrl string `json:"mongoDbUrl" yaml:"mongoDbUrl"` // MongoDB 连接地址
	DbName     string `json:"dbName" yaml:"dbName"`         // 数据库名
}

// MySQLDBCfg 数据库配置
type MySQLDBCfg struct {
	Prefix   string `json:"prefix" yaml:"prefix"`     // 表明前缀
	Username string `json:"username" yaml:"username"` // 用户名
	Password string `json:"password" yaml:"password"` // 密码
	Host     string `json:"host" yaml:"host"`         // 主机
	DB       string `json:"db" yaml:"db"`             // 数据库
}

// RedisCfg Redis配置
type RedisCfg struct {
	Prefix   string `json:"prefix" yaml:"prefix"`     // 前缀
	Address  string `json:"address" yaml:"address"`   // Redis地址
	Password string `json:"password" yaml:"password"` // 密码
	DB       int    `json:"db" yaml:"db"`             // 数据库
}

// RabbitMQ 配置
type RabbitMQCfg struct {
	Url                 string `json:"url" yaml:"url"`                                 // URL
	GameSettleQueueName string `json:"gameSettleQueueName" yaml:"gameSettleQueueName"` // 结算队列名
}

// TCP服务器 配置
type TCPServiceConfig struct {
	Port int `json:"port" yaml:"port"` // 端口
}

// HTTP服务器配置
type HTTPServiceConfig struct {
	HTTPAddr           string `json:"httpAddr" yaml:"httpAddr"`                     // HTTP地址
	HTTPStaticLocation string `json:"httpStaticLocation" yaml:"httpStaticLocation"` // HTTP静态文件路径(URL)
	HTTPStaticPath     string `json:"httpStaticPath" yaml:"httpStaticPath"`         // HTTP静态文件本地路径
	HTTPSAddr          string `json:"httpsAddr" yaml:"httpsAddr"`                   // HTTPS地址
	HTTPSCertPath      string `json:"httpsCertPath" yaml:"httpsCertPath"`           // HTTP证书
	HTTPSKeyPath       string `json:"httpsKeyPath" yaml:"httpsKeyPath"`             // HTTP Key
	SignKey            string `json:"signKey" yaml:"signKey"`                       // 参数签名Key
}

// RPCService配置
type RPCServiceConfig struct {
	Address string `json:"address" yaml:"address"`
}

// AppCfg 应用配置
type AppCfg struct {
	Debug             bool               `json:"debug" yaml:"debug"`
	UUID              string             `json:"uuid" yaml:"uuid"`
	RoutinePoolCfg    *RoutinePoolCfg    `json:"routinePoolCfg" yaml:"routinePoolCfg"`
	MongoDBCfg        *MongoDBCfg        `json:"mongoDbCfg" yaml:"mongoDbCfg"`
	MySQLDBCfg        *MySQLDBCfg        `json:"mySqldbCfg" yaml:"mySqldbCfg"`
	RedisCfg          *RedisCfg          `json:"redisCfg" yaml:"redisCfg"`
	RabbitMQCfg       *RabbitMQCfg       `json:"rabbitMqCfg" yaml:"rabbitMqCfg"`
	HTTPServiceConfig *HTTPServiceConfig `json:"httpServiceConfig" yaml:"httpServiceConfig"`
	TCPServiceConfig  *TCPServiceConfig  `json:"tcpServiceConfig" yaml:"tcpServiceConfig"`
	RPCServiceConfig  *RPCServiceConfig  `json:"rpcServiceConfig" yaml:"rpcServiceConfig"`
}

// newAppCfg 工厂方法
func newAppCfg() *AppCfg {
	return &AppCfg{
		RoutinePoolCfg:    &RoutinePoolCfg{},
		MongoDBCfg:        &MongoDBCfg{},
		MySQLDBCfg:        &MySQLDBCfg{},
		RedisCfg:          &RedisCfg{},
		RabbitMQCfg:       &RabbitMQCfg{},
		HTTPServiceConfig: &HTTPServiceConfig{},
		TCPServiceConfig:  &TCPServiceConfig{},
		RPCServiceConfig:  &RPCServiceConfig{},
	}
}

// FromApolloMap 工厂方法
func (object *AppCfg) FromApolloMap(cfgMap map[string]string) {
	object.RoutinePoolCfg.MinRoutinePoolSize = xstring.String(cfgMap["routinePoolCfg.minRoutinePoolSize"]).
		ToInt(true)
	object.MongoDBCfg.MongoDBUrl = cfgMap["dbCfg.mongoDbUrl"]
	object.RedisCfg.Address = cfgMap["redisCfg.address"]
	object.RedisCfg.DB = xstring.String(cfgMap["redisCfg.db"]).ToInt(true)
	object.RabbitMQCfg.Url = cfgMap["rabbitMqCfg.url"]
	object.HTTPServiceConfig.HTTPAddr = cfgMap["httpServiceConfig.httpAddr"]
	object.HTTPServiceConfig.HTTPStaticLocation = cfgMap["httpServiceConfig.httpStaticLocation"]
	object.HTTPServiceConfig.HTTPStaticPath = cfgMap["httpServiceConfig.httpStaticPath"]
	object.HTTPServiceConfig.HTTPSAddr = cfgMap["httpServiceConfig.httpsAddr"]
	object.HTTPServiceConfig.HTTPSCertPath = cfgMap["httpServiceConfig.httpsCertPath"]
	object.HTTPServiceConfig.HTTPSKeyPath = cfgMap["httpServiceConfig.httpsKeyPath"]
	object.HTTPServiceConfig.SignKey = cfgMap["httpServiceConfig.signKey"]
	object.TCPServiceConfig.Port = xstring.String(cfgMap["tcpServiceConfig.port"]).ToInt(true)
	object.RPCServiceConfig.Address = cfgMap["rpcServiceConfig.address"]
}

// FromAppCfg 工厂方法
func (object *AppCfg) FromAppCfg(path string) (err error) {
	var fi os.FileInfo
	fi, err = os.Stat(path)
	if nil != err {
		if _, ok := err.(*os.PathError); !ok {
			return
		}
		err = nil
	}
	var bytes []byte
	if nil != fi {
		bytes, err = ioutil.ReadFile(path)
		if nil == err {
			err = yaml.Unmarshal(bytes, object)
		}
		return
	}
	bytes, err = yaml.Marshal(object)
	if nil == err {
		err = ioutil.WriteFile(path, bytes, 0600)
	}
	return
}

// valid routinePool
func (object *AppCfg) validRoutinePool() bool {
	return nil != object.RoutinePoolCfg &&
		0 < object.RoutinePoolCfg.MinRoutinePoolSize
}

// valid mongodb
func (object *AppCfg) validMongoDB() bool {
	return true
	return nil != object.MongoDBCfg &&
		0 < len(object.MongoDBCfg.MongoDBUrl)
}

// valid mysql
func (object *AppCfg) validMySQL() bool {
	return nil != object.MySQLDBCfg &&
		0 < len(object.MySQLDBCfg.Host) &&
		0 < len(object.MySQLDBCfg.Username) &&
		0 < len(object.MySQLDBCfg.Password) &&
		0 < len(object.MySQLDBCfg.DB) &&
		0 < len(object.MySQLDBCfg.Prefix)
}

// valid release
func (object *AppCfg) validRedis() bool {
	return nil != object.RedisCfg &&
		0 < len(object.RedisCfg.Prefix) &&
		0 < len(object.RedisCfg.Address)
}

// valid rabbitmq
func (object *AppCfg) validRabbitMQ() bool {
	return true
	return nil != object.RabbitMQCfg &&
		0 < len(object.RabbitMQCfg.Url)
}

// valid http server config
func (object *AppCfg) validHTTPService() bool {
	return true
	return nil != object.HTTPServiceConfig &&
		0 < len(object.HTTPServiceConfig.SignKey) &&
		(0 < len(object.HTTPServiceConfig.HTTPAddr) &&
			0 < len(object.HTTPServiceConfig.HTTPStaticLocation) &&
			0 < len(object.HTTPServiceConfig.HTTPStaticPath)) ||
		(0 < len(object.HTTPServiceConfig.HTTPSAddr) &&
			0 < len(object.HTTPServiceConfig.HTTPSCertPath) &&
			0 < len(object.HTTPServiceConfig.HTTPSKeyPath))
}

// valid http
func (object *AppCfg) validTCPService() bool {
	return true
	return nil != object.TCPServiceConfig &&
		0 < object.TCPServiceConfig.Port
}

// valid rpc
func (object *AppCfg) validRPC() bool {
	return true
	return nil != object.RPCServiceConfig &&
		0 < len(object.RPCServiceConfig.Address)
}

// Valid 校验配置
func (object *AppCfg) Valid() bool {
	return 0 < len(object.UUID) &&
		object.validRoutinePool() &&
		object.validMongoDB() &&
		object.validMySQL() &&
		object.validRedis() &&
		object.validRabbitMQ() &&
		object.validHTTPService() &&
		object.validTCPService() &&
		object.validRPC()
}

// GetInstance 获取单实例
func GetInstance() *AppCfg {
	once.Do(func() {
		instance = newAppCfg()
	})
	return instance
}
