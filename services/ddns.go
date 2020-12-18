package services

import (
	"github.com/micro-plat/ddns/local"
	"github.com/micro-plat/hydra"
)

// DdnsHandler Handler
type DdnsHandler struct {
}

// NewDdnsHandler 构建DdnsHandler
func NewDdnsHandler() *DdnsHandler {
	return &DdnsHandler{}
}

//RequestHandle 保存动态域名信息
func (u *DdnsHandler) RequestHandle(ctx hydra.IContext) (r interface{}) {
	ctx.Log().Info("--------------保存动态域名信息---------------")

	ctx.Log().Info("1. 检查必须参数")
	var domain Domain
	if err := ctx.Request().Bind(&domain); err != nil {
		return err
	}

	ctx.Log().Info("2. 检查并创建解析信息")
	if err := local.R.CreateOrUpdate(domain.Domain, domain.IP, domain.Value); err != nil {
		return err
	}
	return "success"
}

//QueryHandle 查询域名信息
func (u *DdnsHandler) QueryHandle(ctx hydra.IContext) (r interface{}) {
	ctx.Log().Info("--------------查询域名信息---------------")
	return local.R.GetDomainDetails()
}

//PlatNamesHandle 查询平台名及对应的域名信息
func (u *DdnsHandler) PlatNamesHandle(ctx hydra.IContext) (r interface{}) {
	ctx.Log().Info("--------------查询平台名及对应的域名信息---------------")

	ctx.Log().Info("1. 获取注册中心")
	rgst, err := registry.GetRegistry(hydra.G.RegistryAddr, ctx.Log())
	if err != nil {
		return err
	}

	ctx.Log().Info("2. 获取域名节点")
	domains, _, err := rgst.GetChildren("/dns")
	if err != nil {
		return err
	}

	ctx.Log().Info("3. 处理域名")
	result := make(map[string]string, 0)
	for _, domain := range domains {
		val, _, err := rgst.GetValue(registry.Join("/dns", domain))
		if err != nil {
			return err
		}
		value := make(types.XMap, 0)
		err = json.Unmarshal(val, &value)
		if err != nil {
			ctx.Log().Errorf("处理域名%s:%v", domain, err)
			continue
		}
		cnPlatName := value.GetString("cn_plat_name")
		result[cnPlatName] = domain
	}
	return result
}
