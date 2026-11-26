#!/usr/bin/env bash
set -Eeuo pipefail

REPO_URL="https://github.com/ClemilsonAzevedo/go-tray.git"
APP_NAME="Look News"
BUNDLE_ID="com.clema.looknews"
MIN_GO_VERSION="1.26.0"
SYSTEM_APP="/Applications/${APP_NAME}.app"
USER_APP="${HOME}/Applications/${APP_NAME}.app"
WORK_DIR=""
STAGE_DIR=""
DEST_APP=""
KEEP_STAGE=0
MISSING_PREREQUISITES=0

fail() {
  printf '  ✗ %s\n' "$1" >&2
  exit 1
}

cleanup() {
  if [[ -n "$STAGE_DIR" && -d "$STAGE_DIR" ]]; then
    if [[ -n "$DEST_APP" && -e "$STAGE_DIR/Previous ${APP_NAME}.app" && ! -e "$DEST_APP" ]]; then
      KEEP_STAGE=1
      printf '  ⚠ A cópia anterior foi preservada em %s\n' "$STAGE_DIR/Previous ${APP_NAME}.app" >&2
    fi
    if [[ "$KEEP_STAGE" -eq 0 ]]; then
      rm -rf "$STAGE_DIR"
    fi
  fi
  if [[ -n "$WORK_DIR" && -d "$WORK_DIR" ]]; then
    rm -rf "$WORK_DIR"
  fi
}

installed_version() {
  local app_path="$1"
  if [[ -f "$app_path/Contents/Info.plist" ]]; then
    /usr/libexec/PlistBuddy -c 'Print :CFBundleShortVersionString' "$app_path/Contents/Info.plist" 2>/dev/null || printf 'desconhecida'
  else
    printf 'desconhecida'
  fi
}

report_installation() {
  local app_path="$1"
  if [[ -e "$app_path" || -L "$app_path" ]]; then
    printf '  Encontrado: %s (versão %s)\n' "$app_path" "$(installed_version "$app_path")"
  fi
}

printf '══════════════════════════════════════\n'
printf '  Look News — Instalador para macOS\n'
printf '══════════════════════════════════════\n\n'

[[ "$(uname -s)" == "Darwin" ]] || fail 'Este instalador funciona somente no macOS.'

printf '[1/5] Verificando instalação e pré-requisitos...\n'

XCODE_TOOLS_READY=0
if command -v xcode-select >/dev/null 2>&1 && xcode-select -p >/dev/null 2>&1; then
  if command -v xcrun >/dev/null 2>&1 && xcrun --find clang >/dev/null 2>&1; then
    XCODE_TOOLS_READY=1
  else
    printf '  ✗ O compilador Clang das ferramentas do Xcode não está acessível.\n' >&2
    printf '    Instale ou repare as ferramentas com: xcode-select --install\n' >&2
    MISSING_PREREQUISITES=1
  fi
else
  printf '  ✗ As ferramentas de linha de comando do Xcode não estão instaladas/configuradas.\n' >&2
  printf '    Instale-as com: xcode-select --install\n' >&2
  MISSING_PREREQUISITES=1
fi

if command -v git >/dev/null 2>&1; then
  if [[ "$XCODE_TOOLS_READY" -eq 1 ]] && ! git --version >/dev/null 2>&1; then
    printf '  ✗ Git foi encontrado, mas não pode ser executado. Reinstale-o ou instale-o com: brew install git\n' >&2
    MISSING_PREREQUISITES=1
  fi
else
  printf '  ✗ Git não encontrado. Instale as ferramentas do Xcode ou, com Homebrew, execute: brew install git\n' >&2
  MISSING_PREREQUISITES=1
fi

if command -v go >/dev/null 2>&1; then
  GO_VERSION_OUTPUT=$(go version 2>/dev/null || true)
  if [[ "$GO_VERSION_OUTPUT" =~ go([0-9]+)\.([0-9]+)(\.([0-9]+))? ]]; then
    GO_VERSION_NUMBER="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.${BASH_REMATCH[4]:-0}"
    GO_MAJOR=${BASH_REMATCH[1]}
    GO_MINOR=${BASH_REMATCH[2]}
    GO_PATCH=${BASH_REMATCH[4]:-0}
    if (( GO_MAJOR < 1 || (GO_MAJOR == 1 && GO_MINOR < 26) )); then
      printf '  ✗ Go %s encontrado; o projeto requer Go %s ou posterior.\n' "$GO_VERSION_NUMBER" "$MIN_GO_VERSION" >&2
      printf '    Atualize o Go em: https://go.dev/dl/\n' >&2
      MISSING_PREREQUISITES=1
    fi
  else
    printf '  ✗ Não foi possível identificar a versão do Go. Atualize-o em: https://go.dev/dl/\n' >&2
    MISSING_PREREQUISITES=1
  fi
else
  printf '  ✗ Go não encontrado. Instale Go %s ou posterior em: https://go.dev/dl/\n' "$MIN_GO_VERSION" >&2
  MISSING_PREREQUISITES=1
fi

case "$(uname -m)" in
  arm64|x86_64) ;;
  *)
    printf '  ✗ Arquitetura não suportada: %s.\n' "$(uname -m)" >&2
    MISSING_PREREQUISITES=1
    ;;
esac

