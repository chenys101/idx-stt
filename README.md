# 启动
## create file data.db
## copy configs/config_demo.yaml to config.yaml
## go mod tidy
## go run cmd/main.go

curl -X POST http://localhost:9002/api/v1/std/createStdMonitor -H "Content-Type: application/json" -d '{"MonitorValue":"shxxx,shxxx,szxxx","Code":"sg"}'
while true; do curl http://localhost:9002/api/v1/std/nc?code=sg; echo '--------------------------------------'; sleep 60; done


