package view

import (
	"fmt"

	"github.com/slack-go/slack"
)

func SelectProjecs(options []*slack.OptionBlockObject, nextActionID string) slack.Message {
	blocks := []slack.Block{
		slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "plain_text",
				Text: "対象のプロジェクトを選択してください",
			},
			Accessory: slack.NewAccessory(&slack.SelectBlockElement{
				Type: slack.OptTypeStatic,
				Placeholder: &slack.TextBlockObject{
					Type: "plain_text",
					Text: "プロジェクト一覧から選択する",
				},
				ActionID: nextActionID,
				Options:  options,
			}),
		},
	}

	msg := slack.NewBlockMessage(blocks...)
	msg.ResponseType = slack.ResponseTypeInChannel

	return msg
}

func SelectProjectRoles(options []*slack.OptionBlockObject, nextActionID string) []slack.Block {
	return []slack.Block{
		slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "plain_text",
				Text: "付与する Role を選択してください",
			},
			Accessory: slack.NewAccessory(&slack.MultiSelectBlockElement{
				Type: slack.MultiOptTypeStatic,
				Placeholder: &slack.TextBlockObject{
					Type: "plain_text",
					Text: "Role 一覧から選択する",
				},
				ActionID: nextActionID,
				Options:  options,
			}),
		},
	}
}

func AddedProjectRole(member string, roleID string, success bool) []slack.Block {
	if success {
		return SuccessAddedProjectRole(member, roleID)
	} else {
		return FailedAddedProjectRole(member, roleID)
	}
}

func SuccessAddedProjectRole(member string, roleID string) []slack.Block {
	return []slack.Block{
		slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf(":ok: `%s` に `%s` を付与しました", member, roleID),
			},
		},
	}
}

func FailedAddedProjectRole(member string, roleID string) []slack.Block {
	return []slack.Block{
		slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf(":ng: `%s` に `%s` を付与できませんでした", member, roleID),
			},
		},
	}
}
