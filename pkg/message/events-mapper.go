package message

import goeh "github.com/hetacode/go-eh"

func NewEventsMapper() *goeh.EventsMapper {
	eventsMapper := new(goeh.EventsMapper)
	eventsMapper.Register(new(AgentRegistrationMessage))

	return eventsMapper
}
