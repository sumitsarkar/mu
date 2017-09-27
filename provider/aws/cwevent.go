package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchevents/cloudwatcheventsiface"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stelligent/mu/common"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/aws"
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

func (eventsMgr *eventsManager) CreateEvent() {
	eventsAPI := eventsMgr.eventsAPI

	eventsAPI.PutEvents(&cloudwatchevents.PutEventsInput{
		Entries: []*cloudwatchevents.PutEventsRequestEntry{
			&cloudwatchevents.PutEventsRequestEntry{
				Detail: aws.String(""),
				DetailType: aws.String(""),
				Resources: []*string{
					aws.String(""),
				},
				Source: aws.String(""),
			},
		},
	})
}

func (eventsMgr *eventsManager) DeleteEvent() {

}

func (eventsMgr *eventsManager) EditEvent() {

}

func (eventsMgr *eventsManager) PutTarget(event string, custer string) {
	eventsAPI := eventsMgr.eventsAPI

	eventsAPI.PutTargets(&cloudwatchevents.PutTargetsInput{

	})
}