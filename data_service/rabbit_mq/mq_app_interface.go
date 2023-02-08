package rabbit_mq

type MQApp interface {
	Consumer() int8
	Name() int8
}
