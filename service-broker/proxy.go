package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pkghttp "net/http"
	"os"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	exampleapi "github.com/ishii1648/slack-bot-fleet/api/example"
	"github.com/ishii1648/slack-bot-fleet/pkg/route"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.opencensus.io/trace"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

func Run(ctx context.Context) (string, *AppError) {
	logger := zerolog.Ctx(ctx)

	body, ok := ctx.Value("requestBody").([]byte)
	if !ok {
		return "", Error(pkghttp.StatusBadRequest, "requestBody not found")
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		return "", Errorf(pkghttp.StatusBadRequest, "failed to parse event: %v", err)
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		challangeRes, err := responseURLVerification(logger, body)
		if err != nil {
			return "", Errorf(pkghttp.StatusInternalServerError, "failed to response URLVerification: %v", err)
		}
		return challangeRes, nil
	case slackevents.CallbackEvent:
		if err := proxy(ctx, logger, eventsAPIEvent); err != nil {
			return "", Error(pkghttp.StatusInternalServerError, err.Error())
		}
		return "ok", nil
	case slackevents.AppRateLimited:
		return "", Error(pkghttp.StatusBadRequest, "app's event subscriptions are being rate limited")
	}

	return "", Error(pkghttp.StatusBadRequest, "no matched any eventsAPIEvent.Type")
}

func responseURLVerification(logger *zerolog.Logger, body []byte) (string, error) {
	var res *slackevents.ChallengeResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	return res.Challenge, nil
}

func proxy(ctx context.Context, logger *zerolog.Logger, eventsAPIEvent slackevents.EventsAPIEvent) error {
	projectID, err := util.FetchProjectID()
	if err != nil {
		return fmt.Errorf("failed to fetch projectID: %v", err)
	}

	slackBotToken, err := util.FetchSecretLatestVersion(ctx, "slack-bot-token", projectID)
	if err != nil {
		return err
	}
	slackClient := slack.New(slackBotToken)

	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.ReactionAddedEvent:
		e, err := newReactionAddedEvent(ctx, slackClient, event, logger, projectID)
		if err != nil {
			return fmt.Errorf("failed to create ReactionAddedEvent: %w", err)
		}
		logger.Infof("recieved ReactionAddedEvent(user: %s, channel: %s, reaction: %s)", e.userRealName, e.channelName, event.Reaction)

		items, err := route.ParseReactionAddedItem("./routing.yml")
		if err != nil {
			return err
		}

		item, body, err := e.getMatchedItem(ctx, items)
		if err != nil {
			return err
		}

		if err := e.createHTTPTaskWithToken(ctx, body, item); err != nil {
			return err
		}

		return nil
	}

	return nil
}

type ReactionAddedEvent struct {
	userRealName string
	channelName  string
	projectID    string
	locationID   string
	event        *slackevents.ReactionAddedEvent
	logger       *zerolog.Logger
}

func newReactionAddedEvent(ctx context.Context, s *slack.Client, event *slackevents.ReactionAddedEvent, logger *zerolog.Logger, projectID string) (*ReactionAddedEvent, error) {
	sc := trace.FromContext(ctx).SpanContext()
	_, span := trace.StartSpanWithRemoteParent(ctx, "newReactionAddedEvent", sc)
	defer span.End()

	user, err := s.GetUserInfoContext(ctx, event.User)
	if err != nil {
		return nil, err
	}

	channel, err := s.GetConversationInfoContext(ctx, event.Item.Channel, false)
	if err != nil {
		return nil, err
	}

	location, isSet := os.LookupEnv("CLOUD_RUN_LOCATION")
	if !isSet {
		location = "asia-northeast1"
	}

	return &ReactionAddedEvent{
		userRealName: user.RealName,
		channelName:  channel.Name,
		event:        event,
		logger:       logger,
		projectID:    projectID,
		locationID:   location,
	}, nil
}

func (e *ReactionAddedEvent) getMatchedItem(ctx context.Context, items []route.ReactionAddedItem) (*route.ReactionAddedItem, *exampleapi.RequestBody, error) {
	sc := trace.FromContext(ctx).SpanContext()
	_, span := trace.StartSpanWithRemoteParent(ctx, "ReactionAddedEvent.getMatchedItem", sc)
	defer span.End()

	for _, item := range items {
		if ok := item.Match(e.userRealName, e.event.Reaction, e.channelName); ok {
			requestBody := &exampleapi.RequestBody{
				Reaction:      e.event.Reaction,
				User:          e.userRealName,
				ItemChannel:   e.channelName,
				ItemTimestamp: e.event.Item.Timestamp,
			}
			return &item, requestBody, nil
		}
	}

	return nil, nil, errors.New("no matched item")
}

func (e *ReactionAddedEvent) createHTTPTaskWithToken(ctx context.Context, requestBody *exampleapi.RequestBody, item *route.ReactionAddedItem) error {
	sc := trace.FromContext(ctx).SpanContext()
	_, span := trace.StartSpanWithRemoteParent(ctx, "ReactionAddedEvent.createHTTPTaskWithToken", sc)
	defer span.End()

	serviceURL, err := item.FetchServiceURL(ctx, e.projectID, e.locationID)
	if err != nil {
		return err
	}
	e.logger.Debugf("get destination service address : %s", serviceURL)

	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to Marshal RequestBody: %w", err)
	}

	serviceAccountID, isSet := os.LookupEnv("SERVICE_ACCOUNT_ID")
	if !isSet {
		serviceAccountID = "run-invoker"
	}

	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("cloudtasks.NewClient: %w", err)
	}
	defer client.Close()

	traceID, ok := ctx.Value("x-cloud-trace-context").(string)
	if !ok {
		return errors.New("traceID not found")
	}

	queueID := item.ServiceName + "-task-queue"
	req := &taskspb.CreateTaskRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/queues/%s", e.projectID, e.locationID, queueID),
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#HttpRequest
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        serviceURL,
					Headers:    map[string]string{"X-Cloud-Trace-Context": traceID},
					Body:       requestBodyJson,
					AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
						OidcToken: &taskspb.OidcToken{
							ServiceAccountEmail: fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccountID, e.projectID),
						},
					},
				},
			},
		},
	}

	task, err := client.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("cloudtasks.CreateTask: %w", err)
	}
	e.logger.Infof("created task: %s", task.Name)

	return nil
}
