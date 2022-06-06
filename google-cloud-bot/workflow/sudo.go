package workflow

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ishii1648/slack-bot-fleet/google-cloud-bot/view"
	"github.com/ishii1648/slack-bot-fleet/pkg/googlecloud"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

type SudoWorkflow struct {
	logger    echo.Logger
	slackC    *slack.Client
	action    *ActionWorkflow
	projectID string
}

func NewSudo(logger echo.Logger, commandName string, slackC *slack.Client) *SudoWorkflow {
	sudoWorkflow := &SudoWorkflow{
		logger: logger,
		slackC: slackC,
	}
	sudoWorkflow.action = NewActionWorkflow(commandName, []Action{
		{
			Name:    "select-project-roles",
			Handler: sudoWorkflow.SelectProjectRoles,
		},
		{
			Name:    "set-iam-policy",
			Handler: sudoWorkflow.SetIamPolicy,
		},
	})

	return sudoWorkflow
}

func (w *SudoWorkflow) CommandName() string {
	return w.action.CommandName
}

func (w *SudoWorkflow) Action() *ActionWorkflow {
	return w.action
}

func (w *SudoWorkflow) Handle(ctx echo.Context) error {
	bucketName, isSet := os.LookupEnv("BUCKET_NAME")
	if !isSet {
		return errors.New("BUCKET_NAME is not set")
	}

	projects, err := googlecloud.FetchProjects(ctx.Request().Context(), bucketName)
	if err != nil {
		return err
	}

	var options []*slack.OptionBlockObject
	for i, project := range projects {
		options = append(options, &slack.OptionBlockObject{
			Value:       "value-" + strconv.Itoa(i),
			Text:        slack.NewTextBlockObject("plain_text", project.ID, false, false),
			Description: slack.NewTextBlockObject("plain_text", project.Name, false, false),
		})
	}

	return ctx.JSON(http.StatusOK, view.SelectProjecs(options, w.Action().NextActionID(nil)))
}

func (w *SudoWorkflow) SelectProjectRoles(ctx context.Context, interaction slack.AttachmentActionCallback, nextActionID string) error {
	w.projectID = interaction.ActionCallback.BlockActions[0].SelectedOption.Text.Text

	bucketName, isSet := os.LookupEnv("BUCKET_NAME")
	if !isSet {
		return errors.New("BUCKET_NAME is not set")
	}

	projectRoles, err := googlecloud.FetchProjectRoles(ctx, bucketName)
	if err != nil {
		return err
	}

	var options []*slack.OptionBlockObject
	for i, role := range projectRoles {
		options = append(options, &slack.OptionBlockObject{
			Value:       "value-" + strconv.Itoa(i),
			Text:        slack.NewTextBlockObject("plain_text", role.ID, false, false),
			Description: slack.NewTextBlockObject("plain_text", role.Name, false, false),
		})
	}

	if _, _, err := w.slackC.PostMessage(
		interaction.Channel.ID,
		slack.MsgOptionBlocks(view.SelectProjectRoles(options, nextActionID)...),
		slack.MsgOptionTS(interaction.Message.Timestamp),
	); err != nil {
		return err
	}

	return nil
}

func (w *SudoWorkflow) SetIamPolicy(ctx context.Context, interaction slack.AttachmentActionCallback, nextActionID string) error {
	type result struct {
		roleID string
		ok     bool
	}

	member, ok := os.LookupEnv("IAM_MEMBER")
	if !ok {
		return errors.New("IAM_MEMBER is not set")
	}

	resultsCh := make(chan []*result)

	go func() {
		for {
			select {
			case results := <-resultsCh:
				var blocks []slack.Block
				for _, result := range results {
					blocks = append(blocks, view.AddedProjectRole(member, result.roleID, result.ok)[0])
				}
				if _, _, err := w.slackC.PostMessage(
					interaction.Channel.ID,
					slack.MsgOptionBlocks(blocks...),
					slack.MsgOptionTS(interaction.Message.Timestamp),
				); err != nil {
					w.logger.Errorf("failed to PostMessage :%s", err)
				}
			case <-time.After(60 * time.Second):
				w.logger.Error("timeout: SetIamPolicyWithFixedTime")
				return
			}
		}
	}()

	selectedOptions := interaction.ActionCallback.BlockActions[0].SelectedOptions

	// this function done before complete goroutine to exec.Command for response within 3 secounds
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		var results []*result
		wg.Done()
		for _, option := range selectedOptions {
			roleID := option.Text.Text
			if err := googlecloud.SetIamPolicyWithFixedTime(member, w.projectID, roleID); err != nil {
				w.logger.Errorf("failed to SetIamPolicyWithFixedTime: %s(%s)", err, roleID)
				results = append(results, &result{roleID, false})
				continue
			}
			w.logger.Debugf("done to SetIamPolicyWithFixedTime: %s", roleID)
			results = append(results, &result{roleID, true})
		}
		resultsCh <- results
	}()

	wg.Wait()

	return nil
}
