package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	http.HandleFunc("/deploy", deployHandler)
	fmt.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

type Request struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func deployHandler(w http.ResponseWriter, r *http.Request) {
	auth := os.Getenv("AUTH_TOKEN")
	if auth == "" {
		http.Error(w, "Server authentication token not set", http.StatusInternalServerError)
		return
	}

	clientToken := r.Header.Get("Authorization")
	if clientToken != auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.URL == "" {
		http.Error(w, "Missing Name / URL in request", http.StatusBadRequest)
		return
	}

	composeFilePath := filepath.Join(os.TempDir(), req.Name+".yml")
	err = downloadFile(composeFilePath, req.URL)
	if err != nil {
		http.Error(w, "Failed to download compose file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cmd := exec.Command("docker-compose", "-f", composeFilePath, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		http.Error(w, "Failed to pull images: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cmd = exec.Command("docker-compose", "-p", req.Name, "-f", composeFilePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		http.Error(w, "Failed to bring up stack: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.RemoveAll(composeFilePath)
	if err != nil {
		http.Error(w, "Failed to clean up compose file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Stack %s deployed successfully", req.Name)
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
