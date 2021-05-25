FROM golang:alpine

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GIN_MODE=release \
    MONGODB_NAME=gomongo \
    MONGO_URI=mongodb://localhost:27017 \
    CLIENT_ORIGIN=http://localhost:3000 \
    APP_ACCESS_SECRET=tEuxpYpBmyMpUwBxD1mjYbWcrPSB57BP

WORKDIR /build

COPY ["go.mod", "go.sum", "./"] 
RUN go mod download

COPY . .

RUN go build -o main .

WORKDIR /dist

RUN cp /build/main .

EXPOSE 5000

CMD ["/dist/main"]