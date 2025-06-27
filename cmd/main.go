package main

import (
	"flag"
	"fmt"
	"image/gif"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/brother-blue/low-poly-converter/internal"
	"github.com/schollz/progressbar/v3"
)

var (
	resizeFlag       = flag.String("resize", "", "Resize the image to the specified dimensions (e.g., 800x600)")
	intensityFlag    = flag.Int("intensity", 100, "Set the intensity of the image processing (1-100)")
	debugFlag        = flag.Bool("debug", false, "Enable debug mode (creates prof files for use with `go tool pprof ... ./cpu.prof`)")
	showProgressFlag = flag.Bool("showProgress", false, "Show progress bar during processing (only for GIFs)")
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

func handleGifProcess(inputPath, outputPath string, width, height, intensity int) {
	images, err := internal.LoadGIF(inputPath)
	if err != nil {
		fmt.Println("Error loading GIF: ", err)
		os.Exit(1)
	}
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating output file: ", err)
		os.Exit(1)
	}
	defer outFile.Close()

	var bar *progressbar.ProgressBar
	if *showProgressFlag {
		bar = progressbar.NewOptions(
			len(images.Image),
			progressbar.OptionSetDescription("Processing frames..."),
		)
	}

	if err := gif.EncodeAll(
		outFile,
		internal.ResizeGif(images, width, height, intensity, bar)); err != nil {
		fmt.Println("Error saving GIF: ", err)
		os.Exit(1)
	}
	fmt.Println("GIF processed successfully")
}

func handleStaticImageProcess(inputPath, outputPath, ext string, width, height, intensity int) {
	var bar *progressbar.ProgressBar
	if *showProgressFlag {
		bar = progressbar.NewOptions(
			1,
			progressbar.OptionSetDescription("Processing image..."),
		)
	}
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
	bar.Add(1)
	if err := internal.SaveImage(lowPolyImage, outputPath, format); err != nil {
		fmt.Println("Error saving image:", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: poly-convert [options] <input image>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *debugFlag {
		f, _ := os.Create("cpu.prof")
		pprof.StartCPUProfile(f)
		defer f.Close()
		defer pprof.StopCPUProfile()

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic:", r)
				fmt.Println("Stack trace:", debug.Stack())
			}
		}()
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

	switch ext {
	case ".gif":
		handleGifProcess(inputPath, outputPath, width, height, intensity)
	case ".jpg", ".jpeg", ".png":
		handleStaticImageProcess(inputPath, outputPath, ext, width, height, intensity)
	default:
		fmt.Printf("Unsupported image format: %s\n", ext)
	}
	fmt.Println("Image processing complete. Low-poly image saved successfully.")

	if *debugFlag {
		f1, _ := os.Create("mem.prof")
		pprof.WriteHeapProfile(f1)
		f1.Close()
	}
}
