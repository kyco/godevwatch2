package watcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kyco/godevwatch/internal/build"
	"github.com/kyco/godevwatch/internal/config"
	"github.com/kyco/godevwatch/internal/logger"
)

// Watcher manages file watching and build execution
type Watcher struct {
	config       *config.Config
	fsWatcher    *fsnotify.Watcher
	buildTracker *build.Tracker

	// Process management
	mu            sync.RWMutex
	runningBuilds map[string]*RunningBuild // rule name -> running build

	// Debouncing
	debounceTimer map[string]*time.Timer // rule name -> timer
	debounceMu    sync.Mutex
	debounceDelay time.Duration

	// Callbacks
	buildSuccessCallback func()
}

// RunningBuild tracks a currently executing build process
type RunningBuild struct {
	Rule    *config.BuildRule
	Process *exec.Cmd
	Tracker *build.Tracker
	Cancel  context.CancelFunc
	BuildID string
}

// NewWatcher creates a new file watcher
func NewWatcher(cfg *config.Config) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fs watcher: %w", err)
	}

	return &Watcher{
		config:        cfg,
		fsWatcher:     fsWatcher,
		runningBuilds: make(map[string]*RunningBuild),
		debounceTimer: make(map[string]*time.Timer),
		debounceDelay: 100 * time.Millisecond, // 100ms debounce
	}, nil
}

// Start begins watching files and handling changes
func (w *Watcher) Start(ctx context.Context) error {
	// Add all watch patterns to the file system watcher
	if err := w.setupWatchers(); err != nil {
		return fmt.Errorf("failed to setup watchers: %w", err)
	}

	fmt.Printf("[watcher] Started watching files\n")

	// Main event loop
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[watcher] Stopping watcher\n")
			w.stopAllBuilds()
			return w.fsWatcher.Close()

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return fmt.Errorf("watcher events channel closed")
			}
			w.handleFileEvent(event)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			fmt.Printf("[watcher] Error: %v\n", err)
		}
	}
}

// setupWatchers adds all directories that need to be watched
func (w *Watcher) setupWatchers() error {
	watchedDirs := make(map[string]bool)

	for _, rule := range w.config.BuildRules {
		for _, pattern := range rule.Watch {
			dirs, err := w.getDirectoriesToWatch(pattern)
			if err != nil {
				return fmt.Errorf("failed to get directories for pattern %s: %w", pattern, err)
			}

			for _, dir := range dirs {
				// Skip directories that match ignore patterns
				if w.shouldIgnoreDirectory(dir, &rule) {
					continue
				}

				if !watchedDirs[dir] {
					if err := w.fsWatcher.Add(dir); err != nil {
						return fmt.Errorf("failed to watch directory %s: %w", dir, err)
					}
					watchedDirs[dir] = true
					fmt.Printf("[watcher] Watching directory: %s\n", dir)
				}
			}
		}
	}

	return nil
}

