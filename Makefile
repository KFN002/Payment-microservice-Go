CONFIG_PATH = configs/local.env
include ${CONFIG_PATH}

run:
	CONFIG_PATH=${CONFIG_PATH} go run main.go

build:
	CONFIG_PATH=${CONFIG_PATH} go build main.go

test:
	go test -cover ./...

generate:
	protoc --proto_path=proto proto/*.proto --go_out=. --go-grpc_out=.
	sqlc -f sqlc/sqlc.yml generate 

up:
	docker compose -f deployments/docker-compose.yml --env-file ${CONFIG_PATH} up -d

stop:
	docker compose -f deployments/docker-compose.yml --env-file ${CONFIG_PATH} stop

db-up:
	docker compose -f deployments/docker-compose.yml --env-file ${CONFIG_PATH} up -d postgres

db-stop:
	docker compose -f deployments/docker-compose.yml --env-file ${CONFIG_PATH} stop postgres

db-sh:
	docker exec -ti postgres psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}
