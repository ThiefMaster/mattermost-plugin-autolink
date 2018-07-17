package main

import (
	"fmt"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/mlog"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/utils/markdown"
)

// Plugin the main struct for everything
type Plugin struct {
	api   plugin.API
	links atomic.Value
}

// OnActivate is invoked when the plugin is activated.
func (p *Plugin) OnActivate(api plugin.API) error {
	p.api = api

	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	return nil
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var c Configuration
	err := p.api.LoadPluginConfiguration(&c)
	if err != nil {
		return err
	}

	links := make([]*AutoLinker, 0)

	for _, l := range c.Links {
		al, lerr := NewAutoLinker(l)
		if lerr != nil {
			mlog.Error("Error creating autolinker: ")
		}

		links = append(links, al)
	}

	p.links.Store(links)
	return nil
}

// MessageWillBePosted is invoked when a message is posted by a user before it is commited
// to the database.
func (p *Plugin) MessageWillBePosted(post *model.Post) (*model.Post, string) {
	links := p.links.Load().([]*AutoLinker)
	postText := post.Message
	offset := 0
	markdown.Inspect(post.Message, func(node interface{}) bool {
		switch node.(type) {
		// never descend into the text content of a link/image
		case *markdown.InlineLink:
			return false
		case *markdown.InlineImage:
			return false
		case *markdown.ReferenceLink:
			return false
		case *markdown.ReferenceImage:
			return false
		}

		if textNode, ok := node.(*markdown.Text); ok {
			startPos, endPos := textNode.Range.Position+offset, textNode.Range.End+offset
			origText := postText[startPos:endPos]
			if textNode.Text != origText {
				mlog.Error(fmt.Sprintf("Markdown text did not match range text, '%s' != '%s'", textNode.Text, origText))
				return true
			}

			newText := origText
			for _, l := range links {
				newText = l.Replace(newText)
			}

			if origText != newText {
				postText = postText[:startPos] + newText + postText[endPos:]
				offset += len(newText) - len(origText)
			}
		}
		return true
	})
	post.Message = postText

	return post, ""
}
