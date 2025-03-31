package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var (
	developersGoogleDocs = "https://developers.google.com/workspace/drive/api/quickstart/go?hl=pt-br"
	logger               *Logger
)

func getClient(config *oauth2.Config) *http.Client {
	tf := "token.json"
	t, err := tokenFromFile(tf)
	if err != nil {
		t = getTokenFromWeb(config)
		saveToken(tf, t)
	}
	return config.Client(context.Background(), t)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Cria um canal para receber o código de autorização
	authCodeCh := make(chan string)

	// Configura um handler para capturar o código de autorização
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		authCodeCh <- authCode

		fmt.Fprintf(w, "Autenticação concluída. Você pode fechar esta janela.")
	})

	// Configura a URL de autorização com redirect_uri
	config.RedirectURL = "http://localhost:8989/callback"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// Inicia o servidor local
	go func() {
		fmt.Println("Abra este link no seu navegador:")
		fmt.Println(authURL)
		logger.Info("Servidor de autenticação iniciado na porta 8989")
		logger.Error(http.ListenAndServe(":8989", nil))
	}()

	// Espera pelo código de autorização
	authCode := <-authCodeCh

	// Troca o código por um token
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		logger.Errorf("Erro ao recuperar token: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	logger.Infof("Salvando arquivo de credenciais em: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logger.Errorf("Não foi possível salvar o token oauth: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func UploadFileToDrive(srv *drive.Service, filePath string) error {
	// Abre o arquivo que deseja fazer upload
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %v", err)
	}
	defer file.Close()

	// Cria o arquivo no Google Drive
	driveFile := &drive.File{
		Name: filepath.Base(filePath),
	}

	// Realiza o upload
	_, err = srv.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		return fmt.Errorf("erro ao fazer upload do arquivo: %v", err)
	}

	logger.Infof("Arquivo %s enviado com sucesso!\n", filePath)
	return nil
}

func GoogleDriveSetup() (*drive.Service, error) {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		logger.Errorf("Não foi possível ler o arquivo de credenciais %s: %v", developersGoogleDocs, err)
		return nil, err
	}

	// Escopos necessários para upload
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		logger.Errorf("Não foi possível analisar o arquivo de credenciais %s: %v", developersGoogleDocs, err)
		return nil, err
	}

	// Obtém o cliente autenticado
	client := getClient(config)

	// Cria o serviço do Drive
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logger.Errorf("Não foi possível recuperar o cliente do Drive %s: %v", developersGoogleDocs, err)
		return nil, err
	}

	return srv, nil
}
