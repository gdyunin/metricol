package patterns

import (
	"errors"
	"fmt"
)

// Observer defines the interface for an observer in the observer pattern.
// It requires the implementation of the OnNotify method, which is called
// when the observer is notified of a change.
type Observer interface {
	OnNotify()
}

// ObserveSubject defines the interface for a subject in the observer pattern.
// It includes methods for registering, removing, and notifying observers.
type ObserveSubject interface {
	RegisterObserver(observer Observer) error
	RemoveObserver(observer Observer) error
	NotifyObservers()
}

func Subscribe(observer any, subject any) error {
	// Verify the collectors implements the Observer interface.
	o, ok := observer.(Observer)
	if !ok {
		return errors.New("collectors does not implement the Observer interface")
	}

	// Verify the producer implements the ObserveSubject interface.
	s, ok := subject.(ObserveSubject)
	if !ok {
		return errors.New("producer does not implement the ObserveSubject interface")
	}

	if err := s.RegisterObserver(o); err != nil {
		return fmt.Errorf("failed to subscribe observer to subject: %w", err)
	}

	return nil
}
