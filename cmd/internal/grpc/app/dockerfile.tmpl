FROM golang:1.23.1 AS builder

WORKDIR /project

COPY . .

RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN go mod download
RUN go generate ./pkg/{{.AppPkg}}/ent
RUN go mod tidy

RUN CGO_ENABLED=0 go build ./pkg/{{.AppPkg}} -o ./bin/main

FROM scratch

WORKDIR /bin

COPY --from=builder /project/bin/main .

EXPOSE 8000
EXPOSE 50051

ENTRYPOINT ["./main"]

CMD ["--config", "/data/config.toml"]
