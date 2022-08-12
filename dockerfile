FROM golang:1.17-alpine
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ENV CGO_ENABLED 0
RUN go build -o main .
CMD ["/app/main"]