package kong

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kong, err := RunContainer(
				ctx,
				testcontainers.WithImage(tt.image),
				WithConfig(filepath.Join(".", "testdata", "kong-plugin.yaml")),
				WithGoPlugin(filepath.Join(".", "go-plugins", "bin", "goplug")),
			)

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
				JSON().Object().ContainsKey("url").HasValue("url", "http://localhost/request/mock/requests")
		})
	}
}

func TestKongGoPlugin_ModifiesHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	kong, err := RunContainer(ctx,
		testcontainers.WithImage("kong/kong:3.4.0"),
		WithConfig(filepath.Join(".", "testdata", "kong-plugin.yaml")),
		WithGoPlugin(filepath.Join(".", "go-plugins", "bin", "goplug")),
		WithLogLevel("info"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Listening on socket: /usr/local/kong/goplug.socket").WithStartupTimeout(30*time.Second),
		),
	)
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
