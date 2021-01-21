# protoc-gen-gin

将proto文件生成 gin http server

调整一下接口：
*  *gin.Context 作为接口
* ShouldBindWith

跟：
* github.com/grpc-ecosystem/grpc-gateway
* github.com/gogo/protobuf

搭配使用更佳

## Install
```shell
go install github.com/liu-junyong/protoc-gen-gin
```

## Usage
```shell
protoc --gin_out=:. path_to_your_proto
```
check [Makefile](example/Makefile) for more usage.
