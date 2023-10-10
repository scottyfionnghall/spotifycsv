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
)

type Song struct {
	ArtistName string `json:"artist_name"`
	TrackName  string `json:"track_name"`
	AlbumName  string `json:"album_name"`
	SpotifyId  string `json:"spotfy_id"`
}

func createJsonFile(dir string, file string) {
	err := os.Mkdir("json", 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	f, err := os.Open(dir + file)
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(f)
	csvReader.FieldsPerRecord = -1

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
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

	jsonData, err := json.Marshal(playlist)
	if err != nil {
		log.Fatal(err)
	}

	jsonFile, err := os.Create(fmt.Sprintf("./json/%s.json", file[:len(file)-4]))
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(jsonData)
}

func createSpotifyLinkFiles(dir string, file string) {
	err := os.Mkdir("links", 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	f, err := os.Open(dir + file)
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(f)
	csvReader.FieldsPerRecord = -1

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	output_file, err := os.Create(fmt.Sprintf("./links/%s.txt", file[:len(file)-4]))
	defer output_file.Close()
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i < len(data); i++ {
		_, err := output_file.WriteString(fmt.Sprintf("https://open.spotify.com/track/%s\n", data[i][0]))
		if err != nil {
			log.Fatal(err)
		}
	}

}

func validatePath(dir string) error {
	if dir == " " {
		return_err := errors.New("empty command-line argument")
		return return_err
	}

	var pattern string
	if runtime.GOOS != "windows" {
		pattern = `^([a-zA-Z0-9_+//]*)`
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

	for _, file := range files {

		valid, err := validateFile(file.Name())

		if err != nil {
			log.Fatal(err)
		}

		if valid {
			if *json {
				createJsonFile(*dir, file.Name())
			}

			if *link {
				createSpotifyLinkFiles(*dir, file.Name())
			}
		} else {
			log.Printf("%s is not a csv file, skipping", file.Name())
		}
	}

	fmt.Println("Everything was successful!\nPress enter to exit")
	fmt.Scanln()
}
