package config

import (
	"github.com/lecex/user/core/config"
	"github.com/lecex/user/core/env"
)

// Conf 配置
var Conf config.Config = config.Config{
	Name:    env.Getenv("MICRO_API_NAMESPACE", "go.micro.api.") + "vipspt",
	Version: "v1.3.2",
}
