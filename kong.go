package kong

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type kongContainer struct {
	testcontainers.Container
	URI      string
	ProxyURI string
}

func SetupKong(ctx context.Context, image string, environment map[string]string) (*kongContainer, error) {

	req := testcontainers.ContainerRequest{
		// needed because the official Docker image does not have the go-plugins/bin directory already created
		FromDockerfile: testcontainers.FromDockerfile{
			Context: ".",
			BuildArgs: map[string]*string{
				"TC_KONG_IMAGE": &image,
			},
		},
		ExposedPorts: []string{"8001/tcp", "8000/tcp"},
		WaitingFor:   wait.ForListeningPort("8001/tcp"),
		//Cmd:          []string{"kong", "start"},
		Cmd: []string{"kong", "start"},
		Env: environment,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "./kong.yaml",
				ContainerFilePath: "/usr/local/kong/kong.yaml",
				FileMode:          0644, // see https://github.com/supabase/cli/pull/132/files
			},
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		log.Fatal(err)
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

	proxyMappedPort, err := container.MappedPort(ctx, "8000")
	if err != nil {
		return nil, err
	}

	pUri := fmt.Sprintf("http://%s:%s", ip, proxyMappedPort.Port())

	return &kongContainer{Container: container, URI: uri, ProxyURI: pUri}, nil
}
