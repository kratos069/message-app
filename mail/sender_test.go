package mail

import (
	"testing"

	"github.com/kratos069/message-app/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	// in Makefile we set the short flag to skip this test,
	// this will skip tests which take longer time to run (like this one)
	// when all tests ran (avoid spamming emails)
	if testing.Short() {
		t.Skip()
	}

	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName,
		config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "a test email"
	content := `
	<h1>Hello from My Message App</h1>
	<p>this is a message from <a href="message-app.com"> Production grade message app </a></p>
	`
	to := []string{"moazzankamran110@gmail.com"}
	attachFiles := []string{"../notes"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
