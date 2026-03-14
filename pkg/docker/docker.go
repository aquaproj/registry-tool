package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/aquaproj/registry-tool/pkg/osexec"
)

// Manager manages Docker container operations.
type Manager struct {
	config Config
}

// NewManager creates a new Manager with the given configuration.
func NewManager(config Config) *Manager {
	return &Manager{config: config}
}

// Config returns the container configuration.
func (dm *Manager) Config() Config {
	return dm.config
}

// EnsureContainer ensures the container is running.
// If recreate is true, it will stop and remove the existing container first.
func (dm *Manager) EnsureContainer(ctx context.Context, logger *slog.Logger, recreate bool) error {
	if recreate {
		if err := dm.RemoveContainer(ctx, logger); err != nil {
			return err
		}
	}

	// Build image if needed
	if err := dm.ensureImage(ctx, logger); err != nil {
		return err
	}

	exists, err := dm.ContainerExists(ctx, logger)
	if err != nil {
		return err
	}

	if !exists {
		logger.Info("creating a container", "container_name", dm.config.Name)
		return dm.runContainer(ctx, logger)
	}

	running, err := dm.ContainerRunning(ctx, logger)
	if err != nil {
		return err
	}

	if running {
		return dm.handleRunningContainer(ctx, logger)
	}

	return dm.handleStoppedContainer(ctx, logger)
}

// ContainerExists checks if the container exists.
func (dm *Manager) ContainerExists(ctx context.Context, logger *slog.Logger) (bool, error) {
	logger.Info("checking if the container exists", "container_name", dm.config.Name)
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", //nolint:gosec
		"--filter", "name="+dm.config.Name,
		"--format", "{{.Names}}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("docker ps: %w", err)
	}

	for line := range strings.SplitSeq(stdout.String(), "\n") {
		if strings.TrimSpace(line) == dm.config.Name {
			return true, nil
		}
	}
	return false, nil
}

// ContainerRunning checks if the container is running.
func (dm *Manager) ContainerRunning(ctx context.Context, logger *slog.Logger) (bool, error) {
	logger.Info("checking if the container is running", "container_name", dm.config.Name)
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", //nolint:gosec
		"--filter", "name="+dm.config.Name,
		"--filter", "status=running",
		"--format", "{{.Names}}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("docker ps: %w", err)
	}

	for line := range strings.SplitSeq(stdout.String(), "\n") {
		if strings.TrimSpace(line) == dm.config.Name {
			return true, nil
		}
	}
	return false, nil
}

// RemoveContainer stops and removes the container.
func (dm *Manager) RemoveContainer(ctx context.Context, logger *slog.Logger) error {
	exists, err := dm.ContainerExists(ctx, logger)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	cmd := exec.CommandContext(ctx, "docker", "stop", "-t", "1", dm.config.Name) //nolint:gosec
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	_ = cmd.Run() // Ignore error if container is not running

	cmd = exec.CommandContext(ctx, "docker", "rm", dm.config.Name) //nolint:gosec
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker rm: %w", err)
	}
	return nil
}

// Exec executes a command in the container.
func (dm *Manager) Exec(ctx context.Context, logger *slog.Logger, env map[string]string, command ...string) error {
	args := []string{"exec"}
	if dm.config.WorkingDir != "" {
		args = append(args, "-w", dm.config.WorkingDir)
	}
	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, dm.config.Name)
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	logger.Info("+ " + RedactSecrets(cmd.String(), env))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker exec: %w", err)
	}
	return nil
}

// ExecBash executes a bash command in the container.
func (dm *Manager) ExecBash(ctx context.Context, logger *slog.Logger, bashCmd string) error {
	return dm.Exec(ctx, logger, nil, "bash", "-c", bashCmd)
}

// CopyTo copies a file from the host to the container.
func (dm *Manager) CopyTo(ctx context.Context, logger *slog.Logger, src, dst string) error {
	containerPath := dm.config.Name + ":" + dst
	cmd := exec.CommandContext(ctx, "docker", "cp", src, containerPath)
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker cp to container: %w", err)
	}
	return nil
}

// CopyFrom copies a file from the container to the host.
func (dm *Manager) CopyFrom(ctx context.Context, logger *slog.Logger, src, dst string) error {
	containerPath := dm.config.Name + ":" + src
	cmd := exec.CommandContext(ctx, "docker", "cp", containerPath, dst)
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker cp from container: %w", err)
	}
	return nil
}

func (dm *Manager) handleRunningContainer(ctx context.Context, logger *slog.Logger) error {
	upToDate, err := dm.checkImageUpToDate(ctx, logger)
	if err != nil {
		return err
	}
	if upToDate {
		logger.Info("Dockerfile isn't updated")
		return nil
	}
	logger.Info("Dockerfile is updated, so the container is being recreated", "container_name", dm.config.Name)
	if err := dm.RemoveContainer(ctx, logger); err != nil {
		return err
	}
	return dm.runContainer(ctx, logger)
}

