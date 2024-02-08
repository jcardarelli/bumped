FROM golang:alpine3.18 AS build
WORKDIR /build
ENV CGO_ENABLED=1
RUN apk add --no-cache \
    gcc \
    musl-dev
COPY go.mod go.sum main.go /build/
COPY templates/ /build/templates/
RUN go mod download && \
    # go install -ldflags='-s -w -extldflags "-static"' ./main.go && \
	go build -a -installsuffix cgo -o fancy-api .

FROM alpine:latest
WORKDIR /app/
ENV GIN_MODE=debug
COPY --from=build /build/templates/ ./templates/
COPY --from=build /build/fancy-api .
RUN mkdir /app/nocodb
EXPOSE 8083
CMD ["/app/fancy-api"]
