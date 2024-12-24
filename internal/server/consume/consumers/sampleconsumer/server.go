package sampleconsumer

type SampleConsumer struct{}

func (s *SampleConsumer) StartConsume() error {
	panic("it`s sample consumer, it no released some logic")
}
