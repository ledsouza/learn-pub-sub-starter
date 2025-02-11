package pubsub

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)
