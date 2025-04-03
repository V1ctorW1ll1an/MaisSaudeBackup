package gdrive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// getOAuthClient obtém um http.Client autenticado, gerenciando o token.
// (função não exportada, auxiliar para NewDriveUploader)
func getOAuthClient(ctx context.Context, logger *slog.Logger, config *oauth2.Config, tokenFile string) (*http.Client, error) {
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		logger.Info("Token não encontrado ou inválido, iniciando fluxo de autorização web.", slog.String("token_file", tokenFile), slog.Any("error", err))
		tok, err = getTokenFromWeb(ctx, logger, config) // Não passa tokenFile aqui
		if err != nil {
			logger.Error("Falha ao obter token via web", slog.Any("error", err))
			return nil, fmt.Errorf("obtenção de token via web falhou: %w", err)
		}
		err = saveToken(logger, tokenFile, tok)
		if err != nil {
			logger.Warn("Falha ao salvar o novo token no arquivo", slog.String("path", tokenFile), slog.Any("error", err))
			// Continua mesmo assim, pois temos o token em memória
		}
	} else {
		logger.Info("Token carregado do arquivo com sucesso.", slog.String("token_file", tokenFile))
	}
	return config.Client(ctx, tok), nil
}

// getTokenFromWeb realiza o fluxo OAuth2 via navegador para obter um novo token.
// (função não exportada)
func getTokenFromWeb(ctx context.Context, logger *slog.Logger, config *oauth2.Config) (*oauth2.Token, error) {
	config.RedirectURL = "http://localhost:8989/callback" // Hardcoded por simplicidade, poderia ser configurável
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Abra este link no seu navegador para autorizar a aplicação:\n%v\n", authURL)
	logger.Info("Aguardando autorização do usuário via navegador...", slog.String("url", authURL))

	authCodeCh := make(chan string)
	errCh := make(chan error)

	server := &http.Server{Addr: ":8989"}

	mux := http.NewServeMux() // Usa um mux dedicado
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			logger.Error("Callback recebido sem código de autorização")
			http.Error(w, "Erro: Código de autorização não encontrado na URL.", http.StatusBadRequest)
			errCh <- errors.New("callback sem código de autorização")
			return
		}
		logger.Info("Código de autorização recebido via callback.")
		fmt.Fprintln(w, "Autorização recebida! Você pode fechar esta janela.")
		authCodeCh <- code
	})
	server.Handler = mux // Associa o mux ao servidor

	go func() {
		logger.Info("Servidor de callback OAuth2 iniciado em :8989")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Erro ao iniciar servidor de callback", slog.Any("error", err))
			errCh <- fmt.Errorf("servidor de callback falhou: %w", err)
		}
		logger.Debug("Servidor de callback HTTP encerrado.")
	}()

	defer func() {
		logger.Debug("Solicitando shutdown do servidor de callback...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Erro durante shutdown do servidor de callback", slog.Any("error", err))
		} else {
			logger.Info("Servidor de callback desligado com sucesso.")
		}
	}()

	select {
	case code := <-authCodeCh:
		logger.Info("Trocando código de autorização por token...")
		tok, err := config.Exchange(ctx, code)
		if err != nil {
			logger.Error("Falha ao trocar código por token", slog.Any("error", err))
			return nil, fmt.Errorf("troca de código falhou: %w", err)
		}
		logger.Info("Token obtido com sucesso.")
		return tok, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		logger.Warn("Obtenção de token cancelada devido ao contexto principal.")
		return nil, ctx.Err()
	}
}

// tokenFromFile carrega um token OAuth2 de um arquivo JSON.
// (função não exportada)
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("abrir arquivo %s falhou: %w", file, err)
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	if err != nil {
		return nil, fmt.Errorf("decodificar token de %s falhou: %w", file, err)
	}
	return tok, nil
}

// saveToken salva um token OAuth2 em um arquivo JSON.
// (função não exportada)
func saveToken(logger *slog.Logger, path string, token *oauth2.Token) error {
	logger.Info("Salvando token de credenciais", slog.String("path", path))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logger.Error("Não foi possível abrir/criar arquivo para salvar token", slog.String("path", path), slog.Any("error", err))
		return fmt.Errorf("abrir/criar %s para salvar token falhou: %w", path, err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		logger.Error("Não foi possível encodar token para o arquivo", slog.String("path", path), slog.Any("error", err))
		return fmt.Errorf("encodar token para %s falhou: %w", path, err)
	}
	return nil
}
