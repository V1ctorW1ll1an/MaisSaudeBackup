package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/config"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/logger"
	_ "github.com/denisenkom/go-mssqldb" // Driver SQL Server (import anônimo)
)

func main() {
	cfg, err := config.NewDBBackupConfig()
	if err != nil {
		log.Printf("Erro ao carregar a configuração: %v", err)
		flag.Usage()
		os.Exit(1)
	}

	flag.Parse()

	l, logFile, err := logger.Setup(cfg.LogDir, cfg.LogLevel)
	if err != nil {
		log.Fatalf("Erro ao configurar o logger: %v", err)
	}
	defer logFile.Close()

	// --- Validação Simples dos Flags ---
	config.ValidateBackupFlags(cfg)

	// --- Construção da Connection String ---
	connString := fmt.Sprintf("sqlserver://%s:%s@%s?database=%s", cfg.User, cfg.Password, cfg.Server, cfg.Database)

	fmt.Println(connString)

	// --- Conectar ao Banco ---
	l.Info("Conectando ao servidor SQL Server...", slog.String("server", cfg.Server))
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		l.Error("Erro ao preparar conexão", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// Verifica a conexão
	err = db.Ping()
	if err != nil {
		l.Error("Erro ao conectar ao banco de dados", slog.Any("error", err))
		os.Exit(1)
	}
	l.Info("Conexão estabelecida com sucesso.")

	// --- Preparar Comando de Backup ---
	timestamp := time.Now().Format("20060102_150405") // Formato YYYYMMDD_HHMMSS
	bakFilename := fmt.Sprintf("%s_%s.bak", cfg.Database, timestamp)
	// IMPORTANTE: Este path é no *servidor SQL Server*
	bakFilePathOnServer := filepath.Join(cfg.BackupDir, bakFilename)
	// Substitui barras para o formato Windows, caso Join use barra normal
	bakFilePathOnServer = filepath.ToSlash(bakFilePathOnServer)

	backupSQL := fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = N'%s' WITH FORMAT, NAME = N'%s Backup'",
		cfg.Database, bakFilePathOnServer, cfg.Database)

	l.Info("Preparando para executar backup",
		slog.String("database", cfg.Database),
		slog.String("backup_path_on_server", bakFilePathOnServer))
	l.Debug("Comando SQL de Backup", slog.String("sql", backupSQL))

	// --- Executar Backup ---
	_, err = db.Exec(backupSQL)
	if err != nil {
		l.Error("Erro ao executar o comando de backup", slog.Any("error", err))
		os.Exit(1)
	}
	l.Info("Comando de backup executado com sucesso no servidor.")

	// --- Preparar Arquivo Zip ---
	// zipFilename := fmt.Sprintf("%s_%s.zip", cfg.Database, timestamp)
	// zipFilePathLocal := filepath.Join(cfg.ZipDir, zipFilename)

	// l.Info("Criando arquivo zip local", slog.String("path", zipFilePathLocal))
	// zipFile, err := os.Create(zipFilePathLocal)
	// if err != nil {
	// 	l.Error("Erro ao criar arquivo zip", slog.String("path", zipFilePathLocal), slog.Any("error", err))
	// }
	// defer zipFile.Close()

	// zipWriter := zip.NewWriter(zipFile)
	// defer zipWriter.Close()

	// // --- Abrir o Arquivo .bak (Acessando o path do servidor) ---
	// l.Info("Abrindo arquivo de backup do servidor", slog.String("path", bakFilePathOnServer))
	// bakFile, err := os.Open(bakFilePathOnServer)
	// if err != nil {
	// 	l.Error("Erro ao abrir arquivo .bak. Verifique o caminho e as permissões.",
	// 		slog.String("path", bakFilePathOnServer),
	// 		slog.Any("error", err))
	// 	os.Exit(1)
	// }
	// defer bakFile.Close()

	// // --- Criar Entrada no Zip e Copiar Dados ---
	// l.Info("Adicionando arquivo ao zip", slog.String("filename_in_zip", bakFilename))
	// zipEntryWriter, err := zipWriter.Create(bakFilename)
	// if err != nil {
	// 	l.Error("Erro ao criar entrada no zip", slog.Any("error", err))
	// 	os.Exit(1)
	// }

	// l.Info("Copiando dados do backup para o arquivo zip...")
	// bytesCopied, err := io.Copy(zipEntryWriter, bakFile)
	// if err != nil {
	// 	l.Error("Erro ao copiar dados do .bak para o .zip", slog.Any("error", err))
	// 	os.Exit(1)
	// }
	// l.Info("Dados copiados para o arquivo zip", slog.Int64("bytes_copied", bytesCopied))

	// // --- Finalização ---
	// l.Info("Backup concluído e zipado com sucesso",
	// 	slog.String("database", cfg.Database),
	// 	slog.String("zip_file", zipFilePathLocal))
}
