.PHONY: run
run:
	@echo "🔨 running $(APP_NAME)..."
	@rm db.db
	go run cmd/main.go
	hexdump -C db.db
