package driver

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"strings"
)

var (
	NoSuchPoolErr = errors.New("no such pool")
)

type ZFSInterface interface {
	ImportZPool(poolName string) error
	CreateZPool(poolName string, devices []string) error
	CreateZFS(pathName string) error
	ExportZPool(poolName string) error
	Mount(source string, target string, opts ...string) error
	Unmount(target string) error
	GetStatistics(volumePath string) (VolumeStatistics, error)
	Resize(poolName string, deviceName string) error
}

type ZFSUtil struct {
}

func NewZFSUtil() ZFSInterface {
	return &ZFSUtil{}
}

type VolumeStatistics struct {
	availableBytes, totalBytes, usedBytes    int64
	availableInodes, totalInodes, usedInodes int64
}

func (z *ZFSUtil) ImportZPool(poolName string) error {

	_, err := os.Stat("/etc/hostid")

	if err != nil && os.IsNotExist(err) {

		err := z.execCommand("genhostid", []string{})

		if err != nil {
			return err
		}
	}

	cmd := fmt.Sprintf("zpool import -o multihost=on -o cachefile=none %s", poolName)
	err = z.execCommand("bash", []string{"-c", cmd})

	if err != nil {
		if strings.Contains(err.Error(), "no such pool") {
			return NoSuchPoolErr
		} else {
			return err
		}
	}
	return nil
}

func (z *ZFSUtil) CreateZPool(poolName string, devices []string) error {
	cmd := fmt.Sprintf("zpool create -o multihost=on -o cachefile=none %s %s",
		poolName, strings.Join(devices, " "))
	err := z.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return err
	}
	cmd = fmt.Sprintf("zfs set compression=lz4 %s", poolName)
	err = z.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return err
	}
	cmd = fmt.Sprintf("zpool set autoexpand=on %s", poolName)
	err = z.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return err
	}

	return nil
}

func (z *ZFSUtil) CreateZFS(pathName string) error {
	cmd := fmt.Sprintf("zfs create -p %s", pathName)
	err := z.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return err
	}
	return nil
}

func (z *ZFSUtil) ExportZPool(poolName string) error {
	cmd := fmt.Sprintf("zpool export %s", poolName)
	err := z.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		if strings.Contains(err.Error(), "no such pool") {
			return NoSuchPoolErr
		} else {
			return err
		}
	}
	return nil
}

func (z *ZFSUtil) Mount(source string, target string, opts ...string) error {
	mountCmd := "mount"
	var mountArgs []string

	if source == "" {
		return errors.New("source is not specified for mounting the volume")
	}

	if target == "" {
		return errors.New("target is not specified for mounting the volume")
	}

	if len(opts) > 0 {
		mountArgs = append(mountArgs, "-o", strings.Join(opts, ","))
	}

	mountArgs = append(mountArgs, source)
	mountArgs = append(mountArgs, target)

	// create target, os.Mkdirall is noop if it exists
	err := os.MkdirAll(target, 0750)
	if err != nil {
		return err
	}

	err = z.execCommand(mountCmd, mountArgs)

	if err != nil {
		return err
	}

	return nil
}

func (z *ZFSUtil) Unmount(target string) error {
	umountCmd := "umount"
	if target == "" {
		return errors.New("target is not specified for unmounting the volume")
	}

	umountArgs := []string{target}

	err := z.execCommand(umountCmd, umountArgs)

	if err != nil {
		return err
	}

	return nil
}

func (z *ZFSUtil) execCommand(command string, args []string) error {

	cmd := exec.Command(command, args...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New(fmt.Sprintf("cmd: %s, args: %v, output: %s, err: %s", command, args,
			out, err.Error()))
	}

	return nil
}

func (z *ZFSUtil) GetStatistics(volumePath string) (VolumeStatistics, error) {
	var statfs unix.Statfs_t
	// See http://man7.org/linux/man-pages/man2/statfs.2.html for details.
	err := unix.Statfs(volumePath, &statfs)
	if err != nil {
		return VolumeStatistics{}, err
	}

	volStats := VolumeStatistics{
		availableBytes: int64(statfs.Bavail) * int64(statfs.Bsize),
		totalBytes:     int64(statfs.Blocks) * int64(statfs.Bsize),
		usedBytes:      (int64(statfs.Blocks) - int64(statfs.Bfree)) * int64(statfs.Bsize),

		availableInodes: int64(statfs.Ffree),
		totalInodes:     int64(statfs.Files),
		usedInodes:      int64(statfs.Files) - int64(statfs.Ffree),
	}

	return volStats, nil
}

func (z *ZFSUtil) Resize(poolName string, deviceName string) error {
	cmd := fmt.Sprintf("zpool online -e %s %s", poolName, deviceName)
	err := z.execCommand("bash", []string{"-c", cmd})
	if err != nil {
		return err
	}
	return nil
}
