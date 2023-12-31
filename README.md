# fx-sample-app

# In this sample service
- sample fx dependency injection
- sample life cycle hooks
- sample redis cache implementation
- sample config provider for use with fx
- sample gRPC
- sample REST proxy requests to gRPC endpoints
- sample postgres concurrent actions

# Below is a how to for generating the needed proto files.
## Env Vars
Add GOPATH/bin to PATH.

## Install plugins
```
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## generate proto structs, client, server, openapi proxy gateway
```
protoc --proto_path=proto/ --go_out=proto/fxsample/ --go_opt=module=fx-sample-app/proto/fxsample --go-grpc_out=proto/fxsample/ --go-grpc_opt=module=fx-sample-app/proto/fxsample --grpc-gateway_out=proto/fxsample --grpc-gateway_opt=module=fx-sample-app/proto/fxsample ./proto/fxsample/sample.proto
```
