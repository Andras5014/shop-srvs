package config

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"db" json:"db"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}
type RedisConfig struct {
	Host       string `mapstructure:"host" json:"host"`
	Port       int    `mapstructure:"port" json:"port"`
	Expiration int    `mapstructure:"expiration" json:"expiration"`
}
type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}
type JaegerConfig struct {
	Name string `mapstructure:"name" json:"name"`
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}
type NacosConfig struct {
	Host      string `mapstructure:"host" json:"host"`
	Port      int    `mapstructure:"port" json:"port"`
	Namespace string `mapstructure:"namespace" json:"namespace"`
	User      string `mapstructure:"user" json:"user"`
	Password  string `mapstructure:"password" json:"password"`
	DataId    string `mapstructure:"data_id" json:"data_id"`
	Group     string `mapstructure:"group" json:"group"`
}
type GoodsServerConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
	Name string `mapstructure:"name" json:"name"`
}
type InventoryServerConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
	Name string `mapstructure:"name" json:"name"`
}
type ServerConfig struct {
	Name                string                `mapstructure:"name" json:"name"`
	Host                string                `mapstructure:"host" json:"host"`
	Tags                []string              `mapstructure:"tags" json:"tags"`
	RedisInfo           RedisConfig           `mapstructure:"redis" json:"redis"`
	MysqlInfo           MysqlConfig           `mapstructure:"mysql" json:"mysql"`
	ConsulInfo          ConsulConfig          `mapstructure:"consul" json:"consul"`
	GoodsServerInfo     GoodsServerConfig     `mapstructure:"goods_srv" json:"goods_srv"`
	InventoryServerInfo InventoryServerConfig `mapstructure:"inventory_srv" json:"inventory_srv"`
	JaegerInfo          JaegerConfig          `mapstructure:"jaeger" json:"jaeger"`
}
