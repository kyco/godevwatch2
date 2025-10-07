package build

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Tracker manages build status files
type Tracker struct {
	statusDir string
	buildID   string
}

// NewTracker creates a new build tracker
func NewTracker(statusDir string) *Tracker {
	return &Tracker{
		statusDir: statusDir,
	}
}

// generateBuildID creates a unique build ID with timestamp
func (t *Tracker) generateBuildID() string {
	// Format: YYYYMMDD-HHMMSS-milliseconds
	now := time.Now()
	return fmt.Sprintf("%s-%03d", now.Format("20060102-150405"), now.Nanosecond()/1000000)
}

// Start marks the beginning of a build
func (t *Tracker) Start() error {
	// Ensure status directory exists
	if err := os.MkdirAll(t.statusDir, 0755); err != nil {
		return fmt.Errorf("failed to create status directory: %w", err)
	}

	// Generate new build ID
	t.buildID = t.generateBuildID()
	fmt.Printf("[build] Build ID: %s\n", t.buildID)

	// Write current build ID
	currentBuildIDPath := filepath.Join(t.statusDir, "current-build-id")
	if err := os.WriteFile(currentBuildIDPath, []byte(t.buildID), 0644); err != nil {
		return fmt.Errorf("failed to write current-build-id: %w", err)
	}
	fmt.Printf("[build] Created %s\n", filepath.Join(t.statusDir, "current-build-id"))

	// Create building marker file
	buildingMarkerPath := filepath.Join(t.statusDir, fmt.Sprintf("%s-building", t.buildID))
	if err := os.WriteFile(buildingMarkerPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write building marker: %w", err)
	}
	fmt.Printf("[build] Created %s\n", filepath.Join(t.statusDir, fmt.Sprintf("%s-building", t.buildID)))

	return nil
}

// Complete marks the successful completion of a build
func (t *Tracker) Complete() error {
	// Remove all files except current-build-id
	entries, err := os.ReadDir(t.statusDir)
	if err != nil {
		return fmt.Errorf("failed to read status directory: %w", err)
	}

	fmt.Printf("[build] Cleaning up status directory\n")
	for _, entry := range entries {
		if entry.Name() != "current-build-id" {
			path := filepath.Join(t.statusDir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
			}
			fmt.Printf("[build] Removed %s\n", path)
		}
	}

	// Write last-success-build-id
	lastSuccessPath := filepath.Join(t.statusDir, "last-success-build-id")
	if err := os.WriteFile(lastSuccessPath, []byte(t.buildID), 0644); err != nil {
		return fmt.Errorf("failed to write last-success-build-id: %w", err)
	}
	fmt.Printf("[build] Created %s\n", lastSuccessPath)

	return nil
}

// Fail marks a build as failed
func (t *Tracker) Fail() error {
	fmt.Printf("[build] Marking build as failed\n")

	// Remove the building marker
	buildingMarkerPath := filepath.Join(t.statusDir, fmt.Sprintf("%s-building", t.buildID))
	if err := os.Remove(buildingMarkerPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove building marker: %w", err)
	}
	fmt.Printf("[build] Removed %s\n", buildingMarkerPath)

	// Create failed marker file
	failedMarkerPath := filepath.Join(t.statusDir, fmt.Sprintf("%s-failed", t.buildID))
	if err := os.WriteFile(failedMarkerPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to write failed marker: %w", err)
	}
	fmt.Printf("[build] Created %s\n", failedMarkerPath)

	return nil
}
