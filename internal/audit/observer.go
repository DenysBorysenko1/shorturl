package audit

type Observer interface {
	LogEvent(event Event) error
	Close() error
}

type Subject interface {
	Attach(observer Observer)
	Detach(observer Observer)
	Notify(event Event)
}
