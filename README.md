# file-fairy

A function for the "Execute" plugin for Deluge that organises downloaded media into kids and adults
categories and retrieves subtitles from OpenSubtitles.

Assumes Plex. Make sure you update filepaths in `main.go`

## Requirements

- Deluge with the "Execute" plugin
- OMDb API key
- OpenSubtitles API key, username, and password

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/file-fairy.git
    ```

2. Navigate to the project directory:
    ```sh
    cd file-fairy
    ```

3. Create a `.env` file with your API keys and credentials:
    ```
    OMDB_API_KEY=your_omdb_api_key
    OPENSUBTITLES_API_KEY=your_opensubtitles_api_key
    OPENSUBTITLES_USERNAME=your_username
    OPENSUBTITLES_PASSWORD=your_password
    ```

4. Build the project:
    ```sh
    ./build.sh
    ```

## Usage

1. Configure the Deluge "Execute" plugin to run the generated `execute_script` binary upon torrent completion.
2. The script will automatically sort media and retrieve subtitles.

## License

MIT License. See [LICENSE](LICENSE) for more information.
