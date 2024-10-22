package position

type PositionUpdateConfig struct {
	Enabled  bool `mapstructure:"enabled"`  // 是否进行启动
	Interval int  `mapstructure:"interval"` // 发送时间间隔
}
