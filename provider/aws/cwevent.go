package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stelligent/mu/common"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

type eventsManager struct {
	eventsAPI cloudwatcheventsiface.CloudWatchEventsAPI
}

func newEventsManager(sess *session.Session) (common.CWEventsManager, error) {
	log.Debug("Connecting to CloudWatch Logs service")
	eventsAPI := cloudwatchevents.New(sess)

	return &eventsManager{
		eventsAPI: eventsAPI,
	}, nil
}

func (eventsMgr *eventsManager) PutTarget(rule *string, clusters []*cloudwatchevents.Target) {
	eventsAPI := eventsMgr.eventsAPI

	eventsAPI.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule: rule,
		Targets: clusters,
	})
}