# go-redisのテスト
# usage
```
docker pull redis
docker run -p 6379:6379 --name some-redis -d redis
go mod tidy
go run redis.go
```
