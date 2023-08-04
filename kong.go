package kong

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type kongContainer struct {
	testcontainers.Container
	URI      string
	ProxyURI string
}

// TODO: mention all ports
var (
	defaultProxyPort    = "8000/tcp"
	defaultAdminAPIPort = "8001/tcp"
)

func SetupKong(ctx context.Context,
	image string,
	environment map[string]string,
	files []testcontainers.ContainerFile,
	opts ...testcontainers.ContainerCustomizer) (*kongContainer, error) {

	req := testcontainers.ContainerRequest{
		// needed because the official Docker image does not have the go-plugins/bin directory already created
		FromDockerfile: testcontainers.FromDockerfile{
			Context: ".",
			BuildArgs: map[string]*string{
				"TC_KONG_IMAGE": &image,
			},
			PrintBuildLog: true,
		},
		ExposedPorts: []string{
			defaultProxyPort,
			defaultAdminAPIPort},
		WaitingFor: wait.ForListeningPort(nat.Port(defaultAdminAPIPort)),
		Cmd:        []string{"kong", "start"},
		Env:        environment,
		Files:      files,
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

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(defaultAdminAPIPort))
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	proxyMappedPort, err := container.MappedPort(ctx, nat.Port(defaultProxyPort))
	if err != nil {
		return nil, err
	}

	pUri := fmt.Sprintf("http://%s:%s", ip, proxyMappedPort.Port())

	return &kongContainer{Container: container, URI: uri, ProxyURI: pUri}, nil
}