// getDirectoriesToWatch extracts directories from glob patterns
func (w *Watcher) getDirectoriesToWatch(pattern string) ([]string, error) {
	var dirs []string

	// Handle recursive patterns like **/*.go
	if strings.Contains(pattern, "**") {
		// Add current directory and walk subdirectories
		dirs = append(dirs, ".")

		err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && !strings.HasPrefix(path, ".git") {
				dirs = append(dirs, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// For simple patterns, watch the directory containing the files
		dir := filepath.Dir(pattern)
		if dir == "." || dir == "" {
			dirs = append(dirs, ".")
		} else {
			dirs = append(dirs, dir)
		}
	}

	return dirs, nil
}

// handleFileEvent processes file system events
func (w *Watcher) handleFileEvent(event fsnotify.Event) {
	// Skip temporary files and hidden files
	if strings.HasPrefix(filepath.Base(event.Name), ".") ||
		strings.HasSuffix(event.Name, "~") ||
		strings.Contains(event.Name, ".tmp") {
		return
	}

	// Skip files that match ignore patterns for any rule
	if w.shouldIgnoreFile(event.Name) {
		return
	}

	// Only handle write and create events
	if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
		return
	}

	fmt.Printf("[watcher] File changed: %s\n", event.Name)

	// Check which build rules should be triggered
	for i := range w.config.BuildRules {
		rule := &w.config.BuildRules[i]
		if w.shouldTriggerBuild(event.Name, rule) {
			w.debounceBuild(rule)
		}
	}
}

// shouldTriggerBuild checks if a file change should trigger a build rule
func (w *Watcher) shouldTriggerBuild(filename string, rule *config.BuildRule) bool {
	relativePath, err := filepath.Rel(".", filename)
	if err != nil {
		relativePath = filename
	}

	for _, pattern := range rule.Watch {
		if w.matchesPattern(relativePath, pattern) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a file path matches a glob pattern (including ** support)
func (w *Watcher) matchesPattern(path, pattern string) bool {
	// Handle simple patterns without **
	if !strings.Contains(pattern, "**") {
		matched, err := filepath.Match(pattern, path)
		return err == nil && matched
	}

	// Handle ** patterns - split on ** and match each part
	parts := strings.Split(pattern, "**/")
	if len(parts) == 1 {
		// Pattern ends with ** (like "src/**")
		prefix := strings.TrimSuffix(parts[0], "**")
		return strings.HasPrefix(path, prefix)
	}

	// Pattern has ** in the middle (like "**/**.go")
	if len(parts) == 2 {
		prefix := parts[0]
		suffix := parts[1]

		// Check if path starts with prefix (if any) and ends with suffix pattern
		if prefix != "" && !strings.HasPrefix(path, prefix) {
			return false
		}

		// Check if the remaining path matches the suffix pattern
		pathToCheck := path
		if prefix != "" {
			pathToCheck = strings.TrimPrefix(path, prefix)
			pathToCheck = strings.TrimPrefix(pathToCheck, "/")
		}

		// Handle suffix pattern matching
		if suffix == "" {
			return true // Pattern is just "prefix/**"
		}

		// Check if any part of the path matches the suffix
		pathParts := strings.Split(pathToCheck, string(filepath.Separator))
		for i := range pathParts {
			subPath := filepath.Join(pathParts[i:]...)
			matched, err := filepath.Match(suffix, subPath)
			if err == nil && matched {
				return true
			}
		}

		// Also check the full remaining path
		matched, err := filepath.Match(suffix, pathToCheck)
		return err == nil && matched
	}

	// More complex ** patterns - fallback to simple matching
	simplePattern := strings.ReplaceAll(pattern, "**/", "")
	matched, err := filepath.Match(simplePattern, filepath.Base(path))
	return err == nil && matched
}

// debounceBuild implements debouncing to avoid rapid successive builds
func (w *Watcher) debounceBuild(rule *config.BuildRule) {
	w.debounceMu.Lock()
	defer w.debounceMu.Unlock()

	// Cancel existing timer for this rule
	if timer, exists := w.debounceTimer[rule.Name]; exists {
		timer.Stop()
	}

	// Set new timer
	w.debounceTimer[rule.Name] = time.AfterFunc(w.debounceDelay, func() {
		w.executeBuild(rule)
	})
}

// executeBuild runs a build rule, aborting any existing build for the same rule
func (w *Watcher) executeBuild(rule *config.BuildRule) {
	w.mu.Lock()
	defer w.mu.Unlock()

	fmt.Printf("[watcher] Triggering build: %s\n", rule.Name)

	// Check if there's already a running build for this rule
	if runningBuild, exists := w.runningBuilds[rule.Name]; exists {
		fmt.Printf("[watcher] Aborting previous build: %s\n", rule.Name)
		w.abortBuild(runningBuild)
	}

	// Start new build
	ctx, cancel := context.WithCancel(context.Background())
	tracker := build.NewTracker(w.config.BuildStatusDir, w.config.DebugMode)

	// Start tracking
	if err := tracker.Start(); err != nil {
		fmt.Printf("[watcher] Failed to start build tracking: %v\n", err)
		cancel()
		return
	}

	// Create command
	cmd := exec.CommandContext(ctx, "sh", "-c", rule.Command)
	cmd.Stdout = logger.NewPrefixWriter(fmt.Sprintf("[build:%s] ", rule.Name), os.Stdout)
	cmd.Stderr = logger.NewPrefixWriter(fmt.Sprintf("[build:%s] ", rule.Name), os.Stderr)

	runningBuild := &RunningBuild{
		Rule:    rule,
		Process: cmd,
		Tracker: tracker,
		Cancel:  cancel,
		BuildID: tracker.GetBuildID(),
	}

	w.runningBuilds[rule.Name] = runningBuild

	// Start the build process
	go w.runBuildProcess(runningBuild)
}

// runBuildProcess executes the build in a goroutine
func (w *Watcher) runBuildProcess(rb *RunningBuild) {
	defer func() {
		w.mu.Lock()
		delete(w.runningBuilds, rb.Rule.Name)
		w.mu.Unlock()
		rb.Cancel()
	}()

	// Run the command
	err := rb.Process.Run()

	if err != nil {
		// Check if it was canceled (aborted)
		if rb.Process.ProcessState != nil && rb.Process.ProcessState.Exited() {
			if exitError, ok := err.(*exec.ExitError); ok {
				// Check if process was killed (aborted)
				if exitError.ExitCode() == -1 ||
					(exitError.ProcessState.Sys().(syscall.WaitStatus)).Signal() == syscall.SIGKILL ||
					(exitError.ProcessState.Sys().(syscall.WaitStatus)).Signal() == syscall.SIGTERM {
					// This was an abort, not a failure
					return
				}
			}
		}

		// This was a genuine failure
		fmt.Printf("[watcher] Build failed: %s - %v\n", rb.Rule.Name, err)
		if err := rb.Tracker.Fail(); err != nil {
			fmt.Printf("[watcher] Failed to mark build as failed: %v\n", err)
		}
		return
	}

	// Build succeeded
	fmt.Printf("[watcher] Build completed: %s\n", rb.Rule.Name)
	if err := rb.Tracker.Complete(); err != nil {
		fmt.Printf("[watcher] Failed to mark build as complete: %v\n", err)
	}

	// Call success callback if set
	if w.buildSuccessCallback != nil {
		w.buildSuccessCallback()
	}
}

// abortBuild terminates a running build and marks it as aborted
func (w *Watcher) abortBuild(rb *RunningBuild) {
	// Cancel the context
	rb.Cancel()

	// Kill the process if it's still running
	if rb.Process != nil && rb.Process.Process != nil {
		if err := rb.Process.Process.Kill(); err != nil {
			fmt.Printf("[watcher] Failed to kill process: %v\n", err)
		}
	}

	// Mark as aborted
	if err := rb.Tracker.Abort(); err != nil {
		fmt.Printf("[watcher] Failed to mark build as aborted: %v\n", err)
	}

	fmt.Printf("[watcher] Aborted build: %s\n", rb.Rule.Name)
}

// stopAllBuilds aborts all running builds
func (w *Watcher) stopAllBuilds() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, rb := range w.runningBuilds {
		w.abortBuild(rb)
	}
}

// shouldIgnoreDirectory checks if a directory should be ignored based on rule patterns
func (w *Watcher) shouldIgnoreDirectory(dir string, rule *config.BuildRule) bool {
	relativePath, err := filepath.Rel(".", dir)
	if err != nil {
		relativePath = dir
	}

	for _, pattern := range rule.Ignore {
		if w.matchesPattern(relativePath, pattern) || w.matchesPattern(relativePath+"/", pattern) {
			return true
		}
	}
	return false
}

// shouldIgnoreFile checks if a file should be ignored based on any rule's ignore patterns
func (w *Watcher) shouldIgnoreFile(filename string) bool {
	relativePath, err := filepath.Rel(".", filename)
	if err != nil {
		relativePath = filename
	}

	// Check against all rules' ignore patterns
	for _, rule := range w.config.BuildRules {
		for _, pattern := range rule.Ignore {
			if w.matchesPattern(relativePath, pattern) {
				return true
			}
		}
	}
	return false
}

// SetBuildSuccessCallback sets the callback function to be called when a build succeeds
func (w *Watcher) SetBuildSuccessCallback(callback func()) {
	w.buildSuccessCallback = callback
}
