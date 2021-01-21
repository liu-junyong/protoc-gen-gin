# protoc-gen-gin

将proto文件生成 gin http server

调整一下接口：
*  *gin.Context 作为接口
* ShouldBind
* 可以自定义错误返回，可以在错误中自定义log
* 在实现层次：添加了 errno 代替 error,这样有错误码了

跟：
* github.com/grpc-ecosystem/grpc-gateway
* github.com/gogo/protobuf


依赖组件：
```bazaar
go get github.com/liu-junyong/errno
```
https://github.com/liu-junyong/errno/blob/main/errno.go

就一个文件，用来代替默认的error,如果不喜欢，用replace在 go.mod 替换掉

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
