package common

// CWEventCreator for creating CloudWatch events
type CWEventCreator interface {
	CreateEvent()
	DeleteEvent()
	EditEvent()
	PutTarget(event string, cluster string)
}

// CWEventsManager composite of all events capabilities
type CWEventsManager interface {
	CWEventCreator
}