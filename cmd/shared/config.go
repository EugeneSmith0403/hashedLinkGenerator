package shared

import "link-generator/configs"

func LoadConfigs(config ...*configs.Config) *configs.Config {
	if len(config) > 0 {
		return config[0]
	}
	return configs.LoadConfig()
}
