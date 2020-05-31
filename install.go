// +build prod

package main

import (
	"github.com/micro-plat/hydra"
	"github.com/micro-plat/hydra/conf/server/api"
	"github.com/micro-plat/hydra/registry/conf/server/header"
)

//bindConf 绑定启动配置， 启动时检查注册中心配置是否存在，不存在则引导用户输入配置参数并自动创建到注册中心
func init() {
	hydra.Conf.Ready(func() {
		hydra.Conf.API(":9090", api.WithTimeout(300, 300)).
			Header(header.WithCrossDomain())
	})
}
