FROM bigrocs/golang-gcc:v1.16 as builder

ARG ACCES_STOKEN
RUN apk add git
RUN go env -w GOPRIVATE=github.com/lecex,github.com/bigrocs
RUN git config --global url."https://bigrocs:${ACCES_STOKEN}@github.com".insteadOf "https://github.com"

RUN git config --global --list

WORKDIR /go/src/github.com/lecex/vipspt
COPY . .
ENV GO111MODULE=on CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go mod tidy
RUN go build -a -installsuffix cgo -o bin/vipspt

FROM bigrocs/alpine:ca-data

RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=builder /go/src/github.com/lecex/vipspt/bin/vipspt /usr/local/bin/
CMD ["vipspt"]
EXPOSE 8080
