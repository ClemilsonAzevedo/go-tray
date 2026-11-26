.DEFAULT_GOAL := help

APP_NAME       ?= look-news
MAIN_PACKAGE   ?= ./cmd/look-news
BIN_DIR        ?= bin
BIN            ?= $(BIN_DIR)/$(APP_NAME)
TMP_DIR        ?= tmp
COVERAGE_FILE  ?= coverage.out
ICON           ?= internal/app/assets/icon.png
AIR            ?= air
AIR_CONFIG     ?= .air.toml
FYNE           ?= fyne

.PHONY: help setup tools install-air install-fyne require-fyne run run-once build build-release \
	test test-race cover cover-html vet fmt tidy tidy-check check verify \
	package package-darwin package-linux package-windows package-all install-app clean

##@ Look News · comandos

help: ## Mostra esta lista de comandos
	@awk 'BEGIN {FS = ":.*## "; printf "\nLook News — comandos disponíveis\n\nUso: make \033[36m<comando>\033[0m\n"} \
		/^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5)} \
		/^[a-zA-Z0-9_-]+:.*## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo

setup: tools ## Prepara dependências Go e ferramentas de desenvolvimento
	go mod download

tools: install-air install-fyne ## Instala Air (hot reload) e a CLI do Fyne

install-air: ## Instala a ferramenta Air para desenvolvimento com hot reload
	go install github.com/air-verse/air@latest

install-fyne: ## Instala a CLI usada para empacotar aplicações Fyne
	go install fyne.io/tools/cmd/fyne@latest

##@ Desenvolvimento

run: ## Executa o Look News com hot reload (requer Air; use make install-air)
	@command -v "$(AIR)" >/dev/null 2>&1 || { printf 'Air não encontrado. Execute: make install-air\n' >&2; exit 1; }
	$(AIR) -c $(AIR_CONFIG)

run-once: ## Executa uma vez, diretamente com go run
	go run $(MAIN_PACKAGE)

build: ## Compila o binário local em $(BIN)
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(MAIN_PACKAGE)

build-release: ## Compila um binário otimizado e sem símbolos de depuração
	mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags="-s -w" -o $(BIN) $(MAIN_PACKAGE)

##@ Qualidade

test: ## Executa os testes de todos os pacotes
	go test ./...

test-race: ## Executa os testes com o detector de condições de corrida
	go test -race ./...

cover: ## Executa os testes com cobertura e mostra o resumo
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE) | tail -n 1

cover-html: cover ## Gera e abre o relatório HTML de cobertura
	mkdir -p $(TMP_DIR)
	go tool cover -html=$(COVERAGE_FILE) -o $(TMP_DIR)/coverage.html
	open $(TMP_DIR)/coverage.html

vet: ## Analisa os pacotes com go vet
	go vet ./...

fmt: ## Formata o código Go
	go fmt ./...

tidy: ## Sincroniza go.mod e go.sum com as dependências usadas
	go mod tidy

tidy-check: ## Confere se go.mod e go.sum já estão atualizados, sem alterá-los
	go mod tidy -diff

check: fmt vet test ## Formata, analisa e testa o projeto

verify: tidy-check vet test-race ## Validação completa antes de compartilhar ou publicar
	go build ./...

##@ Empacotamento (Fyne)

require-fyne:
	@command -v "$(FYNE)" >/dev/null 2>&1 || { printf 'CLI do Fyne não encontrada. Execute: make install-fyne\n' >&2; exit 1; }

package: require-fyne ## Empacota para o sistema atual (personalize com ICON=caminho/icone.png)
	cd $(MAIN_PACKAGE) && $(FYNE) package -icon "$(abspath $(ICON))" -release

package-darwin: require-fyne ## Gera o pacote para macOS
	cd $(MAIN_PACKAGE) && $(FYNE) package -os darwin -icon "$(abspath $(ICON))" -release

package-linux: require-fyne ## Gera o pacote para Linux (requer toolchain compatível)
	cd $(MAIN_PACKAGE) && $(FYNE) package -os linux -icon "$(abspath $(ICON))" -release

package-windows: require-fyne ## Gera o pacote para Windows (requer toolchain compatível)
	cd $(MAIN_PACKAGE) && $(FYNE) package -os windows -icon "$(abspath $(ICON))" -release

package-all: package-darwin package-linux package-windows ## Tenta empacotar para macOS, Linux e Windows

install-app: require-fyne ## Instala o Look News no sistema atual com a CLI do Fyne
	cd $(MAIN_PACKAGE) && $(FYNE) install -icon "$(abspath $(ICON))"

##@ Limpeza

clean: ## Remove binários e relatórios gerados pelo projeto
	rm -f $(BIN) $(TMP_DIR)/$(APP_NAME) $(COVERAGE_FILE) $(TMP_DIR)/air-build-errors.log $(TMP_DIR)/coverage.html
	rm -rf "Look News.app"
	rm -f $(APP_NAME).exe $(APP_NAME).tar.gz