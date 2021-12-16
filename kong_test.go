package kong

import (
	"context"
	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/testcontainers/testcontainers-go"
	"net/http"
	"testing"
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

	kong, err := setupKong(ctx)
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

	// go get github.com/stretchr/testify
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	assert.Equal(t, resp.Header.Get("Server"), "kong/2.6.0")

	/*if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}
	if resp.Header.Get("Server") != "kong/2.6.0" {
		t.Fatalf("Expected version %s. Got %s.", "2.6", resp.Header.Get("Server"))
	}*/
}
