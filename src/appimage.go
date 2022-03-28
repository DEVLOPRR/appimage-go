package appimagego

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"debug/elf"
	"strconv"
	"strings"
	"time"

	"appimagego/src/internal/helpers"
	"gopkg.in/ini.v1"
)

/*

TODO List:
* Check if there IS an update
* Download said update

*/

// AppImage handles AppImage files.
type AppImage struct {
	// Public
	Desktop           *ini.File // Desktop is the AppImage's main .desktop file parsed as an ini.File
	Name              string
	Description       string
	Version           string
	Categories        []string
	MimeType	      []string
	Path              string
	UpdateInformation string

	// Private
	reader archiveReader
	offset            int64
	imageType         int
  // Functions Available In This Struct
  //(ai AppImage) Type() (int)
  //(ai AppImage) ExtractFile(filePath string, destination string) (error)
  //(ai AppImage) ExtractFileReader(filepath string) (io.ReadCloser, error)
  //(ai AppImage) Thumbnail() (io.ReadCloser, error)
  //(ai AppImage) Icon() (io.ReadCloser, string, error)
  //(ai AppImage) GetUpdateInformation() (string, error)
  //(ai AppImage) ShallBeIntegrated() (bool)
  //(ai AppImage) ModTime() time.Time
}

// NewAppImage creates an AppImage object from the location defined by path.
// Returns an error if the given path is not an appimage, or is a temporary file.
// In all instances, will still return the AppImage.
func NewAppImage(path string) (*AppImage, error) {
	ai := AppImage{Path: path, imageType: -1}

	// If we got a temp file, exit immediately
	// E.g., ignore typical Internet browser temporary files used during download
	if strings.HasSuffix(path, ".temp") ||
		strings.HasSuffix(path, "~") ||
		strings.HasSuffix(path, ".part") ||
		strings.HasSuffix(path, ".partial") ||
		strings.HasSuffix(path, ".zs-old") ||
		strings.HasSuffix(path, ".crdownload") {
		return &ai, errors.New("given path is a temporary file")
	}

	ai.imageType = ai.determineImageType()

	if ai.imageType < 0 {
		return &ai, errors.New("given path is not an AppImage")
	}

	if ai.imageType > 1 {
		ai.offset = helpers.CalculateElfSize(ai.Path)
	}

	err := ai.populateReader(true, false)
	if err != nil {
		return &ai, err
	}

	//try to load up the desktop file for some information.
	desktopFil, err := ai.reader.FileReader("*.desktop")
	if err != nil {
		return nil, err
	}

	// cleaning the desktop file so it can be parsed properly
	var desktop []byte
	buf := bufio.NewReader(desktopFil)
	for err == nil {
		var line string
		line, err = buf.ReadString('\n')
		if strings.Contains(line, ";") {
			line = strings.ReplaceAll(line, ";", "；") //replacing it with a fullwidth semicolon (unicode FF1B)
		}
		desktop = append(desktop, line...)
	}

	ai.Desktop, err = ini.Load(desktop)
	if err == nil {
		desktopEntry := ai.Desktop.Section("Desktop Entry")

		ai.Name = desktopEntry.Key("Name").Value()
		ai.Version = desktopEntry.Key("X-AppImage-Version").Value()
		ai.MimeType = strings.Split(desktopEntry.Key("MimeType").Value(), "；")
		ai.Categories = strings.Split(desktopEntry.Key("Categories").Value(), "；")
		ai.Description = desktopEntry.Key("Comment").Value()
	}

	if ai.Name == "" {
		ai.Name = ai.calculateNiceName()
	}

	if ai.Version == "" {
		ai.Version = ai.Desktop.Section("Desktop Entry").Key("Version").Value()
		if ai.Version == "" {
			ai.Version = "1.0"
		}
	}

	return &ai, nil
}

//Type is the type of the AppImage. Should be either 1 or 2.
func (ai AppImage) Type() int {
	return ai.imageType
}

func (ai AppImage) ShallBeIntegrated() bool {
	var integrationRequested string = "true"
	var NoDisplay string = "false"

	if ai.Desktop.Section("Desktop Entry").HasKey("X-AppImage-Integrate") {
		integrationRequested = ai.Desktop.Section("Desktop Entry").Key("X-AppImage-Integrate").Value()
	} else if ai.Desktop.Section("Desktop Entry").HasKey("NoDisplay") {
		NoDisplay = ai.Desktop.Section("Desktop Entry").Key("NoDisplay").Value()
	}

	if integrationRequested == "false" || NoDisplay == "true" {
		return false
	} else {
		return true
	}
}

