package main

import (
	"context"
	"io"
	"log"
	"net"

	pb "github.com/anandu86130/Microservice/pb"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedUserServiceServer
}

// Unary RPC
func (s *Server) UserSignup(cxt context.Context, req *pb.UserCreate) (*pb.Response, error) {
	return &pb.Response{Status: "Success", Message: "User Signed up successfully"}, nil
}

func (s *Server) ListUsers(req *pb.FetchAll, stream pb.UserService_ListUsersServer) error {
	users := []*pb.UserList{
		{Id: 1, Username: "Anandu", Email: "anandu86130@gmail.com"},
		{Id: 2, Username: "Aswin", Email: "aswin@gmail.com"},
	}

	for _, user := range users {
		if err := stream.Send(user); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) UploadUsers(stream pb.UserService_UploadUsersServer) error {
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			// End of stream, send response back to client
			return stream.SendAndClose(&pb.Response{Status: "success", Message: "Users uploaded successfully"})
		}
		if err != nil {
			return err
		}
		log.Printf("Received User: %v", request.Username)
	}
}

func (s *Server) Chat(stream pb.UserService_ChatServer) error {
	for {
		// Receive a message from the client
		req, err := stream.Recv()
		if err == io.EOF {
			// End of the stream
			log.Println("End of stream")
			return nil
		}
		if err != nil {
			// Handle errors
			log.Printf("Error receiving message: %v", err)
			return err
		}

		log.Printf("Received message: %v", req.Content)

		// Send a response to the client
		if err := stream.Send(&pb.MessageResponse{Reply: "Echo: " + req.Content}); err != nil {
			log.Printf("Error sending message: %v", err)
			return err
		}
	}
}

func main() {
	list, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalf("failed to listen:%v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &Server{})

	log.Printf("server listening at %v", list.Addr())
	if err := s.Serve(list); err != nil {
		log.Fatalf("Failed to serve:%v", err)
	}
}
