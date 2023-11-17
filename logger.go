package logscale

import (
	"fmt"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type logEventItem struct {
	Fields   map[string]string `json:"fields"`
	Messages []string          `json:"messages"`
}

type LogscaleLogger struct {
	crowdStrikeURL string
	ingest_token   string
	environment    string
	messages       chan logEventItem
	wg             *sync.WaitGroup
}

// NewLogscaleLogger is the main entry point of this library.
// It will hook itself into zerolog and send all log messages to logscale
// environment can be a string like "dev", "staging", "prod" that is logged as field env
// There is no other usage of the returned struct except before ending the application make sure to call WaitTillAllMessagesSend()
func NewLogscaleLogger(crowdStrikeURL, ingest_token, environment string) LogscaleLogger {
	l := LogscaleLogger{
		crowdStrikeURL: crowdStrikeURL,
		ingest_token:   ingest_token,
		environment:    environment,
		messages:       make(chan logEventItem, 20), // totally arbitrary buffer size
		wg:             new(sync.WaitGroup),
	}
	l.setLogscaleHookLog()
	return l
}

func (l LogscaleLogger) sendMsg(client *req.Client, payload []logEventItem) {
	defer l.wg.Done()
	resp, err := client.R().SetBody(payload).Post(l.crowdStrikeURL)
	if err != nil {
		fmt.Printf("Failed to send log to Humio: %v", err)
	} else {
		if resp.StatusCode != 200 {
			fmt.Printf("Humio returned: %v", resp.Body)
		}
	}
}

func (l LogscaleLogger) sendMsgs(client *req.Client) {
	for {
		msg := <-l.messages
		payload := []logEventItem{msg}
		l.sendMsg(client, payload)
	}
}

func (l LogscaleLogger) setLogscaleHookLog() {
	client := req.SetTimeout(5*time.Second).
		SetCommonHeader("Content-Type", "application/json").
		SetCommonHeader("Authorization", "Bearer "+l.ingest_token).
		SetUserAgent("go client")
	go l.sendMsgs(client)

	hook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		l.wg.Add(1)

		// e.buf contains a json key/value map of previous/wrapped error messages - perhaps even recursive
		// we should log them as well
		// problem is that there is a missing closing bracket in e.buf, we need to have a workaround
		// also e's fields are not exported

		payload := logEventItem{
			Fields:   map[string]string{"env": l.environment, "level": level.String()},
			Messages: []string{message}}
		l.messages <- payload
	})
	// update the global logger
	log.Logger = log.Logger.Hook(hook)
}

// WaitTillAllMessagesSend acts as a flush
func (l LogscaleLogger) WaitTillAllMessagesSend() {
	// we could this make a bit more robust by checking whether lg is defined in case someone calls this function
	//without calling NewLogscaleLogger. This can happen if someone optionally sends data to humio
	// using waitgroup, waiting for len(messages) == 0 did not work, the last message was not sent
	l.wg.Wait()
	fmt.Println("All messages sent to Logscale")
}
