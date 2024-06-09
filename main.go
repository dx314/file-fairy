package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	//	"strconv"
	"github.com/TheForgotten69/go-opensubtitles/opensubtitles"
	"strings"
	"time"
)

const (
	kidsMoviesDir  = "/var/lib/plexmediaserver/Movies"
	kidsTVDir      = "/var/lib/plexmediaserver/TV"
	adultMoviesDir = "/var/lib/plexmediaserver/adult/Movies"
	adultTVDir     = "/var/lib/plexmediaserver/adult/TV"
)

var (
	omdbAPIKey             string
	openSubtitlesUserAgent = "AutoSubber v1.0.0"
	openSubtitlesAPIKey    string
	openSubtitlesUsername  string
	openSubtitlesPassword  string
)

// Struct to hold the JSON response from OMDb
type OMDbResponse struct {
	Title    string `json:"Title"`
	Year     string `json:"Year"`
	Rated    string `json:"Rated"`
	Released string `json:"Released"`
	Runtime  string `json:"Runtime"`
	Genre    string `json:"Genre"`
	Director string `json:"Director"`
	Writer   string `json:"Writer"`
	Actors   string `json:"Actors"`
	Plot     string `json:"Plot"`
	Language string `json:"Language"`
	Country  string `json:"Country"`
	Awards   string `json:"Awards"`
	Poster   string `json:"Poster"`
	Ratings  []struct {
		Source string `json:"Source"`
		Value  string `json:"Value"`
	} `json:"Ratings"`
	Metascore  string `json:"Metascore"`
	ImdbRating string `json:"imdbRating"`
	ImdbVotes  string `json:"imdbVotes"`
	ImdbID     string `json:"imdbID"`
	Type       string `json:"Type"`
	Dvd        string `json:"DVD"`
	BoxOffice  string `json:"BoxOffice"`
	Production string `json:"Production"`
	Website    string `json:"Website"`
	Response   string `json:"Response"`
	Error      string `json:"Error"`
}

func main() {
	if omdbAPIKey == "" || openSubtitlesAPIKey == "" || openSubtitlesUsername == "" || openSubtitlesPassword == "" {
		fmt.Println("Environment variables OMDB_API_KEY, OPENSUBTITLES_API_KEY, OPENSUBTITLES_USERNAME, and OPENSUBTITLES_PASSWORD must be set.")
		os.Exit(1)
	}

	dryRun := flag.Bool("dry-run", false, "Simulate the operation without moving the folder")
	flag.Parse()

	if flag.NArg() < 3 {
		fmt.Println("Usage: ./execute_script [--dry-run] <TorrentID> <Torrent Name> <Torrent Path>")
		return
	}

	torrentID := flag.Arg(0)
	torrentName := flag.Arg(1)
	torrentPath := flag.Arg(2)

	fmt.Printf("TorrentID: %s, Torrent Name: %s, Torrent Path: %s\n", torrentID, torrentName, torrentPath)

	movieTitle, movieYear := parseTorrentName(torrentName)
	if movieTitle == "" || movieYear == "" {
		fmt.Printf("Could not parse movie title or year from torrent name: %s\n", torrentName)
		return
	}

	result, err := fetchOMDbData(movieTitle, movieYear)
	if err != nil {
		fmt.Printf("Error fetching data from OMDb: %v\n", err)
		return
	}

	fmt.Println("Parsed OMDb response:")
	fmt.Printf("Title: %s, Year: %s, Rated: %s, Type: %s\n", result.Title, result.Year, result.Rated, result.Type)

	destinationDir := getDestinationDirectory(result)
	destinationPath := filepath.Join(destinationDir, torrentName)

	if *dryRun {
		fmt.Printf("Dry run: would move %s to %s\n", torrentPath, destinationPath)
	} else {
		// OpenSubtitles integration
		if err := downloadSubtitles(result, filepath.Join(torrentPath, torrentName)); err != nil {
			fmt.Printf("Error downloading subtitles: %v\n", err)
		}

		err = os.Rename(filepath.Join(torrentPath, torrentName), destinationPath)
		if err != nil {
			fmt.Printf("Error moving folder: %v\n", err)
		} else {
			fmt.Printf("Moved %s to %s\n", torrentPath, destinationPath)
		}
	}
}

