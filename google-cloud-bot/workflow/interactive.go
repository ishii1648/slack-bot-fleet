package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

type SlackInteractiveMsg struct {
	Workflows []Workflow
}

func NewSlackInteractiveMsg(workflows ...Workflow) *SlackInteractiveMsg {
	return &SlackInteractiveMsg{
		Workflows: workflows,
	}
}

func (s *SlackInteractiveMsg) Handle(c echo.Context) error {
	body, ok := c.Get("body").([]byte)
	if !ok {
		return errors.New("body is not set")
	}

	payload, err := url.QueryUnescape(strings.TrimPrefix(string(body), "payload="))
	if err != nil {
		return fmt.Errorf("Failed to unespace request body: %s", err)
	}

	var interaction slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(payload), &interaction); err != nil {
		return fmt.Errorf("Failed to decode json message from slack: %s", payload)
	}

	callbackAction := interaction.ActionCallback.BlockActions[0]
	commandName := strings.Split(callbackAction.ActionID, "_")[0]
	actionName := strings.Split(callbackAction.ActionID, "_")[1]

	for _, workflow := range s.Workflows {
		if workflow.CommandName() != commandName {
			continue
		}

		actionW := workflow.Action()
		for i, action := range actionW.Actions {
			if action.Name != actionName {
				continue
			}

			if err := actionW.Actions[i].Handler(c.Request().Context(), interaction, actionW.NextActionID(&action)); err != nil {
				return err
			}

			return nil
		}
	}

	c.Logger().Error("no workflow found")
	return c.String(http.StatusInternalServerError, "error")
}
