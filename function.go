package p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
)

var channel = os.Getenv("CHANNEL")
var apiToken = os.Getenv("API_TOKEN")
var slackAPIURL = "https://slack.com/api/files.upload"

// GCSEvent is the payload of a GCS event. Please refer to the docs for
// additional information regarding GCS events.
type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

type Log struct {
	InsertId         string                 `json:"insertId"`
	LogName          string                 `json:"logName"`
	ReceiveTimestamp string                 `json:"receiveTimestamp"`
	Resource         Resource               `json:"resource"`
	TextPayload      string                 `json:"textPayload"`
	JSONPayload      map[string]interface{} `json:"jsonPayload"`
}

type Resource struct {
	Labels map[string]interface{} `json:"labels"`
	Type   string                 `json:"type"`
}

func NotifyToSlack(ctx context.Context, e GCSEvent) error {
	log.Printf("Processing file: %s", e.Name)
	c, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("[error] new client failed, %s", err)
		return nil
	}

	bkt := c.Bucket(e.Bucket)
	obj := bkt.Object(e.Name)

	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Printf("[error] NewReader failed, %s", err)
		return nil
	}
	defer r.Close()

	var text bytes.Buffer
	d := json.NewDecoder(r)
	var logLines int
	for d.More() {
		logLines++
		var logData Log
		if err := d.Decode(&logData); err != nil {
			log.Printf("[error] Decode failed, %s", err)
			continue
		}
		if logData.TextPayload != "" {
			text.WriteString(logData.TextPayload + "\n")
		}
		if logData.JSONPayload != nil {
			jtxt, err := json.Marshal(logData.JSONPayload)
			if err != nil {
				log.Printf("[error] Marshal jsonPayload failed, %s", err)
				continue
			}
			text.Write(jtxt)
			text.WriteString("\n")
		}
	}

	comment := fmt.Sprintf("log length: %d", logLines)
	if logLines >= 50 {
		comment = fmt.Sprintf("<!channel> %s", comment)
	}

	if err := postToSlack(apiToken, channel, e.Name, comment, text.String()); err != nil {
		log.Printf("[error] postToSlack failed, %s", err)
		return nil
	}

	return nil
}

func postToSlack(token, channel, title, comment, content string) error {
	values := url.Values{}
	values.Set("token", token)
	values.Add("channels", channel)
	values.Add("title", title)
	values.Add("initial_comment", comment)
	values.Add("content", content)

	req, err := http.NewRequest(
		"POST",
		slackAPIURL,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}
