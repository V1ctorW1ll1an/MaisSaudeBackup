package whatsapp

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var url = "https://api.wts.chat/chat/v1/message/send-sync"

type WhatsappConfig struct {
	From       string
	To         string
	Token      string
	Parameters MessageParameters
}

type MessageParameters struct {
	TemplateId string
}

func ConfigWhatsappApi() (*WhatsappConfig, error) {
	token := os.Getenv("WHATSAPP_TOKEN")
	from := os.Getenv("WHATSAPP_PHONE_NUMBER_FROM")
	to := os.Getenv("WHATSAPP_PHONE_NUMBER_TO")
	templateId := os.Getenv("WHATSAPP_TEMPLATE_ID")
	if token == "" {
		return nil, fmt.Errorf("WHATSAPP_TOKEN não foi definido")
	}

	if from == "" {
		return nil, fmt.Errorf("WHATSAPP_PHONE_NUMBER_FROM não foi definido")
	}

	if to == "" {
		return nil, fmt.Errorf("WHATSAPP_PHONE_NUMBER_TO não foi definido")
	}

	return &WhatsappConfig{
		From:  from,
		To:    to,
		Token: token,
		Parameters: MessageParameters{
			TemplateId: templateId,
		},
	}, nil
}

func (wc *WhatsappConfig) Send(nome string, database string, data_hora string, erro string) error {
	payload := strings.NewReader(fmt.Sprintf("{\"body\":{\"parameters\":{\"nome\":\"%s\",\"database\":\"%s\",\"data_hora\":\"%s\",\"erro\":\"%s\"},\"templateId\":\"%s\"},\"from\":\"%s\",\"to\":\"%s\"}", nome, database, data_hora, erro, wc.Parameters.TemplateId, wc.From, wc.To))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/*+json")
	req.Header.Add("Authorization", wc.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message: %s", string(body))
	}

	return nil
}
