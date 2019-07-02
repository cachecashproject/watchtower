package common

import (
	"google.golang.org/grpc"
)

// GRPCDial creates a client connection to the given target.
// XXX: No transport security!
func GRPCDial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(target,
		append([]grpc.DialOption{
			grpc.WithInsecure()},
			opts...)...)
}

// NewGRPCServer creates a new grpc server
func NewGRPCServer(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(
		append([]grpc.ServerOption{},
			opt...)...)
}
