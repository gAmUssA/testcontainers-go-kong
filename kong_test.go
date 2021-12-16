package main

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"testing"
)

type kongContainer struct {
	testcontainers.Container
	URI string
}

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

func setupKong(ctx context.Context) (*kongContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "kong:2.6",
		ExposedPorts: []string{"8001/tcp"},
		WaitingFor:   wait.ForLog("start worker process"),
		Cmd:          []string{"kong", "start"},
		Env: map[string]string{
			"KONG_DATABASE":         "off",
			"KONG_PROXY_ACCESS_LOG": "/dev/stdout",
			"KONG_ADMIN_ACCESS_LOG": "/dev/stdout",
			"KONG_PROXY_ERROR_LOG":  "/dev/stderr",
			"KONG_ADMIN_ERROR_LOG":  "/dev/stderr",
			"KONG_ADMIN_LISTEN":     "0.0.0.0:8001",
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "8001")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	return &kongContainer{Container: container, URI: uri}, nil
}

func TestIntegrationNginxLatestReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	kong, err := setupKong(ctx)
	if err != nil {
		t.Fatal(err)
	}

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
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
	if resp.Header.Get("Server") != "kong/2.6.0" {
		t.Fatalf("Expected version %s. Got %s.", "2.6", resp.Header.Get("Server"))
	}
}
