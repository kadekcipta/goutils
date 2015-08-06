package main

import (
	"fmt"
	"goutils/slackconnect"
)

const webhookUri = "https://hooks.slack.com/services/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"

func main() {
	logger := slackconnect.NewLogger(webhookUri, "slack.db", "#systemd", "Magicsoft-Asia Systems", nil)
	logger.Open()
	defer logger.Close()

	logger.Info("testing")
	//	for i := 0; i < 2; i++ {
	//		go func(n int) {
	//			logger.Info(fmt.Sprintf("Info #%d <https://magicsoft-asia.beanstalkapp.com/goutils/browse/git/bucket/bolt_bucket.go?ref=c-92494c06808d6a1389732ad7934e111ba6d61105|Some Link>", n))
	//		}(i + 1)
	//	}

	fmt.Scanln()
}
