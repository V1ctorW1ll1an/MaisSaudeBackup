# MaisSaudeBackup

![Go](https://img.shields.io/badge/Go-1.24.1-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Tests](https://img.shields.io/badge/tests-passing-green.svg)

Um sistema robusto de backup e monitoramento para bancos de dados SQL Server, com integração ao Google Drive e notificações via WhatsApp.

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Funcionalidades](#-funcionalidades)
- [Tecnologias Utilizadas](#-tecnologias-utilizadas)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Pré-requisitos](#-pré-requisitos)
- [Configuração](#-configuração)
- [Build](#-build)
- [Testes](#-testes)
- [Uso](#-uso)
- [Parâmetros CLI](#-parâmetros-cli)
- [Contribuição](#-contribuição)
- [Licença](#-licença)

## 🌟 Visão Geral

O MaisSaudeBackup é uma solução completa para backup e monitoramento de bancos de dados SQL Server. O sistema oferece:

- Backup automático de bancos de dados
- Upload seguro para o Google Drive
- Monitoramento em tempo real de alterações
- Notificações via WhatsApp
- Sistema de logs detalhado

## ✨ Funcionalidades

- **Backup Automático**: Realiza backups periódicos dos bancos de dados SQL Server
- **Upload para Google Drive**: Armazena os backups de forma segura na nuvem
- **Monitoramento**: Observa alterações nos bancos de dados em tempo real
- **Notificações**: Envia alertas via WhatsApp sobre o status dos backups
- **Logs Detalhados**: Mantém um registro completo de todas as operações

## 🛠️ Tecnologias Utilizadas

- **Go 1.24.1**: Linguagem de programação principal
- **SQL Server**: Banco de dados principal
- **Google Drive API**: Para armazenamento em nuvem
- **WhatsApp API**: Para notificações
- **OpenTelemetry**: Para monitoramento e métricas

## 📁 Estrutura do Projeto

```
MaisSaudeBackup/
├── cmd/
│   ├── dbbackup/     # Comandos para backup de banco de dados
│   └── uploader/     # Comandos para upload de arquivos
├── internal/
│   ├── config/       # Configurações do sistema
│   ├── gdrive/       # Integração com Google Drive
│   ├── logger/       # Sistema de logs
│   ├── watcher/      # Monitoramento de alterações
│   └── whatsapp/     # Integração com WhatsApp
├── backups/          # Diretório de backups locais
├── logs/             # Diretório de logs
├── go.mod            # Dependências do projeto
└── .env              # Variáveis de ambiente
```

## 📋 Pré-requisitos

- Go 1.24.1 ou superior
- SQL Server
- Conta Google com acesso à API do Google Drive
- Conta WhatsApp Business API
- Credenciais de acesso configuradas

## ⚙️ Configuração

1. Clone o repositório:

```bash
git clone https://github.com/V1ctorW1ll1an/MaisSaudeBackup.git
```

2. Instale as dependências:

```bash
go mod download
```

3. Configure as variáveis de ambiente no arquivo `.env`:

```env
DB_SERVER=seu_servidor
DB_USER=seu_usuario
DB_PASSWORD=sua_senha
GOOGLE_CREDENTIALS=path/to/credentials.json
WHATSAPP_TOKEN=seu_token
```

4. Configure as credenciais do Google Drive e WhatsApp conforme necessário.

## 🏗️ Build

Para construir o projeto, execute os seguintes comandos:

```bash
# Construir todos os binários
make build

# Ou construir individualmente
go build -o bin/dbbackup cmd/dbbackup/main.go
go build -o bin/uploader cmd/uploader/main.go
```

Os binários serão gerados no diretório `bin/`.

## 🧪 Testes

Para executar os testes do projeto:

```bash
# Executar todos os testes
make test

# Executar testes com cobertura
make test-coverage

# Executar testes específicos
go test ./internal/... -v
```

## 🚀 Uso

### Backup de Banco de Dados

```bash
# Usando o binário compilado
./bin/dbbackup [parâmetros]

# Ou executando diretamente
go run cmd/dbbackup/main.go [parâmetros]
```

### Upload para Google Drive

```bash
# Usando o binário compilado
./bin/uploader [parâmetros]

# Ou executando diretamente
go run cmd/uploader/main.go [parâmetros]
```

## 📝 Parâmetros CLI

### Backup de Banco de Dados (dbbackup)

```bash
./bin/dbbackup [opções]

Opções:
  -server string
        Servidor do banco de dados (padrão: localhost)
  -port int
        Porta do banco de dados (padrão: 1433)
  -user string
        Usuário do banco de dados
  -password string
        Senha do banco de dados
  -database string
        Nome do banco de dados
  -backup-path string
        Caminho para salvar o backup (padrão: ./backups)
  -schedule string
        Agendamento do backup no formato cron (ex: "0 0 * * *" para diário)
  -retention int
        Número de dias para manter os backups (padrão: 30)
  -compress
        Comprimir o backup (padrão: true)
  -verbose
        Modo verboso para logs detalhados
```

### Upload para Google Drive (uploader)

```bash
./bin/uploader [opções]

Opções:
  -source string
        Caminho do arquivo ou diretório para upload
  -destination string
        ID da pasta de destino no Google Drive
  -credentials string
        Caminho para o arquivo de credenciais do Google (padrão: credentials.json)
  -recursive
        Upload recursivo de diretórios
  -delete-source
        Deletar arquivo fonte após upload bem-sucedido
  -verbose
        Modo verboso para logs detalhados
```

### Exemplos de Uso

```bash
# Backup diário do banco de dados
./bin/dbbackup -server "meu-servidor" -user "admin" -password "senha123" -database "meu_banco" -schedule "0 0 * * *"

# Backup único com compressão
./bin/dbbackup -server "meu-servidor" -user "admin" -password "senha123" -database "meu_banco" -compress

# Upload de arquivo para o Google Drive
./bin/uploader -source "backups/meu_banco.bak" -destination "folder_id" -credentials "path/to/credentials.json"

# Upload recursivo de diretório
./bin/uploader -source "backups/" -destination "folder_id" -recursive
```

## 🔄 Processo de Desenvolvimento

1. Clone o repositório
2. Instale as dependências
3. Configure as variáveis de ambiente
4. Execute os testes
5. Faça o build do projeto
6. Execute os binários com os parâmetros necessários

## 🤝 Contribuição

Contribuições são bem-vindas! Para contribuir:

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 📞 Contato

Victor Willian - [@V1ctorW1ll1an](https://github.com/V1ctorW1ll1an)

---

<div align="center">
  <sub>Desenvolvido com ❤️ por Victor Willian</sub>
</div>
