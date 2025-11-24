DB_URL=postgresql://root:secret@localhost:5432/message-app?sslmode=disable
# PROD_DB_URL in app.env

# to pull an image
pull:
	docker pull postgres:17-alpine

# to run container from postgres image
# localhost network 5432:5432 container network
postgres:
	docker run --name postgres17 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:17-alpine

# creates migration file in db/migration
new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

# docker exec to run commands inside running container
createdb:
	docker exec -it postgres17 createdb --username=root --owner=root message-app

dropdb:
	docker exec -it postgres17 dropdb message-app

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

# for production (AWS RDS)
migrateupp:
	migrate -path db/migration -database "$(PROD_DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	@echo "👉 Running short tests with coverage..."
	go test -v -short ./...

test-race:
	@echo "⚙️ Running race detection..."
	go test -race -v ./...

server:
	go run main.go

client:
	go run client/cmd/main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/kratos069/message-app/db/sqlc Store

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=message-app \
    proto/*.proto
	statik -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker run --name redis -p 6379:6379 -d redis:8-alpine

app-image:
	docker build -t message-app:1.0 .

metrics-view-sc:
	expvarmon -ports="localhost:3010" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

metrics-view:
	expvarmon -ports="localhost:4020" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"
	
talk-metrics:
	expvarmon -ports="localhost:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"


.PHONY: app-image test-race migrateupp client new_migration redis evans proto db_schema db_docs postgres createdb dropdb migrateup migratedown sqlc test server mock migrateup1 migratedown1