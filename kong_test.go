package kong

import (
	"context"
	"fmt"
	"github.com/gavv/httpexpect/v2"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"strings"
	"testing"
	"time"
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

	logLine := string(l.Content)
	g.Msgs = append(g.Msgs, logLine)
	fmt.Print(logLine)
}

func TestKongAdminAPI_ReturnVersion(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		image   string
		version string
		//wait  wait.Strategy
	}{
		{
			name:    "Kong OSS",
			image:   "kong/kong:3.4.0",
			version: "kong/3.4.0",
		},
		{
			name:    "Kong Gateway",
			image:   "kong/kong-gateway:3.4.0.0",
			version: "kong/3.4.0.0-enterprise-edition",
		},
	}
	files := []testcontainers.ContainerFile{
		{
			HostFilePath:      "./fixtures/kong-mockbin.yaml",
			ContainerFilePath: "/usr/local/kong/kong.yaml",
			FileMode:          0644, // see https://github.com/supabase/cli/pull/132/files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kong, err := RunContainer(ctx, testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Image: tt.image,
					Files: files,
				}}))

			require.NoError(t, err)
			// Clean up the container after the test is complete
			t.Cleanup(func() {
				if err := kong.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			adminUrl, proxy, err := kong.KongUrls(ctx)
			require.Nil(t, err)

			e := httpexpect.Default(t, adminUrl)

			e.GET("/").
				Expect().
				Status(http.StatusOK).
				Header("Server").IsEqual(tt.version)

			e = httpexpect.Default(t, proxy)
			e.GET("/mock/requests").
				Expect().Status(http.StatusOK).
				JSON().Object().ContainsKey("url").HasValue("url", "http://localhost/requests")
		})
	}
}

func TestKongGoPlugin_ModifiesHeaders(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	env := map[string]string{
		"KONG_LOG_LEVEL": "info",
		//------------ Kong Plugins -----------------
		"KONG_PLUGINS":                       "goplug",
		"KONG_PLUGINSERVER_NAMES":            "goplug",
		"KONG_PLUGINSERVER_GOPLUG_START_CMD": "/usr/local/bin/goplug",
		"KONG_PLUGINSERVER_GOPLUG_QUERY_CMD": "/usr/local/bin/goplug -dump",
	}

	files := []testcontainers.ContainerFile{
		{
			HostFilePath:      "./fixtures/kong-plugin.yaml",
			ContainerFilePath: "/usr/local/kong/kong.yaml",
			FileMode:          0644, // see https://github.com/supabase/cli/pull/132/files
		},
		{
			HostFilePath:      "./go-plugins/bin/goplug", // copy the already compiled binary to the plugins dir
			ContainerFilePath: "/usr/local/bin/goplug",
			FileMode:          0755,
		},
	}

	image := "kong/kong:3.4.0"
	kong, err := RunContainer(ctx, testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			// needed because the official Docker image does not have the go-plugins/bin directory already created
			Image: image,
			/*FromDockerfile: testcontainers.FromDockerfile{
				Context: ".",
				BuildArgs: map[string]*string{
					"TC_KONG_IMAGE": &image,
				},
				PrintBuildLog: true,
			},*/
			Env:        env,
			Files:      files,
			WaitingFor: wait.ForLog("Listening on socket: /usr/local/kong/goplug.socket").WithStartupTimeout(30 * time.Second),
		},
	}))
	require.NoError(t, err)

	consumer := TestLogConsumer{
		Msgs: []string{},
		Ack:  make(chan bool),
	}
	err = kong.StartLogProducer(ctx)
	assert.Nil(t, err)

	defer kong.StopLogProducer()
	kong.FollowOutput(&consumer)

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := kong.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	_, proxyUrl, err := kong.KongUrls(ctx)
	require.Nil(t, err)

	e := httpexpect.Default(t, proxyUrl)

	r := e.GET("/").
		WithHeader(headers.UserAgent, "Kong Builders").
		Expect()
	r.Status(http.StatusOK).
		Header("X-Kong-Builders").
		IsEqual("Welcome to the jungle 🌴")

	r.Header("Via").IsEqual("kong/3.4.0")

	var res JSONResponse
	r.JSON().Decode(&res)

	value := res.Headers.Host
	assert.True(t, strings.Contains(value, "mockbin"))
}
