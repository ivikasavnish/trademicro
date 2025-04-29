package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

const (
	gcpProjectID = "amiable-alcove-456605-k2"
	gcpZone      = "asia-south1-c"
	gcpInstance  = "instance-20250416-112838"
)

// startTradeSystemHandler starts the trading system (big machine)
func startTradeSystemHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	defer client.Close()

	// Prevent controlling the current controller instance
	if isSelfInstance(gcpInstance) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Refusing to start this control server instance.",
		})
		return
	}

	req := &computepb.StartInstanceRequest{
		Project:  gcpProjectID,
		Zone:     gcpZone,
		Instance: gcpInstance,
	}
	_, err = client.Start(ctx, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "trade_system_started",
		"error":  nil,
	})
}

// stopTradeSystemHandler stops the trading system (big machine)
func stopTradeSystemHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	defer client.Close()

	// Prevent controlling the current controller instance
	if isSelfInstance(gcpInstance) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Refusing to stop this control server instance.",
		})
		return
	}

	req := &computepb.StopInstanceRequest{
		Project:  gcpProjectID,
		Zone:     gcpZone,
		Instance: gcpInstance,
	}
	_, err = client.Stop(ctx, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "trade_system_stopped",
		"error":  nil,
	})
}

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

// listGCEInstancesHandler lists all GCE instances in the configured zone (DEBUG ONLY)
func listGCEInstancesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"instances": nil,
			"error":     err.Error(),
		})
		return
	}
	defer client.Close()

	req := &computepb.ListInstancesRequest{
		Project: gcpProjectID,
		Zone:    gcpZone,
	}
	it := client.List(ctx, req)
	var instances []map[string]interface{}
	var iterErr error
	for {
		inst, err := it.Next()
		if err != nil {
			iterErr = err
			break
		}
		item := map[string]interface{}{
			"name":       inst.GetName(),
			"status":     inst.GetStatus(),
			"externalIP": "",
		}
		for _, ni := range inst.GetNetworkInterfaces() {
			for _, ac := range ni.GetAccessConfigs() {
				if ac.GetNatIP() != "" {
					item["externalIP"] = ac.GetNatIP()
				}
			}
		}
		instances = append(instances, item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"instances": instances,
		"error": func() interface{} {
			if iterErr != nil && iterErr.Error() != "iterator done" {
				return iterErr.Error()
			} else {
				return nil
			}
		}(),
	})
}
