package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/V1ctorW1ll1an/MaisSaudeBackup/config/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var (
	developersGoogleDocs = "https://developers.google.com/workspace/drive/api/quickstart/go?hl=pt-br"
	l                    *logger.Logger
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

	authCodeCh := make(chan string)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		authCodeCh <- authCode

		fmt.Fprintf(w, "Autenticação concluída. Você pode fechar esta janela.")
	})

	config.RedirectURL = "http://localhost:8989/callback"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	go func() {
		fmt.Println("Abra este link no seu navegador:")
		fmt.Println(authURL)
		l.Info("Servidor de autenticação iniciado na porta 8989")
		l.Error(http.ListenAndServe(":8989", nil))
	}()

	authCode := <-authCodeCh

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		l.Errorf("Erro ao recuperar token: %v", err)
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
	l.Infof("Salvando arquivo de credenciais em: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		l.Errorf("Não foi possível salvar o token oauth: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func UploadFileToDrive(ds *drive.Service, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %v", err)
	}
	defer f.Close()

	driveFile := &drive.File{
		Name: filepath.Base(filePath),
	}

	_, err = ds.Files.Create(driveFile).Media(f).Do()
	if err != nil {
		return fmt.Errorf("erro ao fazer upload do arquivo: %v", err)
	}

	l.Infof("Arquivo %s enviado com sucesso!\n", filePath)
	return nil
}

func New() (*drive.Service, error) {
	ctx := context.Background()

	l, err := logger.New("./logs")
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		l.Errorf("Não foi possível ler o arquivo de credenciais %s: %v", developersGoogleDocs, err)
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		l.Errorf("Não foi possível analisar o arquivo de credenciais %s: %v", developersGoogleDocs, err)
		return nil, err
	}

	client := getClient(config)

	ds, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		l.Errorf("Não foi possível recuperar o cliente do Drive %s: %v", developersGoogleDocs, err)
		return nil, err
	}

	return ds, nil
}
