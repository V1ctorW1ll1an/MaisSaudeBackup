package main

import (
	"archive/zip"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/config"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/logger"
	"github.com/V1ctorW1ll1an/MaisSaudeBackup/internal/whatsapp"
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

	// --- Inicializar WhatsApp Client ---
	whatsappClient, err := whatsapp.ConfigWhatsappApi()
	if err != nil {
		l.Error("Erro ao configurar cliente WhatsApp", slog.Any("error", err))
		// Não saímos aqui pois o backup ainda pode funcionar sem WhatsApp
	}

	// --- Construção da Connection String ---
	connString := fmt.Sprintf("sqlserver://%s:%s@%s?database=%s", cfg.User, cfg.Password, cfg.Server, cfg.Database)

	fmt.Println(connString)

	// --- Conectar ao Banco ---
	l.Info("Conectando ao servidor SQL Server...", slog.String("server", cfg.Server))
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		l.Error("Erro ao preparar conexão", slog.Any("error", err))
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao preparar conexão: %v", err))
		}
		os.Exit(1)
	}
	defer db.Close()

	// Verifica a conexão
	err = db.Ping()
	if err != nil {
		l.Error("Erro ao conectar ao banco de dados", slog.Any("error", err))
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao conectar ao banco de dados: %v", err))
		}
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
	// "BACKUP DATABASE [SCM] TO DISK = '%backup_dir%\SCM_%data_completa%_%horario%.bak'"
	backupSQL := fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = '%s\\%s_%s.bak'",
		cfg.Database, cfg.BackupDir, cfg.Database, timestamp)

	l.Info("Backup SQL", slog.String("sql", backupSQL))

	l.Info("Preparando para executar backup",
		slog.String("database", cfg.Database),
		slog.String("backup_path_on_server", bakFilePathOnServer))
	l.Debug("Comando SQL de Backup", slog.String("sql", backupSQL))

	// --- Executar Backup ---
	_, err = db.Exec(backupSQL)
	if err != nil {
		l.Error("Erro ao executar o comando de backup", slog.Any("error", err))
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao executar o comando de backup: %v", err))
		}
		os.Exit(1)
	}
	l.Info("Comando de backup executado com sucesso no servidor.")

	// --- Preparar Arquivo Zip ---
	finalZipFilename := fmt.Sprintf("%s_%s.zip", cfg.Database, timestamp)
	tempZipFilename := fmt.Sprintf("%s_%s.tmp", cfg.Database, timestamp) // Nome temporário
	finalZipPathLocal := filepath.Join(cfg.ZipDir, finalZipFilename)
	tempZipPathLocal := filepath.Join(cfg.ZipDir, tempZipFilename) // Caminho temporário

	l.Info("Criando arquivo zip temporário", slog.String("path", tempZipPathLocal))
	zipFile, err := os.Create(tempZipPathLocal) // Cria com nome .tmp
	if err != nil {
		l.Error("Erro ao criar arquivo zip temporário", slog.String("path", tempZipPathLocal), slog.Any("error", err))
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao criar arquivo zip temporário: %v", err))
		}
		os.Exit(1)
	}
	// Defer o fechamento ANTES do rename e ANTES do zipWriter.Close()
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	// Defer o fechamento do Writer ANTES do rename
	// defer zipWriter.Close() // Movido para depois da cópia

	// --- Abrir o Arquivo .bak (Acessando o path do servidor) ---
	l.Info("Abrindo arquivo de backup do servidor", slog.String("path", bakFilePathOnServer))
	bakFile, err := os.Open(bakFilePathOnServer)
	if err != nil {
		l.Error("Erro ao abrir arquivo .bak. Verifique o caminho e as permissões.",
			slog.String("path", bakFilePathOnServer),
			slog.Any("error", err))
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao abrir arquivo .bak: %v", err))
		}
		os.Exit(1)
	}
	defer bakFile.Close()

	// --- Criar Entrada no Zip e Copiar Dados ---
	l.Info("Adicionando arquivo ao zip", slog.String("filename_in_zip", bakFilename))
	zipEntryWriter, err := zipWriter.Create(bakFilename)
	if err != nil {
		l.Error("Erro ao criar entrada no zip", slog.Any("error", err))
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao criar entrada no zip: %v", err))
		}
		os.Exit(1)
	}

	l.Info("Copiando dados do backup para o arquivo zip...")
	bytesCopied, err := io.Copy(zipEntryWriter, bakFile)
	if err != nil {
		l.Error("Erro ao copiar dados do .bak para o .zip", slog.Any("error", err))
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao copiar dados do .bak para o .zip: %v", err))
		}
		os.Exit(1)
	}
	l.Info("Dados copiados para o arquivo zip", slog.Int64("bytes_copied", bytesCopied))

	// --- Fechar o Zip Writer (IMPORTANTE: Fecha antes de renomear) ---
	l.Debug("Fechando zip writer...")
	err = zipWriter.Close() // Fecha explicitamente para garantir que tudo foi escrito
	if err != nil {
		l.Error("Erro ao fechar o zip writer", slog.String("path", tempZipPathLocal), slog.Any("error", err))
		// Tenta remover o arquivo temporário incompleto
		_ = os.Remove(tempZipPathLocal)
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao finalizar arquivo zip: %v", err))
		}
		os.Exit(1)
	}
	l.Debug("Zip writer fechado.")

	// --- Fechar o arquivo .tmp (opcional aqui, mas boa prática) ---
	// O defer zipFile.Close() já faz isso, mas fechar explicitamente antes do rename pode ser mais claro
	zipFile.Close()

	// --- Renomear o Arquivo Temporário para Final ---
	l.Info("Renomeando arquivo temporário para final", slog.String("from", tempZipPathLocal), slog.String("to", finalZipPathLocal))
	err = os.Rename(tempZipPathLocal, finalZipPathLocal)
	if err != nil {
		l.Error("Erro ao renomear arquivo .tmp para .zip", slog.String("from", tempZipPathLocal), slog.String("to", finalZipPathLocal), slog.Any("error", err))
		// Tenta remover o arquivo temporário se a renomeação falhar
		_ = os.Remove(tempZipPathLocal)
		if whatsappClient != nil {
			whatsappClient.Send("Admin", cfg.Database, time.Now().Format("02/01/2006 15:04:05"), fmt.Sprintf("Erro ao renomear arquivo zip final: %v", err))
		}
		os.Exit(1)
	}

	// --- Finalização ---
	l.Info("Backup concluído e zipado com sucesso",
		slog.String("database", cfg.Database),
		slog.String("zip_file", finalZipPathLocal)) // Loga o nome final
}
