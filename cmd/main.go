package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/brother-blue/low-poly-converter/internal"
)

var (
	resizeFlag    = flag.String("resize", "", "Resize the image to the specified dimensions (e.g., 800x600)")
	intensityFlag = flag.Int("intensity", 100, "Set the intensity of the image processing (1-100)")
)

func parseRezie(dimensions string) (width int, height int, err error) {
	if dimensions == "" {
		return 0, 0, nil
	}
	re := regexp.MustCompile(`^(\d+)x(\d+)$`)
	matches := re.FindStringSubmatch(dimensions)
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("invalid resize format, expected WIDTHxHEIGHT (e.g., 800x600)")
	}
	width, _ = strconv.Atoi(matches[1])
	height, _ = strconv.Atoi(matches[2])
	return width, height, nil
}

func getOutputPath(inputPath string) string {
	extension := filepath.Ext(inputPath)
	name := strings.TrimSuffix(filepath.Base(inputPath), extension)
	dir := filepath.Dir(inputPath)
	return filepath.Join(dir, fmt.Sprintf("%s-low-poly%s", name, extension))
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: poly-convert [options] <input image>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)
	outputPath := getOutputPath(inputPath)
	ext := strings.ToLower(filepath.Ext(inputPath))

	width, height, err := parseRezie(*resizeFlag)
	if err != nil {
		fmt.Println("Error parsing resize dimensions:", err)
		os.Exit(1)
	}

	intensity := *intensityFlag
	if intensity < 1 || intensity > 100 {
		fmt.Println("Intensity must be between 1 and 100")
		os.Exit(1)
	}

	fmt.Printf("Processing image: %s\n", inputPath)
	fmt.Printf("Output will be saved to: %s\n", outputPath)

	if ext == ".gif" {
		images, err := internal.LoadGIF(inputPath)
		if err != nil {
			fmt.Println("Error loading GIF: ", err)
			os.Exit(1)
		}
		for idx, frame := range images.Image {
			img := frame
			if width > 0 || height > 0 {
				resized := internal.ResizeImage(img, width, height)
				if paletted, ok := resized.(*image.Paletted); ok {
					img = paletted
				} else {
					bounds := resized.Bounds()
					palettedImg := image.NewPaletted(bounds, frame.Palette)
					draw.Draw(palettedImg, bounds, resized, bounds.Min, draw.Over)
					img = palettedImg
				}
			}
			processedImage := internal.ApplyLowPoly(img, intensity)
			gifFrame := image.NewPaletted(processedImage.Bounds(), frame.Palette)
			draw.Draw(gifFrame, gifFrame.Rect, processedImage, processedImage.Bounds().Min, draw.Over)
			images.Image[idx] = gifFrame
		}
		outFile, err := os.Create(outputPath)
		if err != nil {
			fmt.Println("Error creating output file: ", err)
			os.Exit(1)
		}
		defer outFile.Close()
		if err := gif.EncodeAll(outFile, images); err != nil {
			fmt.Println("Error saving GIF: ", err)
			os.Exit(1)
		}
		fmt.Println("GIF processed successfully")
		return
	} else {
		img, format, err := internal.LoadImage(inputPath, ext)
		if err != nil {
			fmt.Println("Error loading image:", err)
			os.Exit(1)
		}

		if width > 0 || height > 0 {
			fmt.Printf("Resizing image to %dx%d\n", width, height)
			img = internal.ResizeImage(img, width, height)
		}

		lowPolyImage := internal.ApplyLowPoly(img, intensity)
		if err := internal.SaveImage(lowPolyImage, outputPath, format); err != nil {
			fmt.Println("Error saving image:", err)
			os.Exit(1)
		}
	}

	fmt.Println("Image processing complete. Low-poly image saved successfully.")
}
