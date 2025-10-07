package build

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Tracker manages build status files
type Tracker struct {
	statusDir      string
	buildID        string
	startTimestamp int64
	debugMode      bool
}

// NewTracker creates a new build tracker
func NewTracker(statusDir string, debugMode bool) *Tracker {
	return &Tracker{
		statusDir: statusDir,
		debugMode: debugMode,
	}
}

// generateBuildID creates a unique build ID string
func (t *Tracker) generateBuildID() string {
	// Generate 4 random bytes and encode as hex for a unique ID (8 characters)
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Start marks the beginning of a build
func (t *Tracker) Start() error {
	// Ensure status directory exists
	if err := os.MkdirAll(t.statusDir, 0755); err != nil {
		return fmt.Errorf("failed to create status directory: %w", err)
	}

	// Generate new build ID and capture start timestamp
	t.buildID = t.generateBuildID()
	t.startTimestamp = time.Now().Unix()
	fmt.Printf("[build] Build ID: %s (start timestamp: %d)\n", t.buildID, t.startTimestamp)

	// Write current build ID
	currentBuildIDPath := filepath.Join(t.statusDir, "current-build-id")
	if err := os.WriteFile(currentBuildIDPath, []byte(t.buildID), 0644); err != nil {
		return fmt.Errorf("failed to write current-build-id: %w", err)
	}
	fmt.Printf("[build] Created %s\n", filepath.Join(t.statusDir, "current-build-id"))

	// Create building marker file with actual start timestamp
	buildingMarkerPath := filepath.Join(t.statusDir, fmt.Sprintf("%d-%s-building", t.startTimestamp, t.buildID))
	if err := os.WriteFile(buildingMarkerPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write building marker: %w", err)
	}
	fmt.Printf("[build] Created %s\n", buildingMarkerPath)

	return nil
}

// Complete marks the successful completion of a build
func (t *Tracker) Complete() error {
	// Capture completion timestamp at the exact moment of success
	completionTimestamp := time.Now().Unix()
	successMarkerPath := filepath.Join(t.statusDir, fmt.Sprintf("%d-%s-success", completionTimestamp, t.buildID))
	if err := os.WriteFile(successMarkerPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write success marker: %w", err)
	}
	fmt.Printf("[build] Created %s (completion timestamp: %d)\n", successMarkerPath, completionTimestamp)

	// Write last-success-build-id
	lastSuccessPath := filepath.Join(t.statusDir, "last-success-build-id")
	if err := os.WriteFile(lastSuccessPath, []byte(t.buildID), 0644); err != nil {
		return fmt.Errorf("failed to write last-success-build-id: %w", err)
	}
	fmt.Printf("[build] Created %s\n", lastSuccessPath)

	// Keep all build ID status files for audit purposes
	fmt.Printf("[build] Preserving all build status files for audit\n")

	return nil
}

// Fail marks a build as failed
func (t *Tracker) Fail() error {
	fmt.Printf("[build] Marking build as failed\n")

	// Capture failure timestamp at the exact moment of failure
	failureTimestamp := time.Now().Unix()
	failedMarkerPath := filepath.Join(t.statusDir, fmt.Sprintf("%d-%s-failed", failureTimestamp, t.buildID))
	if err := os.WriteFile(failedMarkerPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write failed marker: %w", err)
	}
	fmt.Printf("[build] Created %s (failure timestamp: %d)\n", failedMarkerPath, failureTimestamp)

	// Note: We keep the building marker file for audit purposes
	fmt.Printf("[build] Preserving building marker for audit\n")

	return nil
}

// Abort marks a build as aborted
func (t *Tracker) Abort() error {
	fmt.Printf("[build] Marking build as aborted\n")

	// Capture abort timestamp at the exact moment of abortion
	abortTimestamp := time.Now().Unix()
	abortedMarkerPath := filepath.Join(t.statusDir, fmt.Sprintf("%d-%s-aborted", abortTimestamp, t.buildID))
	if err := os.WriteFile(abortedMarkerPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write aborted marker: %w", err)
	}
	fmt.Printf("[build] Created %s (abort timestamp: %d)\n", abortedMarkerPath, abortTimestamp)

	// Note: We keep the building marker file for audit purposes
	fmt.Printf("[build] Preserving building marker for audit\n")

	return nil
}

// GetBuildID returns the current build ID
func (t *Tracker) GetBuildID() string {
	return t.buildID
}
