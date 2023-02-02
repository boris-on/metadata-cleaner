```bash
go run cmd/main.go
```
```bash
docker build -t metawipe .
docker run --restart always -dp 80:80 --name metawipe_container metawipe 
```