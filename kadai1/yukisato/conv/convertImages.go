package conv

import (
	"errors"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	contentTypeJPEG  = "image/jpeg"
	contentTypePNG   = "image/png"
	contentTypeOther = "application/octet-stream"
	extensionJPEG    = ".jpg"
	extensionPNG     = ".png"
)

// Indecates file destination to convert.
type fileDest struct {
	from *os.File
	to   *os.File
}

// ConvertImages converts an image file with an extension to another specified by "extFrom" and "extTo" in "destDir" directory.
func ConvertImages(destDir, extFrom, extTo string) error {
	if extFrom == extTo {
		return errors.New("specified extensions must be distinct")
	}

	return filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, extFrom) {
			err = convert(path, extFrom, extTo)
		}

		return err
	})
}

// Convert all the image files with a specified extension under target directory to another file with anotehr extension.
func convert(filepathFrom, extFrom, extTo string) error {
	from, err := os.Open(filepathFrom)
	if err != nil {
		return err
	}
	defer from.Close()

	filepathTo := strings.TrimSuffix(filepathFrom, extFrom) + extTo
	to, err := os.Create(filepathTo)
	if err != nil {
		return err
	}
	defer to.Close()

	switch {
	case extFrom == extensionJPEG && extTo == extensionPNG:
		err = jpegToPNG(fileDest{from, to})
	case extFrom == extensionPNG && extTo == extensionJPEG:
		err = pngToJPEG(fileDest{from, to})
	default:
		err = errors.New("unsupported extension combination to covert from: " + extFrom + " to: " + extTo)
	}

	if err != nil {
		defer os.Remove(filepathTo)
	}

	return err
}

// Convert an image file from jpeg to png.
func jpegToPNG(dest fileDest) error {
	if !isJPEG(dest.from) {
		return errors.New("content type of the original file is not " + contentTypeJPEG)
	}

	jpegImg, err := jpeg.Decode(dest.from)

	if err != nil {
		return err
	}

	png.Encode(dest.to, jpegImg)
	return nil
}

// Convert an image file from png to jpeg.
func pngToJPEG(dest fileDest) error {
	if !isPNG(dest.from) {
		return errors.New("content type of the original file is not " + contentTypePNG)
	}

	pngImg, err := png.Decode(dest.from)

	if err != nil {
		return err
	}

	return jpeg.Encode(dest.to, pngImg, nil)
}

// Determine if the content type of a given file is image/jpeg
func isJPEG(file *os.File) bool {
	contentType, _ := getFileContentType(file)
	return contentType == contentTypeJPEG
}

// Determine if the content type of a given file is image/png
func isPNG(file *os.File) bool {
	contentType, _ := getFileContentType(file)
	return contentType == contentTypePNG
}

// Detect content type from the first 512 bytes of a given file.
func getFileContentType(file *os.File) (string, error) {
	// Using the first 512 bytes to detect the content type.
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	// Reset the file pointer
	file.Seek(0, io.SeekStart)

	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}
