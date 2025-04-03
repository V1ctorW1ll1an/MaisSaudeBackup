package config

import (
	"errors"
	"flag"
)

// Config armazena as configurações da aplicação carregadas via flags.
type Config struct {
	WatchDir        string
	LogDir          string
	CredentialsFile string
	TokenFile       string
	LogLevel        string // e.g., "debug", "info", "warn", "error"
}

// NewConfig define os flags de configuração da aplicação, lê seus valores
// (que devem ter sido previamente parseados por uma chamada a flag.Parse() na main)
// e retorna uma nova instância de Config preenchida.
//
// Flags definidos:
//
//	-watch-dir: Diretório a ser monitorado (obrigatório).
//	-log-dir: Diretório para logs.
//	-credentials-file: Caminho para o arquivo de credenciais OAuth2 (obrigatório).
//	-token-file: Caminho para o arquivo de token OAuth2 (obrigatório).
//	-log-level: Nível de log (debug, info, warn, error).
//
// Retorna um ponteiro para a struct Config preenchida e um erro se os valores
// dos flags obrigatórios (após o parse) estiverem vazios.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	// Define os flags e associa às variáveis da struct cfg.
	// Os valores reais serão preenchidos por flag.Parse() na main.
	flag.StringVar(&cfg.WatchDir, "watch-dir", "./backups", "Diretório a ser monitorado para novos arquivos.")
	flag.StringVar(&cfg.LogDir, "log-dir", "./logs", "Diretório para armazenar arquivos de log.")
	flag.StringVar(&cfg.CredentialsFile, "credentials-file", "credentials.json", "Caminho para o arquivo credentials.json do Google OAuth2.")
	flag.StringVar(&cfg.TokenFile, "token-file", "token.json", "Caminho para salvar/carregar o token OAuth2 do usuário.")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Nível de log (debug, info, warn, error).")

	// Validação dos valores OBTIDOS após o Parse
	// Verifica se os flags obrigatórios receberam algum valor.
	if cfg.WatchDir == "" || cfg.CredentialsFile == "" || cfg.TokenFile == "" {
		return nil, errors.New("os flags -watch-dir, -credentials-file, e -token-file são obrigatórios e não podem estar vazios")
	}

	return cfg, nil
}