func parseTorrentName(torrentName string) (string, string) {
	re := regexp.MustCompile(`^(.*?)[\s\._-]*\(?(\d{4})\)?`)
	matches := re.FindStringSubmatch(torrentName)

	if len(matches) > 2 {
		title := strings.ReplaceAll(matches[1], ".", " ")
		year := matches[2]
		return strings.TrimSpace(title), year
	}

	return "", ""
}

func fetchOMDbData(title, year string) (OMDbResponse, error) {
	url := fmt.Sprintf("http://www.omdbapi.com/?t=%s&y=%s&apikey=%s", title, year, omdbAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		return OMDbResponse{}, err
	}
	defer resp.Body.Close()

	var result OMDbResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return OMDbResponse{}, err
	}

	if result.Response == "False" {
		return OMDbResponse{}, fmt.Errorf("OMDb API error: %s", result.Error)
	}

	return result, nil
}

func getDestinationDirectory(response OMDbResponse) string {
	isChildAppropriate := response.Rated == "G" || response.Rated == "PG"
	if isChildAppropriate {
		if response.Type == "movie" {
			return kidsMoviesDir
		}
		return kidsTVDir
	} else {
		if response.Type == "movie" {
			return adultMoviesDir
		}
		return adultTVDir
	}
}

func downloadSubtitles(movie OMDbResponse, moviePath string) error {
	opensubtitles.UserAgent = openSubtitlesUserAgent
	client := opensubtitles.NewClient(nil, "", opensubtitles.Credentials{
		Username: openSubtitlesUsername,
		Password: openSubtitlesPassword,
	}, openSubtitlesAPIKey)

	client, err := client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to OpenSubtitles: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srtFilePath := strings.Replace(moviePath, filepath.Ext(moviePath), ".srt", 1)

	if _, err := os.Stat(srtFilePath); err == nil {
		fmt.Printf("Subtitle already exists for: %s, skipping...\n", movie.Title)
		return nil
	}

	fmt.Printf("Searching subtitles for: %s\n", movie.Title)
	//	movies, _, err := client.Find.Features(ctx, &opensubtitles.FeatureOptions{ImdbID: movie.ImdbID})
	//	if err != nil {
	//		return fmt.Errorf("error searching for movie: %v", err)
	//	}
	//
	//
	//	if len(movies.Items) == 0 {
	//		return fmt.Errorf("no movies found for title: %s", movie.Title)
	//	}
	//
	//	imdbID, err := strconv.Atoi(movie.ImdbID)
	//	if err != nil {
	//		return fmt.Errorf("error converting movie ID: %v", err)
	//	}

	subtitles, _, err := client.Find.Subtitles(ctx, &opensubtitles.SubtitlesOptions{ImdbID: movie.ImdbID, Languages: "en"})
	if err != nil {
		return fmt.Errorf("error searching for subtitles: %v", err)
	}

	if len(subtitles.Items) == 0 {
		return fmt.Errorf("no subtitles found for movie: %s", movie.Title)
	}

	subtitle := subtitles.Items[0]
	if len(subtitle.Attributes.Files) == 0 {
		return fmt.Errorf("no files found for subtitle: %s", movie.Title)
	}

	subtitleFile := subtitle.Attributes.Files[0]
	subtitleDownload, _, err := client.Download.Download(ctx, &opensubtitles.DownloadOptions{FileID: subtitleFile.FileID})
	if err != nil {
		return fmt.Errorf("error downloading subtitle: %v", err)
	}

	if err := os.WriteFile(srtFilePath, []byte(subtitleDownload.Link), 0644); err != nil {
		return fmt.Errorf("error writing subtitle file: %v", err)
	}

	fmt.Printf("Downloaded subtitles for %s and saved as %s\n", movie.Title, srtFilePath)
	return nil
}
