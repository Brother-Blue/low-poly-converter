package internal

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"sync"

	"github.com/schollz/progressbar/v3"
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

/*
ProcessGifParallel processes each frame of the GIF in parallel.
It applies the low-poly effect to each frame and resizes it if width and height are specified.
*/
func ProcessGifParallel(images *gif.GIF, width, height, intensity int, bar *progressbar.ProgressBar) *gif.GIF {
	var wg sync.WaitGroup
	frames := make([]*image.Paletted, len(images.Image))
	w, h := images.Config.Width, images.Config.Height
	if width > 0 && height > 0 {
		w, h = width, height
	}
	images.Config.Width = w
	images.Config.Height = h

	for idx, frame := range images.Image {
		wg.Add(1)
		go func(idx int, frame *image.Paletted) {
			defer wg.Done()
			rgba := image.NewRGBA(frame.Bounds())
			draw.Draw(rgba, frame.Bounds(), frame, image.Point{}, draw.Src)
			if width > 0 && height > 0 {
				rgba = ResizeImage(rgba, width, height).(*image.RGBA)
			}

			processed := ApplyLowPoly(rgba, intensity)
			bounds := image.Rect(0, 0, w, h)
			paletted := image.NewPaletted(bounds, frame.Palette)
			draw.Draw(paletted, bounds, processed, processed.Bounds().Min, draw.Over)
			frames[idx] = paletted

			if bar != nil {
				bar.Add(1)
			}
		}(idx, frame)
	}
	wg.Wait()
	images.Image = frames
	return images
}
