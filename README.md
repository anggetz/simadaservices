how to run

#requirement
1. install nats-server
2. caddy


##RUN in project simada services
1. RUN nats-server
2. RUN go run cmd/service-config/main.go
3. RUN go run cmd/service-auth/main.go
3. RUN go run cmd/service-transaction/main.go
3. RUN go run cmd/service-report/main.go
3. RUN go run cmd/service-worker/main.go
3. RUN desire service

##RUN in project simada
1. php artisan serve 
2. caddy run --config=docker-config/local/Caddyfile