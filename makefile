


migrate-up:
	@echo "Running migrations up..."
	@go run cmd/migrator/main.go \
		-db "postgres://postgres:postgres@localhost:5432/postgres" \
  		-migrations-path "./migrations" \
		-direction=up


migrate-down:
	@echo "Running migrations down..."
	@go run cmd/migrator/main.go \
		-db "postgres://postgres:postgres@localhost:5432/postgres" \
  		-migrations-path "./migrations" \
		-direction=down


