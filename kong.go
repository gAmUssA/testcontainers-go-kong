package kong

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
)

type kongContainer struct {
	testcontainers.Container
}

// TODO: mention all ports
var (
	defaultProxyPort       = "8000/tcp"
	defaultAdminAPIPort    = "8001/tcp"
	defaultKongManagerPort = "8002/tcp"
)

// RunContainer is the entrypoint to the module
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*kongContainer, error) {

	req := testcontainers.ContainerRequest{
		//Image: "kong/kong-gateway:3.3.0",
		ExposedPorts: []string{
			defaultProxyPort,
			defaultAdminAPIPort},
		WaitingFor: wait.ForListeningPort(nat.Port(defaultAdminAPIPort)),
		Cmd:        []string{"kong", "start"},
		Env: map[string]string{
			// default env variables, can be overwritten in test method
			"KONG_DATABASE":           "off",
			"KONG_LOG_LEVEL":          "debug",
			"KONG_PROXY_ACCESS_LOG":   "/dev/stdout",
			"KONG_ADMIN_ACCESS_LOG":   "/dev/stdout",
			"KONG_PROXY_ERROR_LOG":    "/dev/stderr",
			"KONG_ADMIN_ERROR_LOG":    "/dev/stderr",
			"KONG_ADMIN_LISTEN":       "0.0.0.0:8001",
			"KONG_DECLARATIVE_CONFIG": "/usr/local/kong/kong.yaml",
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(".", "testdata", "kong-config-default.yaml"),
				ContainerFilePath: "/usr/local/kong/kong.yaml",
				FileMode:          0644, // see
			},
		},
	}

	genericContainerRequest := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerRequest)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerRequest)
	if err != nil {
		log.Fatal(err)
	}

	return &kongContainer{Container: container}, nil
}

// KongUrls returns admin url, proxy url, or error
func (c kongContainer) KongUrls(ctx context.Context, args ...string) (string, string, error) {
	ip, err := c.Host(ctx)
	if err != nil {
		return "", "", err
	}
	mappedPort, err := c.MappedPort(ctx, nat.Port(defaultAdminAPIPort))
	if err != nil {
		return "", "", err
	}
	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	proxyMappedPort, err := c.MappedPort(ctx, nat.Port(defaultProxyPort))
	if err != nil {
		return "", "", err
	}

	pUri := fmt.Sprintf("http://%s:%s", ip, proxyMappedPort.Port())

	return uri, pUri, nil
}

// KongCustomizer type represents a container customizer for transferring state from the options to the container
type KongCustomizer struct {
}

// Customize method implementation
func (c KongCustomizer) Customize(req *testcontainers.GenericContainerRequest) testcontainers.ContainerRequest {
	//	req.ExposedPorts = append(req.ExposedPorts, "1234/tcp")
	return req.ContainerRequest
}
