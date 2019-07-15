package common

import (
	"crypto/x509"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GRPCDial creates a client connection to the given target.
func GRPCDial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup certificate pool")
	}

	creds := credentials.NewClientTLSFromCert(pool, "")
	return grpc.Dial(target,
		append([]grpc.DialOption{
			grpc.WithTransportCredentials(creds)},
			opts...)...)
}

// GRPCDialInsecureTransport creates an insecure client connection to the given target with no trasnport security.
func GRPCDialInsecureTransport(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
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
