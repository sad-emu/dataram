package main

import (
	"crypto/tls"
	"os"
	"path/filepath"
)

func setupSettings() {
	GlobalSettings.StorageDir = "./test_storage"
	RemoveAllFilesInDir(GlobalSettings.StorageDir)

	// GlobalSettings.MaxConcurrentTransfers = 10
	// GlobalSettings.MaxTransferSize = 100 * 1024 * 1024 // 100MB
	// GlobalSettings.MaxChunkSize = 32 * 1024            // 32KB
}

func makeTestTLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic("failed to load test TLS cert: " + err.Error())
	}
	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		NextProtos:         []string{"dataram"},
		ServerName:         "localhost",
	}
}

func RemoveAllFilesInDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			err := os.Remove(filepath.Join(dir, entry.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
