# Bumped
Golang API for managing restaurant info

## Stack
* Gin Web Framework for routing
* Sqlite for database

## Docker
```
docker build -t bumped:latest .
docker run -p 8083:8083 --rm bumped:latest
```

## Build
```
go build -o bumped
```

## Run
```
go run main.go
```

or if you built the binary:
```
./bumped
```
