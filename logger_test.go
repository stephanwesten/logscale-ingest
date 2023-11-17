package logscale

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

// Use this test as a template for manual testing
func _TestSetLogger(t *testing.T) {
	token := "xxx-xxx-xxx-xxx-xxx"
	l := NewLogscaleLogger("https://cloud.community.humio.com/api/v1/ingest/humio-unstructured", token, "test")
	log.Info().Msg("test humio 1")
	log.Info().Msg("test humio 2")
	log.Info().Msg("test humio 3")
	log.Error().Msg("test error")
	l.WaitTillAllMessagesSend()
}

func TestIngest(t *testing.T) {
	msgChannel := make(chan string, 1)
	// run local webserver
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("got request: %v\n", r)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		msgChannel <- string(body)
		w.WriteHeader(http.StatusOK)
	})
	go func() {
		err := http.ListenAndServe(":8080", nil) //nolint:gosec
		require.NoError(t, err, " failed to start server")
	}()

	// configure
	l := NewLogscaleLogger("http://localhost:8080", "faketoken", "test")

	// test
	log.Info().Msg("test 1")
	log.Info().Msg("test 2")
	msg := <-msgChannel
	require.Contains(t, msg, "test 1")
	msg = <-msgChannel
	require.Contains(t, msg, "test 2")
	// not supported
	// log.Err(fmt.Errorf("fooerr")).Msg("test error")
	// msg = <-msgChannel
	// require.Contains(t, msg, "fooerr")
	l.WaitTillAllMessagesSend()
}
