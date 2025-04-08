package whatsapp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigWhatsappApi(t *testing.T) {
	os.Setenv("WHATSAPP_TOKEN", "test_token")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "+5511999999999")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "+5511888888888")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	wc, err := ConfigWhatsappApi()
	assert.NoError(t, err, "ConfigWhatsappApi should not return an error")
	assert.NotNil(t, wc, "ConfigWhatsappApi should return a valid client")
	assert.Equal(t, "test_token", wc.Token, "Token should be set correctly")
	assert.Equal(t, "+5511999999999", wc.From, "From should be set correctly")
	assert.Equal(t, "+5511888888888", wc.To, "To should be set correctly")
	assert.Equal(t, "test_template", wc.Parameters.TemplateId, "TemplateId should be set correctly")
}

func TestConfigWhatsappApi_Error(t *testing.T) {
	// Test missing token
	os.Setenv("WHATSAPP_TOKEN", "")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "+5511999999999")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "+5511888888888")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	wc, err := ConfigWhatsappApi()
	assert.Error(t, err, "ConfigWhatsappApi should return an error when token is missing")
	assert.Nil(t, wc, "ConfigWhatsappApi should return a nil client when token is missing")
	assert.Equal(t, "WHATSAPP_TOKEN não foi definido", err.Error(), "Error message should be correct")

	// Test missing from number
	os.Setenv("WHATSAPP_TOKEN", "test_token")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "+5511888888888")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	wc, err = ConfigWhatsappApi()
	assert.Error(t, err, "ConfigWhatsappApi should return an error when from number is missing")
	assert.Nil(t, wc, "ConfigWhatsappApi should return a nil client when from number is missing")
	assert.Equal(t, "WHATSAPP_PHONE_NUMBER_FROM não foi definido", err.Error(), "Error message should be correct")

	// Test missing to number
	os.Setenv("WHATSAPP_TOKEN", "test_token")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "+5511999999999")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	wc, err = ConfigWhatsappApi()
	assert.Error(t, err, "ConfigWhatsappApi should return an error when to number is missing")
	assert.Nil(t, wc, "ConfigWhatsappApi should return a nil client when to number is missing")
	assert.Equal(t, "WHATSAPP_PHONE_NUMBER_TO não foi definido", err.Error(), "Error message should be correct")
}

func TestSendWhatsappMessage_Success(t *testing.T) {
	// Setup environment variables
	os.Setenv("WHATSAPP_TOKEN", "test_token")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "+5511999999999")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "+5511888888888")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "test_token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("accept"))
		assert.Equal(t, "application/*+json", r.Header.Get("content-type"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	// Configure client and override URL
	wc, err := ConfigWhatsappApi()
	assert.NoError(t, err)

	originalURL := url
	url = server.URL
	defer func() { url = originalURL }()

	// Test successful message sending
	err = wc.Send("victor", "scm_test", "07/04/2025 ÀS 16:45", "erro exception")
	assert.NoError(t, err)
}

func TestSendWhatsappMessage_InvalidResponse(t *testing.T) {
	// Setup environment variables
	os.Setenv("WHATSAPP_TOKEN", "test_token")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "+5511999999999")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "+5511888888888")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	// Create mock server with invalid JSON response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	// Configure client and override URL
	wc, err := ConfigWhatsappApi()
	assert.NoError(t, err)

	originalURL := url
	url = server.URL
	defer func() { url = originalURL }()

	// Test that invalid JSON doesn't cause an error if status is 200
	err = wc.Send("victor", "scm_test", "07/04/2025 ÀS 16:45", "erro exception")
	assert.NoError(t, err)
}

func TestSendWhatsappMessage_ServerError(t *testing.T) {
	// Setup environment variables
	os.Setenv("WHATSAPP_TOKEN", "test_token")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "+5511999999999")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "+5511888888888")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	// Create mock server that returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request"}`))
	}))
	defer server.Close()

	// Configure client and override URL
	wc, err := ConfigWhatsappApi()
	assert.NoError(t, err)

	originalURL := url
	url = server.URL
	defer func() { url = originalURL }()

	// Test error response from server
	err = wc.Send("victor", "scm_test", "07/04/2025 ÀS 16:45", "erro exception")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send message")
}

func TestSendWhatsappMessage_ConnectionError(t *testing.T) {
	// Setup environment variables
	os.Setenv("WHATSAPP_TOKEN", "test_token")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "+5511999999999")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "+5511888888888")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	// Configure client
	wc, err := ConfigWhatsappApi()
	assert.NoError(t, err)

	// Test unreachable server
	originalURL := url
	url = "http://localhost:9999/nonexistent"
	defer func() { url = originalURL }()

	err = wc.Send("victor", "scm_test", "07/04/2025 ÀS 16:45", "erro exception")
	assert.Error(t, err)
}

func TestSendWhatsappMessage_ConnectionClosed(t *testing.T) {
	// Setup environment variables
	os.Setenv("WHATSAPP_TOKEN", "test_token")
	os.Setenv("WHATSAPP_PHONE_NUMBER_FROM", "+5511999999999")
	os.Setenv("WHATSAPP_PHONE_NUMBER_TO", "+5511888888888")
	os.Setenv("WHATSAPP_TEMPLATE_ID", "test_template")

	// Create mock server that closes connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("webserver doesn't support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer server.Close()

	// Configure client and override URL
	wc, err := ConfigWhatsappApi()
	assert.NoError(t, err)

	originalURL := url
	url = server.URL
	defer func() { url = originalURL }()

	// Test connection closed error
	err = wc.Send("victor", "scm_test", "07/04/2025 ÀS 16:45", "erro exception")
	assert.Error(t, err)
}
