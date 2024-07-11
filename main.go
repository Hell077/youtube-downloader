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

	fmt.Println("Available formats:")

	var audioFormatIndices []int
	formatMap := make(map[int]youtube.Format)

	for i, format := range formats {
		formatMap[i] = format
		formatLabel := format.QualityLabel
		if format.AudioQuality != "" {
			formatLabel += " (with audio)"
			audioFormatIndices = append(audioFormatIndices, i)
		}
		fmt.Printf("%d. %s\n", i+1, formatLabel)
	}

	if len(formatMap) == 0 {
		log.Fatalf("No formats found for video: %s", url)
	}

	var selectedIndex int
	selectedIndex = 0

	err = keyboard.Open()
	if err != nil {
		log.Fatalf("Error opening keyboard: %v", err)
	}
	defer keyboard.Close()

	fmt.Println("\nUse arrow keys to select quality, then Enter to confirm:")

	for {
		screen.Clear()
		screen.MoveTopLeft()

		for index, format := range formats {
			indicator := " "
			if index == selectedIndex {
				indicator = ">"
			}
			formatLabel := format.QualityLabel
			if format.AudioQuality != "" {
				formatLabel += " (with audio)"
			}
			fmt.Printf("%s %s\n", indicator, formatLabel)
		}

		_, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatalf("Error getting key: %v", err)
		}

		switch key {
		case keyboard.KeyArrowDown:
			if selectedIndex < len(formats)-1 {
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
	selectedFormat := formatMap[selectedIndex]
	fmt.Printf("Selected quality: %s\n", selectedFormat.QualityLabel)

	stream, size, err := client.GetStream(video, &selectedFormat)
	if err != nil {
		log.Fatalf("Error getting stream: %v", err)
	}
	defer stream.Close()

	// Create file for download
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		fileName := video.Title + ".mp4"
		path = filepath.Join(path, sanitizeFileName(fileName))
	}

	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	// Progress bar setup
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
		progressbar.OptionSetWidth(60),
	)

	var wg sync.WaitGroup
	wg.Add(1)

	// Download in a goroutine
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
