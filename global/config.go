package global

type GConfig struct {
	Language string `json:"language"`
	LogDir   string `json:"log_dir"`
}

var GlobalConfig = &GConfig{
	Language: "zh-CN",
	LogDir:   "logs",
}