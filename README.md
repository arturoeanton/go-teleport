# go-teleport
Teleport


# Run in cloud
```
   go run .  -addr1 8081 -addr2 8082
```


# Run in local

```
    docker run --name some-nginx -d -p 8080:80 nginx
    go run . -addr1 cloud.com:8082 -addr2 127.0.0.1:8080
```
