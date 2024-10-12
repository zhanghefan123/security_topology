package delay

type DelayUpdateConfig struct {
	Enabled bool `mapstructure:"enabled"` // 判断是否进行延迟更新服务的启动
}
