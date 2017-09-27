package common

import "github.com/aws/aws-sdk-go/service/cloudwatchevents"

// CWPutEventTarget for creating CloudWatch events
type CWPutEventTarget interface {
	PutTarget(rule *string, clusters []*cloudwatchevents.Target)
}

// CWEventsManager composite of all events capabilities
type CWEventsManager interface {
	CWPutEventTarget
}