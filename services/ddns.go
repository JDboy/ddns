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
	return local.R.CreateOrUpdate(domain.Domain, domain.IP, domain.Value)
}

//QueryHandle 查询域名信息
func (u *DdnsHandler) QueryHandle(ctx hydra.IContext) (r interface{}) {
	ctx.Log().Info("--------------查询域名信息---------------")
	return local.R.GetDomainDetails()
}
