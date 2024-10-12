package position

type PositionUpdateConfig struct {
	Enabled   bool   `mapstructure:"enabled"`    // 是否进行启动
	ImageName string `mapstructure:"image_name"` // 镜像的名称
	Interval  int    `mapstructure:"interval"`   // 发送时间间隔
}
