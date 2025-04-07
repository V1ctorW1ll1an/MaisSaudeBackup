package config

import (
	"flag"
	"log"
	"os"
)

// UpdloaderConfig armazena as configurações da aplicação carregadas via flags.
type UpdloaderConfig struct {
	WatchDir        string
	LogDir          string
	CredentialsFile string
	TokenFile       string
	LogLevel        string // e.g., "debug", "info", "warn", "error"
}

type DbBackupConfig struct {
	Server    string
	Database  string
	User      string
	Password  string
	BackupDir string // Diretório de backup no servidor SQL
	ZipDir    string // Diretório local para salvar o zip
	LogDir    string // Diretório para os logs do dbbackup
	LogLevel  string // Nível de log (debug, info, warn, error)
}

// NewUploaderConfig define os flags de configuração da aplicação, lê seus valores
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
func NewUploaderConfig() (*UpdloaderConfig, error) {
	cfg := &UpdloaderConfig{}

	// Define os flags e associa às variáveis da struct cfg.
	// Os valores reais serão preenchidos por flag.Parse() na main.
	flag.StringVar(&cfg.WatchDir, "watch-dir", "", "Diretório a ser monitorado para novos arquivos.")
	flag.StringVar(&cfg.LogDir, "log-dir", "", "Diretório para armazenar arquivos de log.")
	flag.StringVar(&cfg.CredentialsFile, "credentials-file", "credentials.json", "Caminho para o arquivo credentials.json do Google OAuth2.")
	flag.StringVar(&cfg.TokenFile, "token-file", "token.json", "Caminho para salvar/carregar o token OAuth2 do usuário.")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Nível de log (debug, info, warn, error).")

	return cfg, nil
}

func NewDBBackupConfig() (*DbBackupConfig, error) {
	cfg := &DbBackupConfig{}

	flag.StringVar(&cfg.Server, "server", "", "Endereço do servidor SQL Server (ex: host\\instância ou host,porta)")
	flag.StringVar(&cfg.Database, "database", "", "Nome do banco de dados para backup")
	flag.StringVar(&cfg.User, "user", "", "Usuário do SQL Server (necessário se não usar Windows Auth)")
	flag.StringVar(&cfg.Password, "password", "", "Senha do SQL Server (necessário se não usar Windows Auth)")
	flag.StringVar(&cfg.BackupDir, "backup-dir", "", "Diretório NO SERVIDOR SQL SERVER onde o .bak será salvo (ex: C:\\Backups)")
	flag.StringVar(&cfg.ZipDir, "zip-dir", ".", "Diretório local onde o arquivo .zip final será salvo")
	flag.StringVar(&cfg.LogDir, "log-dir", "./logs", "Diretório para armazenar arquivos de log.")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Nível de log (debug, info, warn, error).")

	return cfg, nil
}

// validateFlags realiza validações adicionais nos flags de configuração
func ValidateBackupFlags(cfg *DbBackupConfig) {
	// Validação de campos obrigatórios
	if cfg.Server == "" {
		log.Fatal("Flag -server é obrigatório")
	}
	if cfg.Database == "" {
		log.Fatal("Flag -database é obrigatório")
	}
	if cfg.User == "" {
		log.Fatal("Flag -user é obrigatório")
	}
	if cfg.Password == "" {
		log.Fatal("Flag -password é obrigatório")
	}
	if cfg.BackupDir == "" {
		log.Fatal("Flag -backup-dir é obrigatório")
	}
	if cfg.ZipDir == "" {
		log.Fatal("Flag -zip-dir é obrigatório")
	}
	if cfg.LogDir == "" {
		log.Fatal("Flag -log-dir é obrigatório")
	}

	// Validação de diretórios
	if _, err := os.Stat(cfg.ZipDir); os.IsNotExist(err) {
		log.Printf("Aviso: O diretório -zip-dir '%s' não existe. Será criado se necessário.", cfg.ZipDir)
	}

	// Validação de nível de log
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[cfg.LogLevel] {
		log.Printf("Aviso: Nível de log '%s' inválido. Usando 'info' como padrão.", cfg.LogLevel)
		cfg.LogLevel = "info"
	}
}

func ValidateUploaderFlags(cfg *UpdloaderConfig) {
	if cfg.WatchDir == "" {
		log.Fatal("Flag -watch-dir é obrigatório")
	}
	if cfg.LogDir == "" {
		log.Fatal("Flag -log-dir é obrigatório")
	}
	if cfg.CredentialsFile == "" {
		log.Fatal("Flag -credentials-file é obrigatório")
	}
	if cfg.TokenFile == "" {
		log.Fatal("Flag -token-file é obrigatório")
	}
	if cfg.LogLevel == "" {
		log.Fatal("Flag -log-level é obrigatório")
	}

}
