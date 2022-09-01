package kong

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

type TestLogConsumer struct {
	Msgs []string
	Ack  chan bool
}

const lastMessage = "DONE"

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	if string(l.Content) == fmt.Sprintf("echo %s\n", lastMessage) {
		g.Ack <- true
		return
	}

	g.Msgs = append(g.Msgs, string(l.Content))
}

func TestKongAdminAPI_ReturnVersion(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	env := map[string]string{
		"KONG_DATABASE":           "off",
		"KONG_LOG_LEVEL":          "debug",
		"KONG_PROXY_ACCESS_LOG":   "/dev/stdout",
		"KONG_ADMIN_ACCESS_LOG":   "/dev/stdout",
		"KONG_PROXY_ERROR_LOG":    "/dev/stderr",
		"KONG_ADMIN_ERROR_LOG":    "/dev/stderr",
		"KONG_ADMIN_LISTEN":       "0.0.0.0:8001",
		"KONG_DECLARATIVE_CONFIG": "/usr/local/kong/kong.yaml",
	}

	kong, err := SetupKong(ctx, "kong:2.8.1", env)
	if err != nil {
		t.Fatal(err)
	}

	// doesn't work 🤷‍♂️
	consumer := TestLogConsumer{
		Msgs: []string{},
		Ack:  make(chan bool),
	}
	err = kong.StartLogProducer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	kong.FollowOutput(&consumer)

	// Clean up the container after the test is complete
	defer kong.Terminate(ctx)

	resp, err := http.Get(kong.URI)
	if err != nil {
		t.Fatal(err)
	}

	// go get github.com/stretchr/testify
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	assert.Equal(t, resp.Header.Get("Server"), "kong/2.8.1")

	get, err := http.Get(kong.ProxyURI)
	if err != nil {
		t.Fatal(err)
	}

	all, err := io.ReadAll(get.Body)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, strings.Contains(string(all), "no Route matched with those values"), true)

	/*if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
	if resp.Header.Get("Server") != "kong/2.6.0" {
		t.Fatalf("Expected version %s. Got %s.", "2.6", resp.Header.Get("Server"))
	}*/
}
