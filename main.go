package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"sync"
	"time"
)

type Song struct {
	ArtistName string `json:"artist_name"`
	TrackName  string `json:"track_name"`
	AlbumName  string `json:"album_name"`
	SpotifyId  string `json:"spotfy_id"`
}

type Worker struct {
	FileName string
	Playlist []Song
}

func gen(files []os.DirEntry, dir string, wg *sync.WaitGroup, done chan struct{}, chans ...chan Worker) {
	defer wg.Done()
	for _, file := range files {
		valid, err := validateFile(file.Name())
		if err != nil {
			log.Fatal(err)
		}
		if valid {
			var result Worker
			playlist, err := createPlaylist(dir, file.Name())
			result.FileName = file.Name()
			result.Playlist = playlist
			if err != nil {
				log.Fatal(err)
			}
			for _, c := range chans {
				c <- result
			}
		}
	}

}

func createPlaylist(dir string, file string) ([]Song, error) {
	f, err := os.Open(dir + file)
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(f)
	csvReader.FieldsPerRecord = -1

	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	var song Song
	var playlist []Song

	for i := 1; i < len(data); i++ {
		song.ArtistName = data[i][4]
		song.TrackName = data[i][2]
		song.AlbumName = data[i][3]
		song.SpotifyId = data[i][0]
		playlist = append(playlist, song)
	}

	return playlist, err
}

func createJsonFile(done chan struct{}) chan Worker {
	out := make(chan Worker)
	go func() {
		defer close(out)
		for {
			select {
			case result := <-out:
				err := os.Mkdir("json", 0750)
				if err != nil && !os.IsExist(err) {
					log.Fatal(err)
				}

				jsonData, err := json.Marshal(result.Playlist)
				if err != nil {
					log.Fatal(err)
				}

				jsonFile, err := os.Create(fmt.Sprintf("./json/%s.json", result.FileName[:len(result.FileName)-4]))
				if err != nil {
					log.Fatal(err)
				}
				defer jsonFile.Close()

				jsonFile.Write(jsonData)
			case <-done:
				break
			}
		}
	}()
	return out
}

func createSpotifyLinkFiles(done chan struct{}) chan Worker {
	out := make(chan Worker)
	go func() {
		defer close(out)
		for {
			select {
			case result := <-out:
				err := os.Mkdir("links", 0750)
				if err != nil && !os.IsExist(err) {
					log.Fatal(err)
				}

				output_file, err := os.Create(fmt.Sprintf("./links/%s.txt", result.FileName[:len(result.FileName)-4]))
				defer output_file.Close()
				if err != nil {
					log.Fatal(err)
				}

				for _, song := range result.Playlist {
					_, err := output_file.WriteString(fmt.Sprintf("https://open.spotify.com/track/%s\n", song.SpotifyId))
					if err != nil {
						log.Fatal(err)
					}
				}
			case <-done:
				break
			}
		}
	}()

	return out
}

func validatePath(dir string) error {
	if dir == " " {
		return_err := errors.New("empty command-line argument")
		return return_err
	}

	var pattern string
	if runtime.GOOS != "windows" {
		pattern = `^(\/[[:ascii:]]*\/)*$`
	} else {
		pattern = `^([A-Z]:\\([a-zA-Z0-9_+]*\\)*)$`
	}

	matched, err := regexp.MatchString(pattern, dir)
	if err != nil || matched != true {
		return_err := errors.New("not a valid path, try adding \\ at the end")
		return return_err
	}
	return nil
}

func validateFile(file string) (bool, error) {
	pattern := "([A-z0-9_+]*\\.csv)"
	matched, err := regexp.MatchString(pattern, file)
	return matched, err
}

func main() {
	dir := flag.String("dir", " ", "Directory to csv files.")
	json := flag.Bool("json", false, "Create JSON files with song info in more pretier format")
	link := flag.Bool("link", false, "Create folder with .txt files containing links to all songs")
	flag.Parse()

	var wg sync.WaitGroup
	done := make(chan struct{})

	if !*json && !*link {
		fmt.Println("\nSpecify either -link or -json flag\nExample spotifycsv -dir \"/foo/bar/\" -link\nIf you need help, use --help argument")
		return
	}

	err := validatePath(*dir)
	if err != nil {
		log.Fatal(err)
	}

	files, err := os.ReadDir(*dir)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	var chans []chan Worker
	if *json {
		chans = append(chans, createJsonFile(done))
	}

	if *link {
		chans = append(chans, createSpotifyLinkFiles(done))
	}

	wg.Add(1)
	go gen(files, *dir, &wg, done, chans...)
	wg.Wait()

	fmt.Println("Everything was successful!")
	fmt.Printf("Program took %s\n", time.Since(start))
	fmt.Println("Press enter to exit")
	fmt.Scanln()
}
