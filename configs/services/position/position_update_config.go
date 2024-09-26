package position

type PositionUpdateConfig struct {
	ImageName string `mapstructure:"image_name"`
	Interval  int    `mapstructure:"interval"`
}
