BINARY_NAME=pod_service

fmt:
go fmt ./...
go mod tidy -v

dep:
 go mod download

vet:
 go vet

lint:
 golangci-lint run --enable-all

    
run:
    docker compose up --build
    
rebuild_and_run:
	docker-compose down ; docker build --rm -t --no-cache ${BINARY_NAME}:latest

stop_all:   
	docker-compose down
	
stop_containers:	 
	docker stop$(docker ps -a)  
	
test_all:
	go test -v  