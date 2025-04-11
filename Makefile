.PHONY: test test-coverage build clean help

# Variáveis
BINARY_DBBACKUP=bin/dbbackup
BINARY_UPLOADER=bin/uploader
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Ajuda
help:
	@echo "Comandos disponíveis:"
	@echo "  make test          - Executa todos os testes"
	@echo "  make test-coverage - Executa testes com cobertura e gera relatório HTML"
	@echo "  make build         - Compila os binários"
	@echo "  make clean         - Remove arquivos gerados"
	@echo "  make help          - Mostra esta ajuda"

# Testes
test:
	@echo "Executando testes..."
	go test ./... -v

test-coverage:
	@echo "Executando testes com cobertura..."
	go test ./... -coverprofile=$(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Relatório de cobertura gerado em $(COVERAGE_HTML)"

# Build
build:
	@echo "Compilando binários..."
	mkdir -p bin
	go build -o $(BINARY_DBBACKUP) cmd/dbbackup/main.go
	go build -o $(BINARY_UPLOADER) cmd/uploader/main.go
	@echo "Binários compilados em ./bin/"

# Limpeza
clean:
	@echo "Limpando arquivos gerados..."
	rm -rf bin/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Limpeza concluída"

# Comando padrão
.DEFAULT_GOAL := help 