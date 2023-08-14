package kong

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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
		Image: "kong/kong:3.4.0",
		ExposedPorts: []string{
			defaultProxyPort,
			defaultAdminAPIPort,
			defaultKongManagerPort,
		},
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
		return nil, err
	}

	return &kongContainer{Container: container}, nil
}

// WithConfig adds the kong config file to the container, in the
// "/usr/local/kong/kong.yaml" directory of the container, and
// setting the KONG_DECLARATIVE_CONFIG environment variable to that path.
func WithConfig(cfg string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      cfg,
			ContainerFilePath: "/usr/local/kong/kong.yaml",
			FileMode:          0644, // see https://github.com/supabase/cli/pull/132/files
		})

		// is this variable needed by the default kong image?
		req.Env["KONG_DECLARATIVE_CONFIG"] = "/usr/local/kong/kong.yaml"
	}
}

// WithKongEnv sets environment variables for kong container, possibly overwriting defaults
func WithKongEnv(env map[string]string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		for k, v := range env {
			req.Env[k] = v
		}
	}
}

// WithLogLevel sets log level for kong container, using the KONG_LOG_LEVEL environment variable.
func WithLogLevel(level string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["KONG_LOG_LEVEL"] = level
	}
}

// WithGoPlugin adds a Go plugin to the container, in the "/usr/local/bin" directory of the container
// appending the plugin name to the KONG_PLUGINS and KONG_PLUGINSERVER_NAMES environment variables,
// and setting the KONG_PLUGINSERVER_GOPLUG_START_CMD and KONG_PLUGINSERVER_GOPLUG_QUERY_CMD to the
// executable path of the plugin.
func WithGoPlugin(goPlugPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		pluginName := filepath.Base(goPlugPath) // should be goplug

		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      goPlugPath,
			ContainerFilePath: "/usr/local/bin/" + pluginName,
			FileMode:          0755,
		})

		req.Env["KONG_PLUGINS"] = appendToCommaSeparatedList(req.Env["KONG_PLUGINS"], pluginName)
		req.Env["KONG_PLUGINSERVER_NAMES"] = appendToCommaSeparatedList(req.Env["KONG_PLUGINSERVER_NAMES"], pluginName)

		pluginNameUpper := strings.ToUpper(pluginName)
		req.Env["KONG_PLUGINSERVER_"+pluginNameUpper+"_START_CMD"] = "/usr/local/bin/" + pluginName
		req.Env["KONG_PLUGINSERVER_"+pluginNameUpper+"_QUERY_CMD"] = "/usr/local/bin/" + pluginName + " -dump"
	}
}

func appendToCommaSeparatedList(list, item string) string {
	if len(list) > 0 {
		return list + "," + item
	}
	return item
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
