module github.com/liu-junyong/protoc-gen-gin

go 1.15

replace github.com/312362115/protoc-gen-gin/ecode => ./ecode

require (
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/go-kratos/kratos v0.6.0
	github.com/golang/protobuf v1.4.3
	github.com/pkg/errors v0.8.1
)
