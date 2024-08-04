package main

import (
	pb "github.com/anandu86130/Microservice/cmd"
)

type UserHandler struct {
	grpcClient pb.UserServiceClient
}