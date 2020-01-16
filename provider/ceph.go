package provider

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/klog"
	"os/exec"
	"path"
	"strings"
	"time"
)

type CephProvider struct {
	timeout  int
	poolName string
}

func NewCephProvider(timeout int, poolName string) *CephProvider {
	return &CephProvider{timeout: timeout, poolName: poolName}
}

const (
	imageWatcherStr = "watcher="
)

var (
	NoSuchFileErr = errors.New("no such file or directory")
	NotMapDevice  = errors.New("not a mapped image or snapshot")
)

func (c *CephProvider) GetVolumeType() VolumeType {
	return CephType
}

func (c *CephProvider) GetDetail() string {
	return "ceph provider"
}

func (c *CephProvider) CreateVolume(ctx context.Context, volId string, size int64) (string, error) {

	cmd := fmt.Sprintf("rbd create %s --size %d --image-feature layering", c.GetImageName(volId), size)

	output, err := c.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return "", fmt.Errorf("failed to create rbd err: %v, command output: %s,"+
			" command: rbd %v", err, string(output), cmd)
	}
	return volId, nil
}

func (c *CephProvider) DeleteVolume(ctx context.Context, volId string) error {
	found, err := c.rbdStatus(c.GetImageName(volId))
	if err != nil {

		if err == NoSuchFileErr {
			return nil
		}

		return err
	}
	if found {
		return fmt.Errorf("rbd %s is still being used", c.GetImageName(volId))
	}
	cmd := fmt.Sprintf("rbd rm %s", c.GetImageName(volId))
	output, err := c.execCommand("bash", []string{"-c", cmd})
	if err == nil {
		return nil
	}
	klog.Errorf("failed to delete rbd image: %v, command output: %s", err, string(output))
	return err
}

func (c *CephProvider) Attach(ctx context.Context, volId string) (string, error) {

	found, err := c.rbdStatus(c.GetImageName(volId))
	if err != nil && err != NoSuchFileErr {
		return volId, err
	}

	if found {

		deviceName, err := c.GetMapDevice(volId)
		if err != nil {
			return "", err
		}
		return deviceName, nil
	}

	cmd := fmt.Sprintf("rbd map %s", c.GetImageName(volId))
	output, err := c.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return "", fmt.Errorf("failed to map rbd err: %v, command output: %s,"+
			" command: %v", err, string(output), cmd)
	}
	return strings.TrimRight(string(output), "\n"), nil
}

func (c *CephProvider) Detach(ctx context.Context, volId string) error {

	_, err := c.rbdStatus(c.GetImageName(volId))
	if err != nil {

		if err == NoSuchFileErr || err == NotMapDevice {
			return nil
		}

		return err
	}

	cmd := fmt.Sprintf("rbd unmap %s", c.GetImageName(volId))
	output, err := c.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return fmt.Errorf("failed to unmap rbd err: %v, command output: %s,"+
			" command: %v", err, string(output), cmd)
	}
	return nil
}

func (c *CephProvider) execCommand(command string, args []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Second)
	defer cancel()

	// Create the command with our context
	cmd := exec.CommandContext(ctx, command, args...)

	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("rbd: Command timed out")
	}

	// If there's no context error, we know the command completed (or errored).
	if err != nil {
		return out, fmt.Errorf("rbd: Command exited with non-zero code: %v", err)
	}

	return out, err
}

func (c *CephProvider) HasVolume(volId string) (bool, error) {
	found, err := c.rbdStatus(c.GetImageName(volId))
	if err != nil {
		return false, err
	}
	return found, nil
}

func (c *CephProvider) rbdStatus(image string) (bool, error) {
	var err error

	args := []string{"status", image}
	cmd := fmt.Sprintf("rbd status %s", image)
	out, err := c.execCommand("bash", []string{"-c", cmd})

	// If command never succeed, returns its last error.
	if err != nil {

		if strings.Contains(string(out), "No such file or directory") {
			return false, NoSuchFileErr
		} else if strings.Contains(string(out), "not a mapped image or snapshot") {
			return false, NotMapDevice
		} else {
			return false, fmt.Errorf("rbd status err: %v, command output: %s, command:rbd %v", err, out, args)
		}
	}

	if strings.Contains(string(out), imageWatcherStr) {
		return true, nil
	}
	return false, nil
}

func (c *CephProvider) GetImageName(volId string) string {
	return path.Join(c.poolName, path.Base(volId))
}

func (c *CephProvider) GetMapDevice(volId string) (string, error) {
	cmd := fmt.Sprintf("rbd showmapped |grep %s|awk '{print $5'}", volId)
	out, err := c.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return "", errors.New(fmt.Sprintf("out: %s, err: %s", out, err.Error()))
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (c *CephProvider) Resize(volId string, size int64) error {
	cmd := fmt.Sprintf("rbd resize --size %d %s", size, c.GetImageName(volId))
	out, err := c.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return errors.New(fmt.Sprintf("out: %s, err: %s", out, err.Error()))
	}
	return nil
}
