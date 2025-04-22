.PHONY: help lint deps keys

# ===========================
# HELP: Список доступных команд
# ===========================
help:  ## Список доступных команд
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "\033[1;32m%-20s\033[0m %s\n", $$1, $$2}'

# ===========================
# LINT: Анализ и форматирование кода
# ===========================
lint:  ## Запускает линтер, форматирует код и генерирует отчет
	-fieldalignment -fix ./... || true
	-goimports -w . || true
	-gofmt -w . || true
	-golines -m 110 -w . || true
	mkdir -p ./golangci-lint
	-golangci-lint run -c .golangci.yml --out-format json > ./golangci-lint/report-unformatted.json || true
	cat ./golangci-lint/report-unformatted.json | jq '{IssuesCount: (.Issues | length), Issues: [.Issues[] | {FromLinter, Text, SourceLines, Filename: .Pos.Filename, Line: .Pos.Line}]}' > ./golangci-lint/report.json
	rm ./golangci-lint/report-unformatted.json

# ===========================
# DEPS: Обновление зависимостей
# ===========================
deps:  ## Очищает и обновляет зависимости проекта
	rm -f go.sum
	go mod tidy -v
	go mod verify

# ===========================
# KEYS: Генерация ключей
# ===========================
keys:  ## Генерирует приватный и публичный ключи
	go run ./cmd/keycli/main.go -private private.pem -public public.pem -size 4096