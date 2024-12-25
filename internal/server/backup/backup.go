package backup

type Backupper interface {
	StartBackup()
	StopBackup()
}
