package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func LogFileOrNil(logname string) *os.File {
	const logDirEnv = "CODE_AGENT_LOGDIR"
	const tmpDirname = "/tmp"
	dirname, exists := os.LookupEnv(logDirEnv)
	if !exists || len(dirname) == 0 {
		redf("environment '%s' absent or set to empty string, using %s instead", logDirEnv, tmpDirname)
		dirname = tmpDirname
	}

	path := filepath.Join(dirname, logname)
	f, err := os.Create(path)
	if err != nil {
		redln("unable to open '%s' for writing", path)
		return nil
	}
	return f
}

func LogFilename(prefix string) string {
	now := time.Now()
	timestamp := now.Format("20060102-1504")
	randomString := GenerateRandomString(now.Format(time.RFC3339Nano), 10)
	return fmt.Sprintf("%s-%s-%s.md", prefix, timestamp, randomString)
}

func GenerateRandomString(t string, length int) string {
	randomBytes := make([]byte, 256)
	rand.Read(randomBytes) // never returns error
	randomBytes = append(randomBytes, ([]byte)(t)...)
	hash := sha256.Sum256(randomBytes)
	return hex.EncodeToString(hash[:length/2])
}

func AppendToLog(logFile *os.File, role string, s string) {
	if logFile == nil {
		return
	}
	if len(role) == 0 {
		role = "unknown"
	}

	var firstErr error = nil
	_, err := logFile.WriteString(fmt.Sprintf("> %s on %s\n\n", role, time.Now().Format(time.RFC1123)))
	if err != nil && firstErr != nil {
		firstErr = err
	}
	_, err = logFile.WriteString(s)
	if err != nil && firstErr != nil {
		firstErr = err
	}
	_, err = logFile.WriteString("\n\n***\n\n")
	if err != nil && firstErr != nil {
		firstErr = err
	}

	if firstErr != nil {
		fmt.Printf("error logging to %s: %s", logFile.Name(), reds(firstErr.Error()))
	}
}
