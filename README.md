<p align="center">
  <img src="internal/app/assets/icon.png" alt="Ícone do Look News" width="180" />
</p>

<h1 align="center">Look News — System Tray</h1>

<p align="center">
  Agregador de notícias RSS na barra de menus do macOS.
</p>

## Instalação

### Pelo Terminal

Execute este comando no Terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/ClemilsonAzevedo/go-tray/main/install.sh | bash
```

O instalador identifica a arquitetura do Mac, baixa o pacote correspondente da última [release](https://github.com/ClemilsonAzevedo/go-tray/releases), verifica o SHA-256 e instala o app em `Aplicativos`. Se não puder gravar em `/Applications`, instala em `~/Applications`. Não é necessário clonar o repositório nem compilar o app.

### Instalação manual

1. Abra a página de [Releases](https://github.com/ClemilsonAzevedo/go-tray/releases) e baixe o ZIP para a arquitetura do seu Mac.
2. Descompacte o arquivo e arraste `Look News.app` para a pasta `Aplicativos`.
3. Abra o app; o ícone aparecerá na barra de menus.

Como o app não é assinado nem notarizado, o macOS pode exibir um aviso na primeira abertura. Se isso acontecer, clique no app com o botão direito (ou `Control` + clique), escolha **Abrir** e confirme.

## Onde encontrar o app

O Look News fica na barra de menus, no canto superior direito. Clique no ícone para ver as notícias; use os controles da janela para atualizar, gerenciar fontes ou sair.

<p align="center">
  <img src="docs/assets/100b8.png" alt="Ilustração do Look News na barra de menus do macOS" width="760" />
</p>

## Configuração

O app usa a API de produção por padrão.

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `LOOK_NEWS_BACKEND_URL` | `https://look-news-api.onrender.com` | URL base da API de notícias. |

## Executar localmente

Requer Go e, no macOS, as ferramentas de linha de comando do Xcode.

```bash
go mod download
go run ./cmd/look-news
```

## Gerar o pacote para macOS

Instale a CLI do Fyne e execute:

```bash
go install fyne.io/tools/cmd/fyne@latest
cd cmd/look-news
fyne package -release
```

O bundle `Look News.app` será gerado no diretório atual.