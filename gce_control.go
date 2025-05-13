package main

import (
	"os"
)

const (
	gcpProjectID = "amiable-alcove-456605-k2"
	gcpZone      = "asia-south1-c"
	gcpInstance  = "instance-20250416-112838"
)

// isSelfInstance returns true if the requested instance name is the current controller instance
func isSelfInstance(targetInstance string) bool {
	// Hardcoded controller instance name and external IP
	if targetInstance == "instance-20250422-132526" || targetInstance == "35.244.30.157" {
		return true
	}
	hostname, err := os.Hostname()
	if err == nil && targetInstance == hostname {
		return true
	}
	return false
}
