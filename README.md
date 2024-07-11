# YouTube Video Downloader in Go

This is a simple console application written in Go that allows you to download YouTube videos by providing the video URL and the desired download path.

## WARNING
- Use only English folder path name

like this

  ```sh
        C:\Video
  ```


## Features

- Select video quality with audio
- Progress bar for download status
- Simple and user-friendly console interface

## Requirements

- Go 1.20 or later
- A terminal supporting keyboard input

## Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/Hell077/youtube-downloader-go
    ```

2. Navigate to the project directory:

    ```sh
    cd your-repository-name
    ```

3. Install dependencies:

    ```sh
    go mod tidy
    ```

## Usage

1. Build the project:

    ```sh
    go build -o youtube-downloader
    ```

2. Run the program:

    ```sh
    ./youtube-downloader
    ```

OR 

Run the exe file from the repository

  ```sh
      youtubeDownloaderGo.exe
   ```

3. Follow the on-screen instructions to input the YouTube video URL and the download path.

4. Use the up/down arrow keys to select the desired video quality and press Enter to confirm.

5. Wait for the download to complete. The progress will be shown on the screen.

## Example

```plaintext
Enter YouTube video URL: https://www.youtube.com/watch?v=dQw4w9WgXcQ
Enter save path: /path/to/save

Available audio qualities:
- 360p
- 720p

Use up/down arrows to select quality, then Enter to confirm:
> 720p
  360p

Downloading [========================================] 100% 3.7 MB/s
Video successfully downloaded
