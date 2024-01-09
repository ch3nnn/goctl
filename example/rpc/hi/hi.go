package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"gitlab.bolean.com/sa-micro-team/goctl/example/rpc/hi/internal/config"
	eventServer "gitlab.bolean.com/sa-micro-team/goctl/example/rpc/hi/internal/server/event"
	greetServer "gitlab.bolean.com/sa-micro-team/goctl/example/rpc/hi/internal/server/greet"
	"gitlab.bolean.com/sa-micro-team/goctl/example/rpc/hi/internal/svc"
	"gitlab.bolean.com/sa-micro-team/goctl/example/rpc/hi/pb/hi"
)

var configFile = flag.String("f", "etc/hi.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		hi.RegisterGreetServer(grpcServer, greetServer.NewGreetServer(ctx))
		hi.RegisterEventServer(grpcServer, eventServer.NewEventServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
