package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"html"
	"log"
	"net/http"

	"os"
	"os/exec"

	guuid "github.com/google/uuid"
)

type StreamResponse struct {
	Id  string
	Url string
}
type StreamResponseABR struct {
	Id       string
	Url360p  string
	Url480p  string
	Url720p  string
	Url1080p string
}

func main() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	// Don't allow accessing /streams/ route
	// (only allow `/streams/:uuid/` or `/streams/:uuid/:uuid.m3u8`)
	http.HandleFunc("/streams/", func(w http.ResponseWriter, r *http.Request) {
		// serves the stream of original resolution
		if r.URL.Path != "/streams/" {
			println(path + " now " + r.URL.Path)
			http.ServeFile(w, r, r.URL.Path[1:])
		} else {
			println("no id provided!")
			http.NotFound(w, r)
			return
		}
	})

	http.HandleFunc("/streamsabr/", func(w http.ResponseWriter, r *http.Request) {
		// serves the stream of specified resolution
		if r.URL.Path != "/streamsabr/" {
			println(path + " now " + r.URL.Path)
			http.ServeFile(w, r, r.URL.Path[1:])
		} else {
			println("no id provided!")
			http.NotFound(w, r)
			return
		}
	})

	http.HandleFunc("/submitabr/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/submitabr/" {
			http.NotFound(w, r)
			return
		}

		if r.Method == "GET" {
			fmt.Fprintf(w, "GET, %q", html.EscapeString(r.URL.Path))
		} else if r.Method == "POST" {
			println(r.FormValue("input"))
			id := guuid.New()
			cmd1 := exec.Command("mkdir", "streamsabr/"+id.String())
			cmd1.Output()
			outputPath := "./streamsabr/" + id.String()
			urlinput := r.FormValue("input")
			println("outpath", outputPath)

			cmda := exec.Command("ffmpeg", "-i", urlinput, "-profile:v", "baseline", "-level", "3.0", "-s", "640x360", "-start_number", "0", "-hls_time", "10", "-hls_list_size", "0", "-f", "hls", outputPath+"/360p.m3u8")
			cmda.Output()
			cmdb := exec.Command("ffmpeg", "-i", urlinput, "-profile:v", "baseline", "-level", "3.0", "-s", "842x480", "-start_number", "0", "-hls_time", "10", "-hls_list_size", "0", "-f", "hls", outputPath+"/480p.m3u8")
			cmdb.Output()
			cmdc := exec.Command("ffmpeg", "-i", urlinput, "-profile:v", "baseline", "-level", "3.0", "-s", "1280x720", "-start_number", "0", "-hls_time", "10", "-hls_list_size", "0", "-f", "hls", outputPath+"/720p.m3u8")
			cmdc.Output()
			cmdd := exec.Command("ffmpeg", "-i", urlinput, "-profile:v", "baseline", "-level", "3.0", "-s", "1920x1080", "-start_number", "0", "-hls_time", "10", "-hls_list_size", "0", "-f", "hls", outputPath+"/1080p.m3u8")
			cmdd.Output()

			if err != nil {
				println(err.Error())
				if err.Error() == "exit status 1" {
					fmt.Fprintf(w, "Please use a valid url (which returns a mp4 file)")
					println("Please use a valid url (which returns a mp4 file)")
				}
				return
			}
			hostURL := "http://" + r.Host + "/streamsabr/"
			streamResponseABR := StreamResponseABR{
				id.String(),
				hostURL + id.String() + "/360p.m3u8",
				hostURL + id.String() + "/480p.m3u8",
				hostURL + id.String() + "/720p.m3u8",
				hostURL + id.String() + "/1080p.m3u8",
			}
			js, err2 := json.Marshal(streamResponseABR)
			if err2 != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(js)

		} else {
			http.Error(w, "Invalid request method.", 405)
		}

	})

	http.HandleFunc("/submit/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/submit/" {
			http.NotFound(w, r)
			return
		}

		if r.Method == "GET" {
			fmt.Fprintf(w, "GET, %q", html.EscapeString(r.URL.Path))
		} else if r.Method == "POST" {
			println(r.FormValue("input"))
			id := guuid.New()
			outputfilename := id.String() + ".m3u8"
			cmd1 := exec.Command("mkdir", "streams/"+id.String())
			cmd1.Output()
			outputPath := "./streams/" + id.String() + "/" + outputfilename
			urlinput := r.FormValue("input")
			resolution, e := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "default=nw=1", urlinput).Output()
			if e != nil {
				log.Fatal(e)
			}
			println("resolution", resolution)
			width := strings.Split(strings.Split(string(resolution), "\n")[0], "=")[1]
			height := strings.Split(strings.Split(string(resolution), "\n")[1], "=")[1]
			reso := width + "x" + height
			// using same resolution as the original video
			fmt.Printf("The resolution is %s\n", reso)

			cmd := exec.Command("ffmpeg", "-i", urlinput, "-profile:v", "baseline", "-level", "3.0", "-s", reso, "-start_number", "0", "-hls_time", "10", "-hls_list_size", "0", "-f", "hls", outputPath)
			stdout, err := cmd.Output()

			if err != nil {
				println(err.Error())
				if err.Error() == "exit status 1" {
					fmt.Fprintf(w, "Please use a valid url (which returns a mp4 file)")
					println("Please use a valid url (which returns a mp4 file)")
				}
				return
			}
			hostURL := "http://" + r.Host + "/streams/"
			streamResponse := StreamResponse{id.String(), hostURL + id.String() + "/" + id.String() + ".m3u8"}
			js, err2 := json.Marshal(streamResponse)
			if err2 != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(js)

			println(string(stdout))
		} else {
			http.Error(w, "Invalid request method.", 405)
		}

	})

	http.ListenAndServe(":80", nil)
}
