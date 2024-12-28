package common

import "fmt"

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

// Subscribe adds the given observer to the subject's list of observers.
// It calls the subject's RegisterObserver method and returns an error if the subscription fails.
func Subscribe(observer Observer, subject ObserveSubject) error {
	if err := subject.RegisterObserver(observer); err != nil {
		return fmt.Errorf("failed to subscribe observer to subject: %w", err)
	}
	return nil
}
