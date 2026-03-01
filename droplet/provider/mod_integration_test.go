package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
)

// TestValidatePassword_LUKS_Integration creates a real LUKS volume in a loopback file
// and tests that password validation works correctly against it.
// Requires cryptsetup and root privileges; skipped otherwise.
func TestValidatePassword_LUKS_Integration(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("Skipping LUKS integration test: requires root")
	}

	if _, err := exec.LookPath("cryptsetup"); err != nil {
		t.Skip("Skipping LUKS integration test: cryptsetup not found")
	}

	// Create a temporary file to use as the LUKS volume
	dir := t.TempDir()
	imgFile := filepath.Join(dir, "luks-test.img")

	// Create a 32MB file for the LUKS volume
	if err := exec.Command("dd", "if=/dev/zero", "of="+imgFile, "bs=1M", "count=32").Run(); err != nil {
		t.Fatalf("Failed to create image file: %v", err)
	}

	// Setup a loopback device
	out, err := exec.Command("losetup", "--find", "--show", imgFile).Output()
	if err != nil {
		t.Fatalf("Failed to setup loop device: %v", err)
	}
	loopDev := string(out[:len(out)-1]) // trim newline
	t.Cleanup(func() {
		exec.Command("losetup", "-d", loopDev).Run()
	})

	password := "test-luks-password-12345"

	// Format as LUKS with the test password
	formatCmd := exec.Command("cryptsetup", "luksFormat", "--batch-mode", "--type", "luks2", loopDev)
	formatStdin, err := formatCmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin pipe: %v", err)
	}
	go func() {
		defer formatStdin.Close()
		fmt.Fprint(formatStdin, password)
	}()
	if output, err := formatCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to format LUKS: %v\n%s", err, output)
	}

	// Extract partition name from loop device path (e.g., /dev/loop0 -> loop0)
	partName := filepath.Base(loopDev)

	// Build the kcrypt-style event input
	partition := Partition{
		Name: partName,
		FS:   "crypto_LUKS",
	}
	eventData, _ := json.Marshal(partition)
	event := pluggable.Event{
		Name: "discovery.password",
		Data: string(eventData),
	}
	stdin, _ := json.Marshal(event)

	conf := config.Config{}

	// Test with CORRECT password
	t.Run("correct_password", func(t *testing.T) {
		resp := pluggable.EventResponse{
			Data:  password,
			Error: "",
		}
		result := validatePassword(resp, conf, stdin)
		if !result {
			t.Fatal("Expected validatePassword to return true for correct LUKS password")
		}
	})

	// Test with WRONG password
	t.Run("wrong_password", func(t *testing.T) {
		resp := pluggable.EventResponse{
			Data:  "completely-wrong-password",
			Error: "",
		}
		result := validatePassword(resp, conf, stdin)
		if result {
			t.Fatal("Expected validatePassword to return false for wrong LUKS password")
		}
	})

	// Test with bypass enabled
	t.Run("bypass_enabled", func(t *testing.T) {
		bypassConf := config.Config{
			DebugConfig: config.DebugConfig{
				BypassPasswordTest: true,
			},
		}
		resp := pluggable.EventResponse{
			Data:  "any-password",
			Error: "",
		}
		result := validatePassword(resp, bypassConf, stdin)
		if !result {
			t.Fatal("Expected validatePassword to return true when bypass is enabled")
		}
	})
}
