package gema

import (
	"fmt"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func StartGRPC(host, port string) fx.Option {
	return fx.Module("grpc",
		fx.Invoke(func(lc fx.Lifecycle, server *grpc.Server) {
			lc.Append(fx.StartHook(func() error {
				fmt.Println("[Gema] Starting gRPC server on " + host + port)
				listener, err := net.Listen("tcp", host+port)
				if err != nil {
					return err
				}

				go server.Serve(listener)
				return nil
			}))

			lc.Append(fx.StopHook(server.GracefulStop))
		}),
	)
}
