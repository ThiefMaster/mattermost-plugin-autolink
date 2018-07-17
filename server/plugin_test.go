package main

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
)

func TestPlugin(t *testing.T) {
	links := make([]*Link, 0)
	links = append(links, &Link{
		Pattern:  "(Mattermost)",
		Template: "[Mattermost](https://mattermost.com)",
	})
	validConfiguration := Configuration{links}

	api := &plugintest.API{Store: &plugintest.KeyValueStore{}}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(func(dest interface{}) error {
		*dest.(*Configuration) = validConfiguration
		return nil
	})

	p := Plugin{}
	p.OnActivate(api)

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(post)

	if rpost.Message != "Welcome to [Mattermost](https://mattermost.com)!" {
		t.Fatal("Posted didn't get transformed")
	}
}

func TestSpecialCases(t *testing.T) {
	links := make([]*Link, 0)
	links = append(links, &Link{
		Pattern:  "(Mattermost)",
		Template: "[Mattermost](https://mattermost.com)",
	}, &Link{
		Pattern:  "(Example)",
		Template: "[Example](https://example.com)",
	}, &Link{
		Pattern:  "(foobar)",
		Template: "fb",
	})
	validConfiguration := Configuration{links}

	api := &plugintest.API{Store: &plugintest.KeyValueStore{}}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(func(dest interface{}) error {
		*dest.(*Configuration) = validConfiguration
		return nil
	})

	p := Plugin{}
	p.OnActivate(api)

	var tests = []struct {
		inputMessage    string
		expectedMessage string
	}{
		{
			"hello ``` Mattermost ``` goodbye",
			"hello ``` Mattermost ``` goodbye",
		},
		{
			"hello\n```\nMattermost\n```\ngoodbye",
			"hello\n```\nMattermost\n```\ngoodbye",
		},
		{
			"Mattermost ``` Mattermost ``` goodbye",
			"[Mattermost](https://mattermost.com) ``` Mattermost ``` goodbye",
		},
		{
			"``` Mattermost ``` Mattermost",
			"``` Mattermost ``` [Mattermost](https://mattermost.com)",
		},
		{
			"Mattermost ``` Mattermost ```",
			"[Mattermost](https://mattermost.com) ``` Mattermost ```",
		},
		{
			"Mattermost ``` Mattermost ```\n\n",
			"[Mattermost](https://mattermost.com) ``` Mattermost ```\n\n",
		},
		{
			"hello ` Mattermost ` goodbye",
			"hello ` Mattermost ` goodbye",
		},
		{
			"hello\n`\nMattermost\n`\ngoodbye",
			"hello\n`\nMattermost\n`\ngoodbye",
		},
		{
			"Mattermost ` Mattermost ` goodbye",
			"[Mattermost](https://mattermost.com) ` Mattermost ` goodbye",
		},
		{
			"` Mattermost ` Mattermost",
			"` Mattermost ` [Mattermost](https://mattermost.com)",
		},
		{
			"Mattermost ` Mattermost `",
			"[Mattermost](https://mattermost.com) ` Mattermost `",
		},
		{
			"Mattermost ` Mattermost `\n\n",
			"[Mattermost](https://mattermost.com) ` Mattermost `\n\n",
		},

		{
			"hello ``` Mattermost ``` goodbye ` Mattermost ` end",
			"hello ``` Mattermost ``` goodbye ` Mattermost ` end",
		},
		{
			"hello\n```\nMattermost\n```\ngoodbye ` Mattermost ` end",
			"hello\n```\nMattermost\n```\ngoodbye ` Mattermost ` end",
		},
		{
			"Mattermost ``` Mattermost ``` goodbye ` Mattermost ` end",
			"[Mattermost](https://mattermost.com) ``` Mattermost ``` goodbye ` Mattermost ` end",
		},
		{
			"``` Mattermost ``` Mattermost",
			"``` Mattermost ``` [Mattermost](https://mattermost.com)",
		},
		{
			"```\n` Mattermost `\n```\nMattermost",
			"```\n` Mattermost `\n```\n[Mattermost](https://mattermost.com)",
		},
		{
			"  Mattermost",
			"  [Mattermost](https://mattermost.com)",
		},
		{
			"    Mattermost",
			"    Mattermost",
		},
		{
			"    ```\nMattermost\n    ```",
			"    ```\n[Mattermost](https://mattermost.com)\n    ```",
		},
		{
			"` ``` `\nMattermost\n` ``` `",
			"` ``` `\n[Mattermost](https://mattermost.com)\n` ``` `",
		},
		{
			"Mattermost \n Mattermost",
			"[Mattermost](https://mattermost.com) \n [Mattermost](https://mattermost.com)",
		},
		{
			"[Mattermost](https://mattermost.com)",
			"[Mattermost](https://mattermost.com)",
		},
		{
			"[  Mattermost  ](https://mattermost.com)",
			"[  Mattermost  ](https://mattermost.com)",
		},
		{
			"[  Mattermost  ][1]\n\n[1]: https://mattermost.com",
			"[  Mattermost  ][1]\n\n[1]: https://mattermost.com",
		},
		{
			"![  Mattermost  ](https://mattermost.com/example.png)",
			"![  Mattermost  ](https://mattermost.com/example.png)",
		},
		{
			"![  Mattermost  ][1]\n\n[1]: https://mattermost.com/example.png",
			"![  Mattermost  ][1]\n\n[1]: https://mattermost.com/example.png",
		},
		{
			"foobar\nExample\nfoobar Mattermost",
			"fb\n[Example](https://example.com)\nfb [Mattermost](https://mattermost.com)",
		},
		{
			"foobar",
			"fb",
		},
		{
			"foobarfoobar",
			"foobarfoobar",
		},
	}

	for _, tt := range tests {
		post := &model.Post{
			Message: tt.inputMessage,
		}

		rpost, _ := p.MessageWillBePosted(post)
		if rpost.Message != tt.expectedMessage {
			t.Fatalf("autolink:\n expected '%v'\n actual   '%v'", tt.expectedMessage, rpost.Message)
		}
	}
}
