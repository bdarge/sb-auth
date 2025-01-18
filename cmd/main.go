package main

import (
	"flag"
	"github.com/bdarge/auth/pkg/interceptors"
	"github.com/bdarge/auth/pkg/models"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
	"time"

	"github.com/bdarge/auth/out/auth"
	"github.com/bdarge/auth/pkg/config"
	"github.com/bdarge/auth/pkg/db"
	"github.com/bdarge/auth/pkg/services"
	"github.com/bdarge/auth/pkg/utils"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

var (
	sleep  = flag.Duration("sleep", time.Second*5, "duration between changes in health")
	system = "" // empty string represents the health of the system
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

	var programLevel = new(slog.LevelVar)
	programLevel.Set(c.LogLevel)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel}))
	slog.SetDefault(logger)

	dbHandler := db.Init(c.DSN)

	jwt := utils.JwtWrapper{
		TokenSecretKey:        c.TokenSecretKey,
		Issuer:                c.TokenIssuer,
		TokenExpOn:            c.TokenExpOn,
		RefreshTokenSecretKey: c.RefreshTokenSecretKey,
		RefreshTokenExpOn:     c.RefreshTokenExpOn,
	}

	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalf("Listing on port %s has failed: %v", c.Port, err)
	}

	slog.Info("Auth service is listening", "Port", c.Port)

	s := services.Server{
		DBHandler: dbHandler,
		Jwt:       jwt,
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.UnaryServerInterceptor()))
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)

	auth.RegisterAuthServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}

	go func() {
		// asynchronously inspect dependencies and toggle serving status as needed
		next := healthpb.HealthCheckResponse_SERVING

		for {
			healthcheck.SetServingStatus(system, next)
			err = isDbConnectionWorks(s.DBHandler.DB)
			if err != nil {
				next = healthpb.HealthCheckResponse_NOT_SERVING
			} else {
				next = healthpb.HealthCheckResponse_SERVING
			}
			time.Sleep(*sleep)
		}
	}()
}

func isDbConnectionWorks(DB *gorm.DB) error {
	return DB.First(&models.Account{}).Error
}
