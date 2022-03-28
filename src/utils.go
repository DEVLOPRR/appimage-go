package appimagego

import (
	"os"
	"bytes"
	"os/exec"
	"strings"
	"path/filepath"
	"appimagego/src/internal/helpers"
)

// Run a Command
func runCommand(cmd *exec.Cmd) (bytes.Buffer, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out, err
}

// Calculate a name from the a filename
func (ai AppImage) calculateNiceName() string {
	niceName := filepath.Base(ai.Path)
	niceName = strings.Replace(niceName, ".AppImage", "", -1)
	niceName = strings.Replace(niceName, ".appimage", "", -1)
	niceName = strings.Replace(niceName, "-x86_64", "", -1)
	niceName = strings.Replace(niceName, "-i386", "", -1)
	niceName = strings.Replace(niceName, "-i686", "", -1)
	niceName = strings.Replace(niceName, "-aarch64", "", -1)
	niceName = strings.Replace(niceName, "-armhf", "", -1)
	niceName = strings.Replace(niceName, "-", " ", -1)
	niceName = strings.Replace(niceName, "_", " ", -1)
	return niceName
}

// Check whether we have an AppImage at all.
// Return image type, or -1 if it is not an AppImage
func (ai AppImage) determineImageType() int {
	// log.Println("appimage: ", ai.path)
	f, err := os.Open(ai.Path)
	// printError("appimage", err)
	if err != nil {
		return -1 // If we were not able to open the file, then we report that it is not an AppImage
	}
	info, err := os.Stat(ai.Path)
	if err != nil {
		return -1
	}
	// Directories cannot be AppImages, so return fast
	if info.IsDir() {
		return -1
	}
	// Very small files cannot be AppImages, so return fast
	if info.Size() < 100*1024 {
		return -1
	}
	if helpers.CheckMagicAtOffset(f, "414902", 8) {
		return 2
	}
	if helpers.CheckMagicAtOffset(f, "414901", 8) {
		return 1
	}
	// ISO9660 files that are also ELF files
	if helpers.CheckMagicAtOffset(f, "7f454c", 0) && helpers.CheckMagicAtOffset(f, "4344303031", 32769) {
		return 1
	}
	return -1
}
