package main

import (
	"fmt"
	"github.com/bdarge/auth/pkg/interceptors"
	"log"
	"net"
	"os"

	"github.com/bdarge/auth/out/auth"
	"github.com/bdarge/auth/pkg/config"
	"github.com/bdarge/auth/pkg/db"
	"github.com/bdarge/auth/pkg/services"
	"github.com/bdarge/auth/pkg/utils"
	"google.golang.org/grpc"
)

func main() {
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "dev"
	}
	c, err := config.LoadConfig(environment)

	if err != nil {
		log.Fatalln("Failed loading config", err)
	}

	dbHandler := db.Init(c.DSN)

	jwt := utils.JwtWrapper{
		SecretKey:       c.JWTSecretKey,
		Issuer:          "go-grpc-auth",
		ExpirationHours: 50,
	}

	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalf("Listing on port %s has failed: %v", c.Port, err)
	}

	fmt.Printf("auth service is listening on %s", c.Port)

	s := services.Server{
		DBHandler: dbHandler,
		Jwt:       jwt,
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.UnaryServerInterceptor()))

	auth.RegisterAuthServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
