package sftp

import (
	//"github.com/kuassh/pkg/go-prompt"
	"github.com/c-bata/go-prompt"
)

type sftpClient struct {
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func (sc *sftpClient) Run() {

}