if [[ "$MISSING_PREREQUISITES" -ne 0 ]]; then
  printf '\nInstale ou atualize os itens indicados e execute este instalador novamente. Nenhuma instalação foi iniciada.\n' >&2
  exit 1
fi

report_installation "$SYSTEM_APP"
report_installation "$USER_APP"
printf '  Go: %s\n' "$GO_VERSION_OUTPUT"
printf '  Arquitetura: %s\n' "$(uname -m)"

if [[ -d "$SYSTEM_APP" && -w /Applications ]]; then
  INSTALL_DIR="/Applications"
elif [[ -d "$USER_APP" && -w "$HOME/Applications" ]]; then
  INSTALL_DIR="$HOME/Applications"
elif [[ -w /Applications ]]; then
  INSTALL_DIR="/Applications"
else
  INSTALL_DIR="$HOME/Applications"
  mkdir -p "$INSTALL_DIR"
  printf '  Sem permissão para gravar em /Applications; usando %s.\n' "$INSTALL_DIR"
fi
DEST_APP="$INSTALL_DIR/${APP_NAME}.app"

WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/look-news-install.XXXXXX")
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

printf '\n[2/5] Clonando o repositório completo (branch main)...\n'
GIT_TERMINAL_PROMPT=0 git clone --progress --branch main "$REPO_URL" "$WORK_DIR/source"
SOURCE_DIR="$WORK_DIR/source"
APP_CONFIG="$SOURCE_DIR/cmd/look-news/FyneApp.toml"
[[ -f "$APP_CONFIG" ]] || fail 'A configuração do app não foi encontrada no repositório clonado.'
SOURCE_VERSION=$(awk -F '"' '/^[[:space:]]*Version[[:space:]]*=/ { print $2; exit }' "$APP_CONFIG")
[[ -n "$SOURCE_VERSION" ]] || fail 'Não foi possível identificar a versão do app no FyneApp.toml.'
printf '  Versão do código-fonte: %s\n' "$SOURCE_VERSION"

printf '\n[3/5] Compilando o app para macOS...\n'
mkdir -p "$WORK_DIR/bin"
GOBIN="$WORK_DIR/bin" go install fyne.io/tools/cmd/fyne@latest
(cd "$SOURCE_DIR/cmd/look-news" && "$WORK_DIR/bin/fyne" package -os darwin -release)
BUILT_APP="$SOURCE_DIR/cmd/look-news/${APP_NAME}.app"
[[ -d "$BUILT_APP/Contents" ]] || fail 'A compilação terminou sem gerar o bundle Look News.app.'
BUILT_VERSION=$(installed_version "$BUILT_APP")
[[ "$BUILT_VERSION" == "$SOURCE_VERSION" ]] || fail "A versão do bundle gerado ($BUILT_VERSION) não corresponde ao código-fonte ($SOURCE_VERSION)."
EXECUTABLE=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleExecutable' "$BUILT_APP/Contents/Info.plist" 2>/dev/null) || fail 'Não foi possível localizar o executável no bundle gerado.'
[[ -x "$BUILT_APP/Contents/MacOS/$EXECUTABLE" ]] || fail 'O executável não foi encontrado dentro do bundle gerado.'

printf '\n[4/5] Instalando em %s...\n' "$INSTALL_DIR"
STAGE_DIR=$(mktemp -d "$INSTALL_DIR/.look-news-install.XXXXXX") || fail "Não foi possível preparar a instalação em $INSTALL_DIR."
STAGED_APP="$STAGE_DIR/${APP_NAME}.app"
BACKUP_APP="$STAGE_DIR/Previous ${APP_NAME}.app"
ditto "$BUILT_APP" "$STAGED_APP"

if /usr/bin/pgrep -x "$APP_NAME" >/dev/null 2>&1; then
  /usr/bin/osascript -e "tell application id \"$BUNDLE_ID\" to quit" || fail 'Feche o Look News e tente novamente.'
  for _ in {1..20}; do
    /usr/bin/pgrep -x "$APP_NAME" >/dev/null 2>&1 || break
    sleep 0.25
  done
  /usr/bin/pgrep -x "$APP_NAME" >/dev/null 2>&1 && fail 'O Look News ainda está aberto; feche o app e tente novamente.'
fi

if [[ -e "$DEST_APP" || -L "$DEST_APP" ]]; then
  mv "$DEST_APP" "$BACKUP_APP" || fail "Não foi possível substituir a instalação existente em $DEST_APP."
  if ! mv "$STAGED_APP" "$DEST_APP"; then
    if mv "$BACKUP_APP" "$DEST_APP"; then
      fail 'A nova versão não pôde ser instalada; a versão anterior foi restaurada.'
    fi
    KEEP_STAGE=1
    fail "A instalação falhou e a cópia anterior foi preservada em $BACKUP_APP."
  fi
else
  mv "$STAGED_APP" "$DEST_APP" || fail "Não foi possível instalar o app em $DEST_APP."
fi

printf '\n[5/5] Abrindo o Look News...\n'
if ! open "$DEST_APP"; then
  printf '  ⚠ O app foi instalado, mas não abriu automaticamente. Abra: %s\n' "$DEST_APP" >&2
  exit 1
fi

printf '\n══════════════════════════════════════\n'
printf '  Instalação concluída! Versão %s\n' "$BUILT_VERSION"
printf '  Confira o ícone do Look News na barra de menus.\n'
printf '  Local: %s\n' "$DEST_APP"
printf '══════════════════════════════════════\n'