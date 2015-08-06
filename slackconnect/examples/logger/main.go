package main

import (
	"fmt"
	"goutils/slackconnect"
)

const webhookUri = "<your webhook uri>"

func main() {
	logger := slackconnect.NewLogger(webhookUri, "slack.db", "#orders", "The GOLEK", nil)
	logger.Open()
	defer logger.Close()
	logger.Info("Hello..")
	fmt.Scanln()
}
