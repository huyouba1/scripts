package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	//BackUptime := time.Now().Format("2006-01-02-15-04")
	// GitLab备份目录
	backupDir, err := filepath.Abs("/var/opt/gitlab/backups")
	if err != nil {
		fmt.Println("Failed to get backup directory:", err)
		os.Exit(1)
	}

	// 创建GitLab备份
	cmd := exec.Command("gitlab-rake", "gitlab:backup:create")
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to create GitLab backup:", err)
		os.Exit(1)
	}

	// 获取备份文件名
	backups, err := filepath.Glob(backupDir + "/*_backup.tar")
	if err != nil {
		fmt.Println("Failed to get backup files:", err)
		os.Exit(1)
	}
	if len(backups) == 0 {
		fmt.Println("No backup files created")
		os.Exit(1)
	}
	backupFile := backups[len(backups)-1]
	backupName := filepath.Base(backupFile)
	// 压缩备份文件为tar格式
	//tarCmd := exec.Command("tar", "-cf", backupDir+"/"+backupName+BackUptime+".tar", backupFile)
	//if err := tarCmd.Run(); err != nil {
	//  fmt.Println("Failed to compress backup file:", err)
	//  os.Exit(1)
	//}

	//remoteIP := "10.20.0.124"
	//remoteDir := "/mnt/nas"
	rsyncCmd := exec.Command("mv", backupDir+"/"+backupName, "/mnt/samba_share/")
	if err := rsyncCmd.Run(); err != nil {
		fmt.Println("Failed to sync backup file to remote server:", err)
		os.Exit(1)
	}

	//rmCmd := exec.Command("rm", "-f", backupDir+"/"+backupName)
	//if err := rmCmd.Run(); err != nil {
	//	fmt.Println("Failed to delete local backup files:", err)
	//	os.Exit(1)
	//}
	fmt.Println("GitLab backup completed successfully.")
}
