package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
)

type eventsManager struct {
	eventsAPI cloudwatcheventsiface.CloudWatchEventsAPI
}

func newEventsManager(sess *session.Session)