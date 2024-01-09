package svc

import (
	"gitlab.bolean.com/sa-micro-team/goctl/example/rpc/hi/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
