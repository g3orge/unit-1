BINARY_NAME=unit1
build:
	GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME}-darwin main.go
 	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux main.go
 	GOARCH=amd64 GOOS=window go build -o ${BINARY_NAME}-windows main.go

run:
	go run -o ${BINARY_NAME} main.go
	./${BINARY_NAME}

clean:
    go clean
    rm ${BINARY_NAME}