package kong

import (
	"context"
	"fmt"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
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

	logLine := string(l.Content)
	g.Msgs = append(g.Msgs, logLine)
	fmt.Print(logLine)
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
		//------------ Kong Plugins -----------------
		"KONG_PLUGINS":                       "goplug",
		"KONG_PLUGINSERVER_NAMES":            "goplug",
		"KONG_PLUGINSERVER_GOPLUG_START_CMD": "/usr/local/kong/go-plugins/bin/goplug",
		"KONG_PLUGINSERVER_GOPLUG_QUERY_CMD": "/usr/local/kong/go-plugins/bin/goplug -dump",
	}

	files := []testcontainers.ContainerFile{
		{
			HostFilePath:      "./kong.yaml",
			ContainerFilePath: "/usr/local/kong/kong.yaml",
			FileMode:          0644, // see https://github.com/supabase/cli/pull/132/files
		},
		{
			HostFilePath:      "./go-plugins/bin/goplug", // copy the already compiled binary to the plugins dir
			ContainerFilePath: "/usr/local/kong/go-plugins/bin/goplug",
			FileMode:          0755,
		},
	}
	kong, err := SetupKong(ctx,
		"kong/kong-gateway-dev:3.4.0.0-rc.1",
		env,
		files)
	require.NoError(t, err)

	// doesn't work 🤷‍♂️
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

	e := httpexpect.Default(t, kong.URI)

	//resp, err := http.Get(kong.URI)
	//assert.Nil(t, err)
	//
	//// go get github.com/stretchr/testify
	//assert.Equal(t, resp.StatusCode, http.StatusOK)
	//assert.Equal(t, resp.Header.Get("Server"), "kong/2.8.1")
	// this code is replaced with httpexpect
	e.GET("/").
		Expect().
		Status(http.StatusOK).
		Header("Server").IsEqual("kong/3.4.0.0-enterprise-edition")

	e = httpexpect.Default(t, kong.ProxyURI)

	//get, err := http.Get(kong.ProxyURI)
	//assert.Nil(t, err)
	//
	//all, err := io.ReadAll(get.Body)
	//assert.Nil(t, err)
	r := e.GET("/").
		WithHeader(headers.UserAgent, "Kong Builders").
		Expect()
	r.Status(http.StatusOK).
		Header("X-Kong-Builders").
		IsEqual("Welcome to the jungle 🌴")

	r.Header("Via").IsEqual("kong/3.4.0.0-enterprise-edition")

	var res JSONResponse
	r.JSON().Decode(&res)

	value := res.Headers.Host
	assert.True(t, strings.Contains(value, "mockbin"))
}

type JSONResponse struct {
	StartedDateTime time.Time `json:"startedDateTime"`
	ClientIPAddress string    `json:"clientIPAddress"`
	Method          string    `json:"method"`
	URL             string    `json:"url"`
	HTTPVersion     string    `json:"httpVersion"`
	Cookies         struct {
	} `json:"cookies"`
	Headers struct {
		Host            string `json:"host"`
		Connection      string `json:"connection"`
		AcceptEncoding  string `json:"accept-encoding"`
		XForwardedFor   string `json:"x-forwarded-for"`
		CfRay           string `json:"cf-ray"`
		XForwardedProto string `json:"x-forwarded-proto"`
		CfVisitor       string `json:"cf-visitor"`
		XForwardedHost  string `json:"x-forwarded-host"`
		XForwardedPort  string `json:"x-forwarded-port"`
		XForwardedPath  string `json:"x-forwarded-path"`
		UserAgent       string `json:"user-agent"`
		CfConnectingIP  string `json:"cf-connecting-ip"`
		CdnLoop         string `json:"cdn-loop"`
		XRequestID      string `json:"x-request-id"`
		Via             string `json:"via"`
		ConnectTime     string `json:"connect-time"`
		XRequestStart   string `json:"x-request-start"`
		TotalRouteTime  string `json:"total-route-time"`
	} `json:"headers"`
	QueryString struct {
	} `json:"queryString"`
	PostData struct {
		MimeType string        `json:"mimeType"`
		Text     string        `json:"text"`
		Params   []interface{} `json:"params"`
	} `json:"postData"`
	HeadersSize int `json:"headersSize"`
	BodySize    int `json:"bodySize"`
}
