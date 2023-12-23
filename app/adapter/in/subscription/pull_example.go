package subscription

import (
	"context"
	"encoding/json"
	"net/http"

	"einar/app/business"
	"einar/app/domain/port/in"
	"einar/app/infrastructure/broker"
	"einar/app/infrastructure/broker/subscription"
	"einar/app/shared/constants"
	"einar/app/shared/container"
	"einar/app/shared/errors"

	"cloud.google.com/go/pubsub"
)

var pullExampleSetup = container.InjectInboundAdapter(func() (pullExampleSubscription, error) {
	instance := pullExampleSubscription{
		ExampleUseCase: business.Example,
	}
	var subscriptionName = "INSERT YOUR SUBSCRIPTION NAME"
	subRef := broker.Client().Subscription(subscriptionName)
	subRef.ReceiveSettings.MaxOutstandingMessages = 5
	settings := subRef.Receive
	go subscription.
		New(subscriptionName, instance.pull, settings).
		WithPushHandler(constants.DefaultPushHandlerPrefix + subscriptionName).
		Start()
	return instance, nil
})

type pullExampleSubscription struct {
	ExampleUseCase in.Example
}

func (p pullExampleSubscription) pull(ctx context.Context, subscriptionName string, m *pubsub.Message) (statusCode int, err error) {
	var dataModel interface{}
	defer func() {
		subscription.HandleMessageAcknowledgement(ctx, &subscription.HandleMessageAcknowledgementDetails{
			MessageID:        m.ID,
			PublishTime:      m.PublishTime.String(),
			SubscriptionName: subscriptionName,
			Error:            err,
			Message:          m,
			ErrorsRequiringNack: []error{
				errors.INTERNAL_SERVER_ERROR,
				errors.EXTERNAL_SERVER_ERROR,
				errors.HTTP_NETWORK_ERROR,
				errors.PUBSUB_BROKER_ERROR,
			},
			CustomLogFields: map[string]interface{}{
				"dataModel": dataModel,
			},
		})
	}()
	if err := json.Unmarshal(m.Data, &dataModel); err != nil {
		return http.StatusBadRequest, err
	}
	return http.StatusOK, nil
}
