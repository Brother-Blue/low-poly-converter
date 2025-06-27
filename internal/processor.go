package internal

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"

	"golang.org/x/image/draw"
)

/*
LoadImage loads an image from the specified path and returns the image along with its format.
The format is determined by the file extension.
Supported formats are JPEG (.jpg, .jpeg) and PNG (.png).
*/
func LoadImage(path string, ext string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	switch ext {
	case ".jpg", ".jpeg":
		img, err := jpeg.Decode(file)
		return img, "jpeg", err
	case ".png":
		img, err := png.Decode(file)
		return img, "png", err
	default:
		return nil, "", fmt.Errorf("unsupported image format: %s", ext)
	}
}

/*
LoadGIF loads a GIF image from the specified path.
It returns a *gif.GIF object which contains all frames of the GIF.
*/
func LoadGIF(path string) (*gif.GIF, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return gif.DecodeAll(file)
}

/*
SaveImage saves the given image to the specified path in the specified format.
Supported formats are JPEG (.jpg, .jpeg) and PNG (.png).
*/
func SaveImage(img image.Image, path string, format string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch format {
	case "jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
	case "png":
		return png.Encode(file, img)
	default:
		return fmt.Errorf("unsupported image format for saving: %s", format)
	}
}

/*
ResizeImage resizes the given image to the specified width and height using Catmull-Rom interpolation.
It returns a new image with the resized dimensions.
*/
func ResizeImage(img image.Image, width int, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}
