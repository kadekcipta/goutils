package slackconnect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goutils/bucket"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	throttlingTime = 2 // seconds
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Fallback   string  `json:"fallback"`
	Color      string  `json:"color"`
	Pretext    string  `json:"pretext"`
	AuthorName string  `json:"author_name"`
	AuthorLink string  `json:"author_link"`
	AuthorIcon string  `json:"author_icon"`
	Title      string  `json:"title"`
	TitleLink  string  `json:"title_link"`
	Text       string  `json:"text"`
	ImageUrl   string  `json:"image_url"`
	Fields     []Field `json:"fields", omitempty`
}

type Payload struct {
	Text        string       `json:"text"`
	Channel     string       `json:"channel"`
	UserName    string       `json:"username"`
	Icon        string       `json:"icon_emoji"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type PayloadFunc func(channel string, userName string, b []byte) Payload

type Logger interface {
	Info(m string)
	Warn(m string)
	Error(m string)
	Critical(m string)
	Open() error
	Close() error
}

type slackLogger struct {
	sync.RWMutex
	sender      string
	dbName      string
	channel     string
	webhookUri  string
	usePrefix   bool
	payloadFunc PayloadFunc
	done        chan struct{}
	b           *bucket.LocalBucket
}

func (l *slackLogger) trySend() {
	l.Lock()
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
		l.Unlock()
	}()

	if l.b == nil {
		return
	}

	k, v, err := l.b.First()
	if err != nil {
		return
	}

	if v == nil {
		return
	}

	var payload Payload

	if l.payloadFunc != nil {
		payload = l.payloadFunc(l.channel, l.sender, v)
	} else {
		payload = Payload{
			Text:     string(v),
			Channel:  l.channel,
			UserName: l.sender,
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data)
	_, err = http.Post(l.webhookUri, "application/json", buf)
	if err != nil {
		return
	}
	l.b.Remove(k)
}

func (l *slackLogger) startSender() {
	for {
		select {
		case <-time.After(time.Second * throttlingTime):
			l.trySend()

		case <-l.done:
			return
		}
	}
}

func (l *slackLogger) log(p, m string) {
	l.Lock()
	defer l.Unlock()

	message := l.timestamp(p, m)

	// put it in the bucket
	key := time.Now().UTC().UnixNano()
	l.b.Put(strconv.FormatInt(key, 10), []byte(message))
}

func (l *slackLogger) Open() error {
	l.Lock()
	defer l.Unlock()

	go l.startSender()

	return l.b.Open()
}

func (l *slackLogger) Close() error {
	l.Lock()
	defer l.Unlock()

	close(l.done)
	return l.b.Close()
}

func (l *slackLogger) timestamp(p, m string) string {
	if l.usePrefix {
		now := time.Now().Format(time.RFC1123Z)
		return fmt.Sprintf("[%s: %s] %s", p, now, m)
	}

	return m
}

func (l *slackLogger) Info(m string) {
	l.log("INFO", m)
}

func (l *slackLogger) Warn(m string) {
	l.log("WARN", m)
}

func (l *slackLogger) Error(m string) {
	l.log("ERROR", m)
}

func (l *slackLogger) Critical(m string) {
	l.log("CRITICAL", m)
}

func NewLogger(webhookUri, dbName, channel, sender string, payloadFunc PayloadFunc, v ...bool) Logger {
	usePrefix := false
	if len(v) > 0 {
		usePrefix = v[0]
	}
	return &slackLogger{
		usePrefix:   usePrefix,
		dbName:      dbName,
		channel:     channel,
		sender:      sender,
		webhookUri:  webhookUri,
		payloadFunc: payloadFunc,
		done:        make(chan struct{}),
		b:           bucket.NewLocalBucket(dbName),
	}
}
