# fx-sample-app

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
protoc --proto_path=proto/ \
--go_out=proto/fxsample/ --go_opt=module=fx-sample-app/proto/fxsample \
--go-grpc_out=proto/fxsample/ --go-grpc_opt=module=fx-sample-app/proto/fxsample \
--grpc-gateway_out=proto/fxsample --grpc-gateway_opt=module=fx-sample-app/proto/fxsample \
./proto/fxsample/sample.proto
```
