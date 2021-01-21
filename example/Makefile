GO_SRC=${GOPATH}/src
GOOGLE_API_DIR=${GO_SRC}/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/
GOGO_API_DIR=${GO_SRC}/github.com/gogo/protobuf/protobuf
gen_rpc = protoc -I=. -I=${GOPATH}/src -I=${GOOGLE_API_DIR} -I=${GOGO_API_DIR}
proto_file = api/demo.proto

rpc:
	$(gen_rpc) --gogo_out=plugins=grpc:. $(proto_file)
	$(gen_rpc) --gin_out=:. $(proto_file)