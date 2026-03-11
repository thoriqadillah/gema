package gema

import (
	"fmt"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type GrpcService interface {
	Register(server *grpc.Server)
}

func AsGrpcService(constructor any) any {
	return fx.Annotate(
		constructor,
		fx.As(new(GrpcService)),
		fx.ResultTags(`group:"grpc_services"`),
	)
}

type grpcParams struct {
	fx.In

	fx.Lifecycle
	*grpc.Server
	Services []GrpcService `group:"grpc_services"`
}

func StartGrpc(host, port string) fx.Option {
	return fx.Module("start_grpc",
		fx.Invoke(func(p grpcParams) {
			for _, service := range p.Services {
				service.Register(p.Server)
			}

			p.Append(fx.StartHook(func() error {
				fmt.Println("[Gema] Starting gRPC server on " + host + port)
				listener, err := net.Listen("tcp", host+port)
				if err != nil {
					return err
				}

				go func() {
					if err := p.Server.Serve(listener); err != nil {
						fmt.Println("[Gema] gRPC server stopped with error: " + err.Error())
					}
				}()
				return nil
			}))

			p.Append(fx.StopHook(p.GracefulStop))
		}),
	)
}
