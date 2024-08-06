package main

import (
	"context"
	"io"
	"log"
	"net/http"

	pb "github.com/anandu86130/Microservice/pb"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserHandler struct {
	grpcClient pb.UserServiceClient
}

func NewUserHandler(grpcClient pb.UserServiceClient) *UserHandler {
	return &UserHandler{grpcClient: grpcClient}
}

func (h *UserHandler) UserSignup(c *gin.Context) {
	var req pb.UserCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.grpcClient.UserSignup(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	req := pb.FetchAll{}
	stream, err := h.grpcClient.ListUsers(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var users []*pb.UserList
	for {
		user, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) UploadUsers(c *gin.Context) {
	stream, err := h.grpcClient.UploadUsers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var users []*pb.UserCreate
	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	for _, user := range users {
		if err := stream.Send(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	result, err := stream.CloseAndRecv()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *UserHandler) Chat(c *gin.Context) {
	stream, err := h.grpcClient.Chat(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func() {
		defer stream.CloseSend()
		for {
			var request pb.Message
			if err := c.ShouldBindJSON(&request); err != nil {
				log.Printf("failed to bind request: %v", err)
				return
			}

			if err := stream.Send(&request); err != nil {
				log.Printf("Failed to send message: %v", err)
				return
			}
		}
	}()

	for {
		result, err := stream.Recv()
		if err == io.EOF {
			log.Println("End of stream")
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func main() {
	connection, err := grpc.NewClient("localhost:3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer connection.Close()

	grpcClient := pb.NewUserServiceClient(connection)
	handler := NewUserHandler(grpcClient)

	router := gin.Default()
	router.POST("/usersignup", handler.UserSignup)
	router.GET("/userlist", handler.ListUsers)
	router.POST("/uploadusers", handler.UploadUsers)
	router.POST("/chat", handler.Chat)

	log.Println("Serving Gin HTTP server on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
