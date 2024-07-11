package main

import (
	"fmt"
	"github.com/inancgumus/screen"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/eiannone/keyboard"
	"github.com/kkdai/youtube/v2"
)

func main() {
	var url, path string
	fmt.Print("Enter YouTube video URL: ")
	fmt.Scanln(&url)
	fmt.Print("Enter save path: ")
	fmt.Scanln(&path)

	client := youtube.Client{}

	video, err := client.GetVideo(url)
	if err != nil {
		log.Fatalf("Error getting video info: %v", err)
	}
	formats := video.Formats
	if len(formats) == 0 {
		log.Fatalf("No formats found for video: %s", url)
	}

	audioFormats := make([]youtube.Format, 0)
	for _, format := range formats {
		if format.AudioQuality != "" {
			audioFormats = append(audioFormats, format)
		}
	}
	if len(audioFormats) == 0 {
		log.Fatalf("No audio formats found for video: %s", url)
	}

	fmt.Println("Available audio qualities:")
	var qualityMap = make(map[string]youtube.Format)
	for _, format := range audioFormats {
		if format.QualityLabel != "" {
			qualityMap[format.QualityLabel] = format
			fmt.Printf("- %s\n", format.QualityLabel)
		}
	}

	err = keyboard.Open()
	if err != nil {
		log.Fatalf("Error opening terminal for input: %v", err)
	}
	defer keyboard.Close()

	var selectedQuality string
	fmt.Println("\nUse up/down arrows to select quality, then Enter to confirm:")

	var selectedIndex int
	selectedIndex = 0

	for {
		screen.Clear()
		screen.MoveTopLeft()

		for index, format := range audioFormats {
			indicator := " "
			if index == selectedIndex {
				indicator = ">"
				selectedQuality = format.QualityLabel
			}
			fmt.Printf("%s %s\n", indicator, format.QualityLabel)
		}

		_, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatalf("Error getting key: %v", err)
		}

		switch key {
		case keyboard.KeyArrowDown:
			if selectedIndex < len(audioFormats)-1 {
				selectedIndex++
			}
		case keyboard.KeyArrowUp:
			if selectedIndex > 0 {
				selectedIndex--
			}
		case keyboard.KeyEnter:
			goto FormatSelected
		}
	}

FormatSelected:
	format, exists := qualityMap[selectedQuality]
	if !exists {
		log.Fatalf("Invalid quality selected")
	}

	fmt.Printf("Selected quality: %s\n", format.QualityLabel)

	stream, size, err := client.GetStream(video, &format)
	if err != nil {
		log.Fatalf("Error getting stream: %v", err)
	}
	defer stream.Close()

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		fileName := video.Title + ".mp4"
		path = filepath.Join(path, sanitizeFileName(fileName))
	}

	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "#",
			SaucerHead:    "^",
			SaucerPadding: "-",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetWidth(60),
	)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, err := io.Copy(io.MultiWriter(file, bar), stream)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	}()

	wg.Wait()

	fmt.Println("\nVideo successfully downloaded")

	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}

func sanitizeFileName(fileName string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(fileName)
}
