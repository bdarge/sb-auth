package interceptors

import (
	"context"
	"fmt"
	"github.com/bdarge/auth/out/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor TODO use external validation lib
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		switch req.(type) {
		case *auth.RegisterRequest:
			if err := validate("Password", req.(*auth.RegisterRequest).Password); err != nil {
				return nil, err
			}
			if err := validate("Email", req.(*auth.RegisterRequest).Email); err != nil {
				return nil, err
			}
		case *auth.LoginRequest:
			if err := validate("Password", req.(*auth.LoginRequest).Password); err != nil {
				return nil, err
			}
			if err := validate("Email", req.(*auth.LoginRequest).Email); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

func validate(field string, value string) error {
	if value == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%s: Field is required", field))
	}
	return nil
}
