#!/usr/bin/env bash
set -euo pipefail

REPO="ClemilsonAzevedo/go-tray"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

fail() {
  printf '  ✗ %s\n' "$1" >&2
  exit 1
}

printf '══════════════════════════════════════\n'
printf '  Look News — Instalador para macOS\n'
printf '══════════════════════════════════════\n\n'

case "$(uname -m)" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64) ARCH="x86_64" ;;
  *) fail "Arquitetura não suportada: $(uname -m)." ;;
esac

printf '[1/4] Procurando a versão mais recente para %s...\n' "$ARCH"
if ! RELEASE_JSON=$(curl -fsSL "$API_URL"); then
  fail "Não foi possível consultar a última release de https://github.com/${REPO}/releases."
fi

ASSET_URLS=$(printf '%s\n' "$RELEASE_JSON" | grep -oE '"browser_download_url"[[:space:]]*:[[:space:]]*"[^"]+"' || true)
ZIP_URL=$(printf '%s\n' "$ASSET_URLS" \
  | grep -E "/Look-News-v[0-9]{4}\.[0-9]{2}\.[0-9]{2}-macOS-${ARCH}\.zip\"$" \
  | sed -E 's/.*"([^"]+)".*/\1/' \
  | head -n 1 || true)
CHECKSUM_URL=$(printf '%s\n' "$ASSET_URLS" \
  | grep -E '/SHA256SUMS\.txt"$' \
  | sed -E 's/.*"([^"]+)".*/\1/' \
  | head -n 1 || true)

if [[ -z "$ZIP_URL" ]]; then
  fail "Não há pacote macOS para ${ARCH} na última release. Confira https://github.com/${REPO}/releases."
fi
if [[ -z "$CHECKSUM_URL" ]]; then
  fail "A última release não contém o arquivo SHA256SUMS.txt."
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
ARCHIVE_NAME=${ZIP_URL##*/}

printf '[2/4] Baixando o app e o checksum...\n'
curl --fail --location --progress-bar --output "$TMP_DIR/$ARCHIVE_NAME" "$ZIP_URL"
curl --fail --location --silent --show-error --output "$TMP_DIR/SHA256SUMS.txt" "$CHECKSUM_URL"

printf '[3/4] Verificando o arquivo baixado...\n'
if ! (cd "$TMP_DIR" && shasum -a 256 -c SHA256SUMS.txt); then
  fail "A verificação SHA-256 falhou; o app não foi instalado."
fi

printf '[4/4] Instalando o app...\n'
unzip -q "$TMP_DIR/$ARCHIVE_NAME" -d "$TMP_DIR"
APP_PATH="$TMP_DIR/Look News.app"
if [[ ! -d "$APP_PATH" ]]; then
  fail "Não foi encontrado Look News.app no arquivo baixado."
fi

if [[ -w /Applications ]]; then
  INSTALL_DIR="/Applications"
else
  INSTALL_DIR="$HOME/Applications"
  mkdir -p "$INSTALL_DIR"
  printf '  /Applications não está disponível para escrita; usando %s.\n' "$INSTALL_DIR"
fi

DEST_APP="$INSTALL_DIR/Look News.app"
if [[ -e "$DEST_APP" ]]; then
  rm -rf "$DEST_APP"
fi
ditto "$APP_PATH" "$DEST_APP"

printf '\n══════════════════════════════════════\n'
printf '  Instalação concluída!\n'
printf '  Abra Look News pelo Launchpad ou Spotlight.\n'
printf '  Local: %s\n' "$DEST_APP"
printf '══════════════════════════════════════\n'