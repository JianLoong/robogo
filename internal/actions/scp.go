package actions

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

func scpAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 4 {
		return types.MissingArgsError("scp", 4, len(args))
	}

	// Check for unresolved variables
	if errorResult := validateArgsResolved("scp", args); errorResult != nil {
		return *errorResult
	}

	operation := fmt.Sprintf("%v", args[0])   // "upload" or "download"
	host := fmt.Sprintf("%v", args[1])        // "user@hostname:22" or "hostname:22"
	localPath := fmt.Sprintf("%v", args[2])   // "/path/to/local/file.txt"
	remotePath := fmt.Sprintf("%v", args[3])  // "/remote/path/file.txt"

	// Parse connection details
	username, hostname, port := parseSSHHost(host)
	
	// Override username if provided in options
	if user, ok := options["username"].(string); ok {
		username = user
	}

	// Extract authentication options
	password := ""
	keyPath := ""
	
	if pass, ok := options["password"].(string); ok {
		password = pass
	}
	if key, ok := options["private_key"].(string); ok {
		keyPath = key
	}

	// Extract timeout
	timeout := 30 * time.Second
	if timeoutStr, ok := options["timeout"].(string); ok {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		}
	}

	// Create SSH client
	sshClient, err := createSSHClient(username, hostname, port, password, keyPath, timeout)
	if err != nil {
		return types.RequestError(fmt.Sprintf("SCP SSH connect %s", hostname), err.Error())
	}
	defer sshClient.Close()

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return types.RequestError(fmt.Sprintf("SCP SFTP connect %s", hostname), err.Error())
	}
	defer sftpClient.Close()

	switch operation {
	case "upload":
		return performSCPUpload(sftpClient, localPath, remotePath)
	case "download":
		return performSCPDownload(sftpClient, localPath, remotePath)
	default:
		return types.InvalidArgError("scp", "operation", 
			"first argument must be 'upload' or 'download'")
	}
}

func parseSSHHost(host string) (username, hostname, port string) {
	// Default values
	username = "root"
	port = "22"
	
	// Parse user@hostname:port format
	parts := strings.Split(host, "@")
	if len(parts) == 2 {
		username = parts[0]
		host = parts[1]
	}
	
	// Parse hostname:port
	hostParts := strings.Split(host, ":")
	hostname = hostParts[0]
	if len(hostParts) == 2 {
		port = hostParts[1]
	}
	
	return username, hostname, port
}

func createSSHClient(username, hostname, port, password, keyPath string, timeout time.Duration) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, use proper host key verification
		Timeout:         timeout,
	}

	// Authentication methods
	if keyPath != "" {
		// Private key authentication
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}

		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if password != "" {
		// Password authentication
		config.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		return nil, fmt.Errorf("no authentication method provided (password or private_key required)")
	}

	// Connect
	addr := fmt.Sprintf("%s:%s", hostname, port)
	return ssh.Dial("tcp", addr, config)
}

func performSCPUpload(client *sftp.Client, localPath, remotePath string) types.ActionResult {
	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return types.RequestError(fmt.Sprintf("SCP upload open %s", localPath), err.Error())
	}
	defer localFile.Close()

	// Create remote directory if needed
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		client.MkdirAll(remoteDir) // Ignore error - directory might exist
	}

	// Create remote file
	remoteFile, err := client.Create(remotePath)
	if err != nil {
		return types.RequestError(fmt.Sprintf("SCP upload create %s", remotePath), err.Error())
	}
	defer remoteFile.Close()

	// Copy file content
	size, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return types.RequestError(fmt.Sprintf("SCP upload copy %s", remotePath), err.Error())
	}

	// Get file info
	fileInfo, _ := localFile.Stat()
	result := map[string]any{
		"operation":   "upload",
		"local_path":  localPath,
		"remote_path": remotePath,
		"size":        size,
		"permissions": fileInfo.Mode().String(),
		"uploaded_at": time.Now().Format(time.RFC3339),
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   result,
	}
}

func performSCPDownload(client *sftp.Client, localPath, remotePath string) types.ActionResult {
	// Create local directory if needed
	localDir := filepath.Dir(localPath)
	if localDir != "." {
		os.MkdirAll(localDir, 0755)
	}

	// Open remote file
	remoteFile, err := client.Open(remotePath)
	if err != nil {
		return types.RequestError(fmt.Sprintf("SCP download open %s", remotePath), err.Error())
	}
	defer remoteFile.Close()

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return types.RequestError(fmt.Sprintf("SCP download create %s", localPath), err.Error())
	}
	defer localFile.Close()

	// Copy file content
	size, err := io.Copy(localFile, remoteFile)
	if err != nil {
		return types.RequestError(fmt.Sprintf("SCP download copy %s", localPath), err.Error())
	}

	// Get remote file info
	remoteInfo, _ := remoteFile.Stat()
	result := map[string]any{
		"operation":      "download",
		"local_path":     localPath,
		"remote_path":    remotePath,
		"size":           size,
		"permissions":    remoteInfo.Mode().String(),
		"downloaded_at":  time.Now().Format(time.RFC3339),
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data:   result,
	}
}