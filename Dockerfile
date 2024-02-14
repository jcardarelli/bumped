FROM golang:alpine3.18 AS build
WORKDIR /build
ENV CGO_ENABLED=1
RUN apk add --no-cache \
    gcc \
    musl-dev
COPY go.mod go.sum vendor main.go /build/
COPY templates/ /build/templates/
RUN go mod tidy && \
    go mod vendor && \
    go build -mod vendor -installsuffix cgo -o bumped .

FROM alpine:latest
ARG DB=nocodb/restaurants.db
WORKDIR /app/
ENV GIN_MODE=debug
COPY --from=build /build/templates/ ./templates/
COPY --from=build /build/bumped .
RUN mkdir /app/nocodb
EXPOSE 8083
CMD ["/app/bumped"]
