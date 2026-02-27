package scaffold

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DockerManager manages Docker container operations.
type DockerManager struct {
	config ContainerConfig
}

// NewDockerManager creates a new DockerManager with the given configuration.
func NewDockerManager(config ContainerConfig) *DockerManager {
	return &DockerManager{config: config}
}

// EnsureContainer ensures the container is running.
// If recreate is true, it will stop and remove the existing container first.
func (dm *DockerManager) EnsureContainer(ctx context.Context, recreate bool) error {
	if recreate {
		if err := dm.RemoveContainer(ctx); err != nil {
			return err
		}
	}

	// Build image if needed
	if err := dm.ensureImage(ctx); err != nil {
		return err
	}

	exists, err := dm.ContainerExists(ctx)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Fprintf(os.Stderr, "[INFO] Creating a container %s\n", dm.config.Name)
		return dm.runContainer(ctx)
	}

	running, err := dm.ContainerRunning(ctx)
	if err != nil {
		return err
	}

	if running {
		return dm.handleRunningContainer(ctx)
	}

	return dm.handleStoppedContainer(ctx)
}

// ContainerExists checks if the container exists.
func (dm *DockerManager) ContainerExists(ctx context.Context) (bool, error) {
	fmt.Fprintf(os.Stderr, "[INFO] Checking if the container %s exists\n", dm.config.Name)
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", //nolint:gosec
		"--filter", "name="+dm.config.Name,
		"--format", "{{.Names}}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
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
func (dm *DockerManager) ContainerRunning(ctx context.Context) (bool, error) {
	fmt.Fprintf(os.Stderr, "[INFO] Checking if the container %s is running\n", dm.config.Name)
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", //nolint:gosec
		"--filter", "name="+dm.config.Name,
		"--filter", "status=running",
		"--format", "{{.Names}}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
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
func (dm *DockerManager) RemoveContainer(ctx context.Context) error {
	exists, err := dm.ContainerExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	fmt.Fprintf(os.Stderr, "+ docker stop -t 1 %s\n", dm.config.Name)
	cmd := exec.CommandContext(ctx, "docker", "stop", "-t", "1", dm.config.Name) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run() // Ignore error if container is not running

	fmt.Fprintf(os.Stderr, "+ docker rm %s\n", dm.config.Name)
	cmd = exec.CommandContext(ctx, "docker", "rm", dm.config.Name) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker rm: %w", err)
	}
	return nil
}

// Exec executes a command in the container.
func (dm *DockerManager) Exec(ctx context.Context, env map[string]string, command ...string) error {
	args := []string{"exec"}
	if dm.config.WorkingDir != "" {
		args = append(args, "-w", dm.config.WorkingDir)
	}
	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, dm.config.Name)
	args = append(args, command...)

	cmdStr := "docker " + strings.Join(args, " ")
	fmt.Fprintln(os.Stderr, "+ "+cmdStr)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker exec: %w", err)
	}
	return nil
}

// ExecBash executes a bash command in the container.
func (dm *DockerManager) ExecBash(ctx context.Context, bashCmd string) error {
	return dm.Exec(ctx, nil, "bash", "-c", bashCmd)
}

// CopyTo copies a file from the host to the container.
func (dm *DockerManager) CopyTo(ctx context.Context, src, dst string) error {
	containerPath := dm.config.Name + ":" + dst
	fmt.Fprintf(os.Stderr, "+ docker cp %s %s\n", src, containerPath)
	cmd := exec.CommandContext(ctx, "docker", "cp", src, containerPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker cp to container: %w", err)
	}
	return nil
}

// CopyFrom copies a file from the container to the host.
func (dm *DockerManager) CopyFrom(ctx context.Context, src, dst string) error {
	containerPath := dm.config.Name + ":" + src
	fmt.Fprintf(os.Stderr, "+ docker cp %s %s\n", containerPath, dst)
	cmd := exec.CommandContext(ctx, "docker", "cp", containerPath, dst)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker cp from container: %w", err)
	}
	return nil
}

func (dm *DockerManager) handleRunningContainer(ctx context.Context) error {
	upToDate, err := dm.checkImageUpToDate(ctx)
	if err != nil {
		return err
	}
	if upToDate {
		fmt.Fprintf(os.Stderr, "[INFO] Dockerfile isn't updated\n")
		return nil
	}
	fmt.Fprintf(os.Stderr, "[INFO] Dockerfile is updated, so the container %s is being recreated\n", dm.config.Name)
	if err := dm.RemoveContainer(ctx); err != nil {
		return err
	}
	return dm.runContainer(ctx)
}

