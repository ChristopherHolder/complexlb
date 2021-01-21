FROM golang:1.13 AS builder
WORKDIR /app
COPY go.mod loadbalancer cmd ./
RUN mkdir loadbalancer cmd &&\
    mv main.go cmd &&\
    mv lb.go algo.go healthcheck.go loadbalancer
WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o lb .
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /app/cmd/lb .
ENTRYPOINT [ "/root/lb" ]