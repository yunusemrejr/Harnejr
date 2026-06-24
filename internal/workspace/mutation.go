package workspace

type ChangeBackupResult struct {
	Result   WriteResult    `json:"result"`
	Backup   SnapshotResult `json:"backup"`
}

func ChangeFileWithBackup(root string, target string, content string, maxBytes int, reason string) (ChangeBackupResult, error) {
	backup, err := SnapshotFile(root, target, reason)
	if err != nil {
		return ChangeBackupResult{}, err
	}
	result, err := Write(root, target, content, maxBytes)
	if err != nil {
		return ChangeBackupResult{}, err
	}
	return ChangeBackupResult{Result: result, Backup: backup}, nil
}
