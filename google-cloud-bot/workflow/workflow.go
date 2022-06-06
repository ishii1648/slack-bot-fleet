package workflow

import (
	"context"

	"github.com/slack-go/slack"
)

type Workflow interface {
	CommandName() string
	Action() *ActionWorkflow
}

type ActionWorkflow struct {
	CommandName string
	Actions     []Action
}

type Action struct {
	Name    string
	Handler Handler
}

type Handler func(ctx context.Context, interaction slack.AttachmentActionCallback, nextActionID string) error

func NewActionWorkflow(commandName string, actions []Action) *ActionWorkflow {
	return &ActionWorkflow{
		CommandName: commandName,
		Actions:     actions,
	}
}

func (a *ActionWorkflow) NextActionID(currentAction *Action) string {
	if currentAction == nil {
		return a.CommandName + "_" + a.Actions[0].Name
	}

	for i, action := range a.Actions {
		if action.Name == currentAction.Name && i < len(a.Actions)-1 {
			return a.CommandName + "_" + a.Actions[i+1].Name
		}
	}

	return ""
}
