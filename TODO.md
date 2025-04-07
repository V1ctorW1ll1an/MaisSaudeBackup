[ ] Create new folder in drive to upload backups, the folder name is the day-month-year of the backup
[ ] Upload the backup to the new folder
[ ] Delete the backup of the last day before today, keeps only the backup of the current day

./dbbackup.exe -backup-dir D:\BackupSQLServer\BackupDiscoZ\Automatico -database SCM_TESTE -password Mscardpanorama1 -server 192.168.1.72:1433 -user app_scm -zip-dir D:\BackupSQLServer\BackupDiscoZ\Automatico\zip -log-dir D:\BackupSQLServer\BackupDiscoZ\logs\backup
