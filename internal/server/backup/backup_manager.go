package backup

// BackupManager defines the methods required for performing backup-related operations.
// Any type that implements these methods can be considered a "BackupManager".
type BackupManager interface {
	// Start begins the backup process. It may trigger operations to store data
	// in a safe location, ensuring that the backup completes successfully.
	// It is expected to handle any errors related to initiating the backup process.
	Start()

	// StopBackup halts an ongoing backup process.
	// It is typically used to safely stop the backup if necessary, for example, when
	// an error occurs or the process is no longer needed.
	// It should handle any interruptions and ensure that no data is left in an inconsistent state.
	Stop()

	// Restore retrieves data from a backup and restores it to its previous state.
	// This function is expected to deal with errors related to restoration, such as
	// restoring from an incomplete or corrupt backup.
	Restore()
}
