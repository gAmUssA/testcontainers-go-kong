package kong

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type kongContainer struct {
	testcontainers.Container
	URI string
}

func setupKong(ctx context.Context) (*kongContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "kong:2.8",
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
