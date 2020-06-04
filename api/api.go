package api

import (
	identiconservice "github.com/murryIsDeveloping/identiconGRPC/api/identicon"
	identiconpb "github.com/murryIsDeveloping/identiconGRPC/api/identicon/proto"
	grpc "google.golang.org/grpc"
)

// CreateGRPCServer registers all the services to the server
func CreateGRPCServer() *grpc.Server {
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	identiconpb.RegisterIdenticonServiceServer(s, &identiconservice.IdenticonService{})
	return s
}
