package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	logging "github.com/ops-k/go-lib-logging"
)

type InputPayload struct {
	SnsTopicArn string `json:"snsTopicArn"`
	Subject     string `json:"subject"`
	Message     string `json:"message"`
}

var (
	logger    = logging.GetLogger()
	snsClient *sns.Client
)

func init() {
	// Initialize the S3 client outside of the handler, during the init phase
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to load SDK config")
	}

	snsClient = sns.NewFromConfig(cfg)
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event json.RawMessage) error {
	var inputPayload InputPayload
	if err := json.Unmarshal(event, &inputPayload); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal event")
		return err
	}

	// Access environment variables
	snsTopicArn := inputPayload.SnsTopicArn
	if snsTopicArn == "" {
		logger.Debug().Msg("sns topic arn not provided in payload, checking environment variable.")
		snsTopicArn = os.Getenv("SNS_TOPIC_ARN")
		if snsTopicArn == "" {
			logger.Error().Msg("no sns topic arn configured in environment or in payload.")
			return fmt.Errorf("missing required sns topic arb in environment variable SNS_TOPIC_ARN or in payload")
		}
		logger.Debug().Msgf("sns topic arn provided in environment variable: %s", snsTopicArn)
	} else {
		logger.Debug().Msgf("sns topic arn provided in payload: %s", snsTopicArn)
	}

	// Publish message to SNS topic
	logger.Info().Msgf("publishing message to sns topic: %s", snsTopicArn)
	_, err := snsClient.Publish(ctx, &sns.PublishInput{
		Subject:  aws.String(inputPayload.Subject),
		Message:  aws.String(inputPayload.Message),
		TopicArn: aws.String(snsTopicArn),
	})
	if err != nil {
		logger.Error().Err(err).Msgf("failed to publish message to sns topic %s", snsTopicArn)
		return err
	}

	logger.Info().Msg("message sent successfully")

	return nil
}

func main() {
	if os.Getenv("LAMBDA_TASK_ROOT") != "" {
		lambda.Start(Handler)
	} else {
		Handler(context.Background(), []byte(`{"subject": "hello", "message": "world!"}`))
	}
}
