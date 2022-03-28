# AppImage Go
AppImage Manipulation From Go. (Fork of [GoAppImage](https://github.com/probonopd/go-appimage/tree/master/src/goappimage))

---
## API

#### `NewAppImage(filePath string) error`
Takes A Path To A AppImage File & Returns A Struct & Error, Here The Struct Contains Properties & Functions To Modify And Get Information About AppImage.

```go
ai, err := appimagego.NewAppImage("/path/to/my/application.appimage")
if err != nil {
  panic(err)
}
```

---

## Functions

#### `(ai AppImage) Type() int`
Returns A Integer Between Which Represents The AppImage Type, Where `-1` means invalid AppImage, [See What Other Integers Represent](https://github.com/AppImage/AppImageSpec/blob/master/draft.md#image-format)

```go
appImageType := ai.Type()
if appImageType == -1 {
  panic("invalid appimage type: " + appImageType)
}
```

#### `(ai AppImage) ExtractFile(filePath string, destination string, resolveSymlinks bool) (error)`
Extract a file from the AppImage to the destination (here destination is path to a file & not a directory).

If `resolveSymlinks` is true & if the filepath specified is a symlink, then the actual file is extracted in it's place.

`resolveSymlinks` will have no effect on absolute symlinks (symlinks that start at root).

```go
err = ai.ExtractFile("application.desktop", "extractedEntry.desktop", true)
if err != nil {
  panic(err)
}
```

#### `(ai AppImage) ExtractFileReader(filepath string) (io.ReadCloser, error)`
Tries to get an `io.ReadCloser` for the file at filepath & Returns an error if the path is pointing to a folder.

If the path is pointing to a symlink, it will try to return the file being pointed to, but only if it's within the AppImage.

```go
fileReader, err := ai.ExtractFileReader("some/file/in/appimage")
```

#### `(ai AppImage) Thumbnail() (io.ReadCloser, error)`
Tries to get an `io.ReadCloser` of the AppImage Thumbnail

```go
thumbnailReader, err := ai.Thumbnail()
```

#### `(ai AppImage) Icon() (io.ReadCloser, string, error)`
Tries to get an `io.ReadCloser` of the AppImage Icon, returns the `io.ReadCloser` if no errors & icon path in the appimage

```go
iconReader, iconName, err := ai.Icon()
```

#### `(ai AppImage) GetUpdateInformation() (string, error)`
Gets the update information string from the AppImage

```go
updateInformation, err := ai.GetUpdateInformation()
```

#### `(ai AppImage) ModTime() time.Time`
Returns last time the AppImage was modified, it will try to get the information from squashfs if fails then returns the last time the AppImage file was modified.

```go
lastTimeModified := ai.ModTime()
fmt.Println("Last Time Modified:", lastTimeModified)
```

---

## Properties

#### `(ai AppImage) Desktop`
The Desktop entry of the appimage as [ini.File](https://pkg.go.dev/gopkg.in/ini.v1#File), which can be used to get properties out of the AppImage's Desktop Entry

#### `(ai AppImage) Name`
Name of the AppImage specified in the Desktop Entry

#### `(ai AppImage) Description`
Description of the AppImage specified in the Desktop Entry (The Comment Property)

#### `(ai AppImage) Version`
Version of the AppImage specified in the Desktop Entry

#### `(ai AppImage) Categories`
String array containing the AppImage's Categories specified in the Desktop Entry

#### `(ai AppImage) MimeType`
String array containing the AppImage's Mime Types

#### `(ai AppImage) Path`
Path to the AppImage

#### `(ai AppImage) UpdateInformation`
String which contains information related to appimage's update, contained in the AppImage itself
