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
	fmt.Print("Введите URL видео на YouTube: ")
	fmt.Scanln(&url)
	fmt.Print("Введите путь для сохранения: ")
	fmt.Scanln(&path)

	client := youtube.Client{}

	video, err := client.GetVideo(url)
	if err != nil {
		log.Fatalf("Ошибка при получении информации о видео: %v", err)
	}
	formats := video.Formats
	if len(formats) == 0 {
		log.Fatalf("Не найдены форматы для видео: %s", url)
	}

	audioFormats := make([]youtube.Format, 0)
	for _, format := range formats {
		if format.AudioQuality != "" {
			audioFormats = append(audioFormats, format)
		}
	}
	if len(audioFormats) == 0 {
		log.Fatalf("Не найдены форматы с аудио для видео: %s", url)
	}

	fmt.Println("Доступные качества с аудио:")
	var qualityMap = make(map[string]youtube.Format)
	for _, format := range audioFormats {
		if format.QualityLabel != "" {
			qualityMap[format.QualityLabel] = format
			fmt.Printf("- %s\n", format.QualityLabel)
		}
	}

	err = keyboard.Open()
	if err != nil {
		log.Fatalf("Ошибка открытия терминала для ввода: %v", err)
	}
	defer keyboard.Close()

	var selectedQuality string
	fmt.Println("\nИспользуйте стрелки вверх/вниз для выбора качества, затем Enter для подтверждения:")

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
			log.Fatalf("Ошибка при получении клавиши: %v", err)
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
		log.Fatalf("Выбрано недопустимое качество")
	}

	fmt.Printf("Выбрано качество: %s\n", format.QualityLabel)

	stream, size, err := client.GetStream(video, &format)
	if err != nil {
		log.Fatalf("Ошибка при получении потока: %v", err)
	}
	defer stream.Close()

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		fileName := video.Title + ".mp4"
		path = filepath.Join(path, sanitizeFileName(fileName))
	}

	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Ошибка при создании файла: %v", err)
	}
	defer file.Close()

	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription("Загрузка"),
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
			log.Fatalf("Ошибка при записи в файл: %v", err)
		}
	}()

	wg.Wait()

	fmt.Println("\nВидео успешно загружено")

	fmt.Println("Нажмите Enter для выхода...")
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