func (dm *DockerManager) handleStoppedContainer(ctx context.Context) error {
	upToDate, err := dm.checkImageUpToDate(ctx)
	if err != nil {
		return err
	}
	if upToDate {
		fmt.Fprintf(os.Stderr, "[INFO] Dockerfile isn't updated\n")
		fmt.Fprintf(os.Stderr, "[INFO] Starting the container %s\n", dm.config.Name)
		return dm.startContainer(ctx)
	}

	fmt.Fprintf(os.Stderr, "[INFO] Dockerfile is updated, so the container %s is being recreated\n", dm.config.Name)
	if err := dm.RemoveContainer(ctx); err != nil {
		return err
	}
	return dm.runContainer(ctx)
}

func (dm *DockerManager) ensureImage(ctx context.Context) error {
	notChanged, err := dm.dockerfileNotChanged()
	if err == nil && notChanged && dm.imageExists(ctx) {
		return nil
	}

	fmt.Fprintf(os.Stderr, "[INFO] Building the docker image %s\n", dm.config.Image)
	return dm.buildImage(ctx)
}

func (dm *DockerManager) imageExists(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "inspect", dm.config.Image) //nolint:gosec
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func (dm *DockerManager) dockerfileNotChanged() (bool, error) {
	// Check if .build/Dockerfile exists
	if _, err := os.Stat(".build/Dockerfile"); os.IsNotExist(err) {
		return false, nil
	}

	// Compare docker/Dockerfile with .build/Dockerfile
	b1, err := os.ReadFile("docker/Dockerfile")
	if err != nil {
		return false, err
	}
	b2, err := os.ReadFile(".build/Dockerfile")
	if err != nil {
		return false, err
	}
	return string(b1) == string(b2), nil
}

func (dm *DockerManager) buildImage(ctx context.Context) error {
	// Copy aqua-policy.yaml to docker directory
	if err := copyFile("aqua-policy.yaml", "docker/aqua-policy.yaml"); err != nil {
		// Ignore if file doesn't exist
		if !os.IsNotExist(err) {
			return fmt.Errorf("copy aqua-policy.yaml: %w", err)
		}
	}

	// Build Docker image
	fmt.Fprintln(os.Stderr, "+ docker build -t "+dm.config.Image+" docker")
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", dm.config.Image, "docker") //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build: %w", err)
	}

	// Save Dockerfile to .build for change detection
	if err := os.MkdirAll(".build", dirPermission); err != nil {
		return fmt.Errorf("create .build directory: %w", err)
	}
	if err := copyFile("docker/Dockerfile", ".build/Dockerfile"); err != nil {
		return fmt.Errorf("copy Dockerfile to .build: %w", err)
	}

	return nil
}

func (dm *DockerManager) checkImageUpToDate(ctx context.Context) (bool, error) {
	containerImageID, err := dm.getContainerImageID(ctx)
	if err != nil {
		return false, err
	}

	imageID, err := dm.getImageID(ctx)
	if err != nil {
		return false, err
	}

	return containerImageID == imageID, nil
}

func (dm *DockerManager) getContainerImageID(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", dm.config.Name) //nolint:gosec
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
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

func (dm *DockerManager) getImageID(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", dm.config.Image) //nolint:gosec
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
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

func (dm *DockerManager) runContainer(ctx context.Context) error {
	token := getGitHubToken()

	args := []string{"run"}

	// Add --privileged for Podman on Linux
	if runtime.GOOS == "linux" && isPodman(ctx) {
		args = append(args, "--privileged")
	}

	args = append(args, "-d", "--name", dm.config.Name)

	if token != "" {
		args = append(args, "-e", "GITHUB_TOKEN="+token)
	}

	args = append(args, dm.config.Image, "tail", "-f", "/dev/null")

	fmt.Fprintf(os.Stderr, "+ docker run -d --name %s %s tail -f /dev/null\n", dm.config.Name, dm.config.Image)
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker run: %w", err)
	}
	return nil
}

func (dm *DockerManager) startContainer(ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "+ docker start %s\n", dm.config.Name)
	cmd := exec.CommandContext(ctx, "docker", "start", dm.config.Name) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker start: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err //nolint:wrapcheck
	}
	return os.WriteFile(dst, data, filePermission) //nolint:wrapcheck
}

func isPodman(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.Contains(stdout.String(), "Podman")
}
