run-with-db:
	go run solution_db/main.go

run-with-redis:
	go run solution_redis/main.go

.PHONY: run-with-db run-with-redis