func (dm *Manager) handleStoppedContainer(ctx context.Context, logger *slog.Logger) error {
	upToDate, err := dm.checkImageUpToDate(ctx, logger)
	if err != nil {
		return err
	}
	if upToDate {
		logger.Info("Dockerfile isn't updated")
		logger.Info("starting the container", "container_name", dm.config.Name)
		return dm.startContainer(ctx, logger)
	}

	logger.Info("Dockerfile is updated, so the container is being recreated", "container_name", dm.config.Name)
	if err := dm.RemoveContainer(ctx, logger); err != nil {
		return err
	}
	return dm.runContainer(ctx, logger)
}

func (dm *Manager) ensureImage(ctx context.Context, logger *slog.Logger) error {
	notChanged, err := dm.dockerfileNotChanged()
	if err != nil {
		return err
	}
	if notChanged && dm.imageExists(ctx, logger) {
		return nil
	}

	logger.Info("building the docker image", "image", dm.config.Image)
	return dm.buildImage(ctx, logger)
}

func (dm *Manager) imageExists(ctx context.Context, logger *slog.Logger) bool {
	cmd := exec.CommandContext(ctx, "docker", "inspect", dm.config.Image) //nolint:gosec
	cmd.Stdout = nil
	cmd.Stderr = nil
	osexec.SetCancel(logger, cmd)
	return cmd.Run() == nil
}

func (dm *Manager) dockerfileNotChanged() (bool, error) {
	// Check if .build/Dockerfile exists
	if _, err := os.Stat(".build/Dockerfile"); os.IsNotExist(err) {
		return false, nil
	}

	// Compare docker/Dockerfile with .build/Dockerfile
	b1, err := os.ReadFile("docker/Dockerfile")
	if err != nil {
		return false, fmt.Errorf("read docker/Dockerfile: %w", err)
	}
	b2, err := os.ReadFile(".build/Dockerfile")
	if err != nil {
		return false, fmt.Errorf("read .build/Dockerfile: %w", err)
	}
	return string(b1) == string(b2), nil
}

func (dm *Manager) buildImage(ctx context.Context, logger *slog.Logger) error {
	// Copy aqua-policy.yaml to docker directory
	if err := CopyFile("aqua-policy.yaml", "docker/aqua-policy.yaml"); err != nil {
		return fmt.Errorf("copy aqua-policy.yaml: %w", err)
	}

	// Build Docker image
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", dm.config.Image, "docker") //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build: %w", err)
	}

	// Save Dockerfile to .build for change detection
	if err := os.MkdirAll(".build", DirPermission); err != nil {
		return fmt.Errorf("create .build directory: %w", err)
	}
	if err := CopyFile("docker/Dockerfile", ".build/Dockerfile"); err != nil {
		return fmt.Errorf("copy Dockerfile to .build: %w", err)
	}

	return nil
}

func (dm *Manager) checkImageUpToDate(ctx context.Context, logger *slog.Logger) (bool, error) {
	containerImageID, err := dm.getContainerImageID(ctx, logger)
	if err != nil {
		return false, err
	}

	imageID, err := dm.getImageID(ctx, logger)
	if err != nil {
		return false, err
	}

	return containerImageID == imageID, nil
}

func (dm *Manager) getContainerImageID(ctx context.Context, logger *slog.Logger) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", dm.config.Name) //nolint:gosec
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker inspect container: %w", err)
	}

	var result []struct {
		Image string `json:"Image"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", fmt.Errorf("parse docker inspect output: %w", err)
	}
	if len(result) == 0 {
		return "", fmt.Errorf("container %s not found", dm.config.Name)
	}
	return result[0].Image, nil
}

func (dm *Manager) getImageID(ctx context.Context, logger *slog.Logger) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", dm.config.Image) //nolint:gosec
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker inspect image: %w", err)
	}

	var result []struct {
		ID string `json:"Id"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", fmt.Errorf("parse docker inspect output: %w", err)
	}
	if len(result) == 0 {
		return "", fmt.Errorf("image %s not found", dm.config.Image)
	}
	return result[0].ID, nil
}

func (dm *Manager) runContainer(ctx context.Context, logger *slog.Logger) error {
	args := []string{"run", "-d", "--name", dm.config.Name}

	// Add --privileged for Podman on Linux
	if runtime.GOOS == "linux" && IsPodman(ctx, logger) {
		args = append(args, "--privileged")
	}

	args = append(args, dm.config.Image, "tail", "-f", "/dev/null")

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker run: %w", err)
	}
	return nil
}

func (dm *Manager) startContainer(ctx context.Context, logger *slog.Logger) error {
	cmd := exec.CommandContext(ctx, "docker", "start", dm.config.Name) //nolint:gosec
	logger.Info("+ " + cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker start: %w", err)
	}
	return nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err //nolint:wrapcheck
	}
	return os.WriteFile(dst, data, FilePermission) //nolint:wrapcheck,gosec
}

// RedactSecrets replaces secret values in a string with <REDACTED>.
func RedactSecrets(s string, env map[string]string) string {
	for _, v := range env {
		if v != "" {
			s = strings.ReplaceAll(s, v, "<REDACTED>")
		}
	}
	return s
}

// IsPodman checks if Docker is actually Podman.
func IsPodman(ctx context.Context, logger *slog.Logger) bool {
	cmd := exec.CommandContext(ctx, "docker", "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil
	osexec.SetCancel(logger, cmd)
	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.Contains(stdout.String(), "Podman")
}