//ExtractFile extracts a file from from filepath (which may contain * wildcards) in an AppImage to the destinationdirpath.
//
//If resolveSymlinks is true & if the filepath specified is a symlink, then the actual file is extracted in it's place.
//resolveSymlinks will have no effect on absolute symlinks (symlinks that start at root).
func (ai AppImage) ExtractFile(filepath string, destinationdirpath string, resolveSymlinks bool) error {
	return ai.reader.ExtractTo(filepath, destinationdirpath, resolveSymlinks)
}

//ExtractFileReader tries to get an io.ReadCloser for the file at filepath.
//Returns an error if the path is pointing to a folder. If the path is pointing to a symlink,
//it will try to return the file being pointed to, but only if it's within the AppImage.
func (ai AppImage) ExtractFileReader(filepath string) (io.ReadCloser, error) {
	return ai.reader.FileReader(filepath)
}

//Thumbnail tries to get the AppImage's thumbnail and returns it as a io.ReadCloser.
func (ai AppImage) Thumbnail() (io.ReadCloser, error) {
	return ai.reader.FileReader(".DirIcon")
}

//Icon tries to get a io.ReadCloser for the icon dictated in the AppImage's desktop file.
//Returns the ReadCloser and the file's name (which could be useful for decoding).
func (ai AppImage) Icon() (io.ReadCloser, string, error) {
	if ai.Desktop == nil {
		return nil, "", errors.New("desktop file wasn't parsed")
	}
	icon := ai.Desktop.Section("Desktop Entry").Key("Icon").Value()
	if icon == "" {
		return nil, "", errors.New("desktop file doesn't specify an icon")
	}
	if strings.HasSuffix(icon, ".png") || strings.HasSuffix(icon, ".svg") {
		rdr, err := ai.reader.FileReader(icon)
		if err == nil {
			return rdr, icon, nil
		}
	}
	rootFils := ai.reader.ListFiles("/")
	for _, fil := range rootFils {
		if strings.HasPrefix(fil, icon) {
			if fil == icon+".png" {
				rdr, err := ai.reader.FileReader(fil)
				if err != nil {
					continue
				}
				return rdr, fil, nil
			} else if fil == icon+".svg" {
				rdr, err := ai.reader.FileReader(fil)
				if err != nil {
					continue
				}
				return rdr, fil, nil
			}
		}
	}
	return nil, "", errors.New("Cannot find the AppImage's icon: " + icon)
}

// ReadUpdateInformation reads updateinformation from an AppImage
func (ai AppImage) GetUpdateInformation() (string, error) {
	elfFile, err := elf.Open(ai.Path)
	if err != nil {
		return "", errors.New("cannot open the appimage file: " + err.Error())
	}

	updInfo := elfFile.Section(".upd_info")
	if updInfo == nil {
		return "", errors.New("missing update section on target elf")
	}
	sectionData, err := updInfo.Data()

	if err != nil {
		return "", errors.New("unable to parse update section: " + err.Error())
	}

	str_end := bytes.Index(sectionData, []byte("\000"))
	update_info := string(sectionData[:str_end])
	return update_info, nil
}

//ModTime is the time the AppImage was edited/created. If the AppImage is type 2,
//it will try to get that information from the squashfs, if not, it returns the file's ModTime.
func (ai AppImage) ModTime() time.Time {
	if ai.imageType == 2 {
		if ai.reader != nil {
			return ai.reader.(*type2Reader).rdr.ModTime()
		}
		result, err := exec.Command("unsquashfs", "-q", "-fstime", "-o", strconv.FormatInt(ai.offset, 10), ai.Path).Output()
		resstr := strings.TrimSpace(string(bytes.TrimSpace(result)))
		if err != nil {
			goto fallback
		}
		if n, err := strconv.Atoi(resstr); err == nil {
			return time.Unix(int64(n), 0)
		}
	}
fallback:
	fil, err := os.Open(ai.Path)
	if err != nil {
		return time.Unix(0, 0)
	}
	stat, _ := fil.Stat()
	return stat.ModTime()
}
