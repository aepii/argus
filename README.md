protoc --proto_path=proto --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative vector_store.proto

docker build -t argus .
docker run -p 50051:50051 --env-file ./.env argus