package provider

import (
	"encoding/json"
	"testing"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
)

func TestGetDevice_ValidInput(t *testing.T) {
	partition := Partition{
		Name:            "sda2",
		FilesystemLabel: "COS_PERSISTENT",
		FS:              "crypto_LUKS",
	}

	eventData, err := json.Marshal(partition)
	if err != nil {
		t.Fatalf("Failed to marshal partition: %v", err)
	}

	event := pluggable.Event{
		Name: "discovery.password",
		Data: string(eventData),
	}

	stdin, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	device := getDevice(stdin)
	if device != "/dev/sda2" {
		t.Fatalf("Expected /dev/sda2, got %q", device)
	}
}

func TestGetDevice_EmptyInput(t *testing.T) {
	device := getDevice([]byte(""))
	if device != "" {
		t.Fatalf("Expected empty device, got %q", device)
	}
}

func TestGetDevice_InvalidJSON(t *testing.T) {
	device := getDevice([]byte("{{{invalid"))
	if device != "" {
		t.Fatalf("Expected empty device for invalid JSON, got %q", device)
	}
}

func TestGetDevice_EmptyEvent(t *testing.T) {
	event := pluggable.Event{
		Name: "discovery.password",
		Data: "{}",
	}

	stdin, _ := json.Marshal(event)
	device := getDevice(stdin)
	// Empty partition name should give "/dev/"
	if device != "/dev/" {
		t.Fatalf("Expected '/dev/', got %q", device)
	}
}

func TestGetDevice_NvmePartition(t *testing.T) {
	partition := Partition{
		Name: "nvme0n1p3",
		FS:   "crypto_LUKS",
	}

	eventData, _ := json.Marshal(partition)
	event := pluggable.Event{
		Name: "discovery.password",
		Data: string(eventData),
	}

	stdin, _ := json.Marshal(event)
	device := getDevice(stdin)
	if device != "/dev/nvme0n1p3" {
		t.Fatalf("Expected /dev/nvme0n1p3, got %q", device)
	}
}

func TestValidatePassword_BypassTest(t *testing.T) {
	conf := config.Config{
		DebugConfig: config.DebugConfig{
			BypassPasswordTest: true,
		},
	}

	event := pluggable.EventResponse{
		Data:  "any-password",
		Error: "",
	}

	result := validatePassword(event, conf, []byte("{}"))
	if !result {
		t.Fatal("Expected validatePassword to return true when bypass is enabled")
	}
}

func TestValidatePassword_EventError(t *testing.T) {
	conf := config.Config{}

	event := pluggable.EventResponse{
		Data:  "",
		Error: "some error occurred",
	}

	result := validatePassword(event, conf, []byte("{}"))
	if result {
		t.Fatal("Expected validatePassword to return false when event has error")
	}
}

func TestValidatePassword_NoDevice(t *testing.T) {
	conf := config.Config{}

	event := pluggable.EventResponse{
		Data:  "test-password",
		Error: "",
	}

	// Invalid stdin means no device, which should return true (skip validation)
	result := validatePassword(event, conf, []byte("not valid json"))
	if !result {
		t.Fatal("Expected validatePassword to return true when no device found (skips validation)")
	}
}

func TestValidatePassword_CryptsetupNotFound(t *testing.T) {
	conf := config.Config{}

	partition := Partition{
		Name: "sda2",
		FS:   "crypto_LUKS",
	}

	eventData, _ := json.Marshal(partition)
	event := pluggable.Event{
		Name: "discovery.password",
		Data: string(eventData),
	}

	stdin, _ := json.Marshal(event)

	resp := pluggable.EventResponse{
		Data:  "test-password",
		Error: "",
	}

	// When cryptsetup is not installed, the exec will fail but not with exit code 2
	// so it should return true (allows password through)
	result := validatePassword(resp, conf, stdin)
	if !result {
		t.Fatal("Expected validatePassword to return true when cryptsetup not available (allows password)")
	}
}
