package backup

// Manager defines the methods required for performing backup-related operations.
//
// Any type that implements these methods can be considered a "Manager" capable of handling
// backup operations such as starting, stopping, and restoring backups.
type Manager interface {
	// Start begins the backup process.
	//
	// This method triggers the operations necessary to store data in a secure location,
	// ensuring the backup process completes successfully. Implementations should handle
	// any errors that occur while initiating or performing the backup process.
	Start()

	// Stop halts an ongoing backup process.
	//
	// This method is used to safely terminate a backup operation when necessary, such as
	// in response to an error or a manual shutdown. It should ensure that no data is left
	// in an inconsistent or incomplete state.
	Stop()

	// Restore retrieves data from a backup and restores it to its previous state.
	//
	// This method handles the restoration process, including scenarios involving incomplete
	// or corrupted backups. Implementations should manage errors gracefully and ensure
	// that the data is restored as accurately as possible.
	Restore()
}

// ManagerAbstractFactory provides a method to create instances of Manager.
//
// Any implementation of this interface acts as a factory for creating Manager instances
// with specific configurations or behaviors.
type ManagerAbstractFactory interface {
	// CreateManager creates and returns a new Manager instance.
	//
	// Returns:
	//   - A new instance of a type that implements the Manager interface.
	CreateManager() Manager
}
