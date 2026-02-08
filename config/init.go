package config

import (
	errorWrapper "serv-executor-bot/internal/utils/error-wrapper"
	"serv-executor-bot/internal/utils/trace"

	"github.com/spf13/viper"
)

type Config struct {
	BotApiToken string  `mapstructure:"bot_api_token"`
	AdminIds    []int64 `mapstructure:"admin_ids"`
	ServerIp    string  `mapstructure:"server_ip"`
}

func Init() (config *Config, err error) {
	viper.AddConfigPath("config/")
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if err != nil {
		return nil, errorWrapper.Error(trace.GetFuncName(), err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errorWrapper.Error(trace.GetFuncName(), err)
	}
	return
}
