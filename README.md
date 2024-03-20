# Movie Grabber

The Movie Grabber is a Go application that interacts with the qBittorrent WEB API to download movies from [topdezfilmes](https://topdezfilmes.de/).

- It will automaticaly download the subtitle and torrent file.
- It will move the subtitle file to the selected directory.
- It will automaticaly add the selected torrent file.
- It will automaticaly remoce the unecessary files.

## Requirements

Install [Qbittorrent](https://www.qbittorrent.org/) and activate the user interface:

1. Open qBittorrent and go to "Tools" > "Options" or "Preferences".
2. Navigate to the "Web UI" or "Web" section in the options/preferences window.
3. Enable the Web UI and set the listening port to 8080.
4. Optionally, configure authentication settings for the Web UI (username and password).
5. Save the changes and close the options/preferences window.

To validate the process access the Web UI using a web browser, use http://localhost:8080 or http://*:8080.

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/manasomali/movie-grabber.git
   ```

2. Install dependencies:
    ```bash
    go mod tidy
    ```

3. Build the project:
    ```bash
    go build
    ```
4. Run the application:
    ```bash
    ./movie-grabber
    ```