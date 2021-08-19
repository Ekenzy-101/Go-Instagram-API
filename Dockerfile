FROM golang:1.16-alpine3.14 as build_image
# FROM golang:1.16-buster as build_image

WORKDIR /app

COPY ["go.mod", "go.sum", "./"] 

RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:3.14
# FROM gcr.io/distroless/base-debian10

RUN apk add --no-cache ca-certificates

WORKDIR /

COPY --from=build_image /app/main .

EXPOSE 5000

ENTRYPOINT ["./main"]