proto:
	 mkdir -p out && protoc ./pb/auth.proto  --go_out=:. --go-grpc_out=:. \
	--go_opt=Mpb/auth.proto=github.com/bdarge/auth/out/auth \
	--go_opt=module=github.com/bdarge/auth \
	--go-grpc_opt=Mpb/auth.proto=github.com/bdarge/auth/out/auth \
	--go-grpc_opt=module=github.com/bdarge/auth
server:
	go run cmd/main.go