# fx-sample-app

## generate proto structs, client, server, openapi proxy gateway
```
protoc --proto_path=proto/ --go_out=proto/fxsample/ --go_opt=module=proto/fxsample --go-grpc_out=proto/fxsample/ --go-grpc_opt=module=proto/fxsample --grpc-gateway_out=proto/fxsample --grpc-gateway_opt=module=proto/fxsample ./proto/fxsample/sample.proto
```
