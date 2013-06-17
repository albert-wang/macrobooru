package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"macrobooru/api/client"
	"macrobooru/models"
)

func getImage(cfg Config, imageID string) (*models.Image, error) {
	query := client.NewQuery()
	client, _ := client.NewClient(cfg.Endpoint + "/v2/api")

	result := []models.Image{}

	query.Add("Image", &result).
		Where(map[string]interface{}{
		"pid =": imageID,
	})

	err := query.Execute(client)

	if err != nil {
		log.Print("Failed to get image")
		return nil, err
	}

	if len(result) != 1 {
		log.Printf("Could not find a image with id %s", imageID)
		return nil, fmt.Errorf("Could not find an image with id %s", imageID)
	}

	return &result[0], nil
}

func downloadImage(cfg Config, image *models.Image) (string, error) {
	//Construct download path.
	mapping := map[string]string{
		"image/jpeg": "jpeg",
		"image/png":  "png",
		"image/gif":  "gif",
	}

	val, ok := mapping[image.Mime]
	if !ok {
		//Not a supported image type.
		return "", fmt.Errorf("Unsupported image mime: %s", image.Mime)
	}

	path := fmt.Sprintf("%s/img/%s.%s", cfg.Endpoint, image.Filehash, val)
	resp, err := http.Get(path)
	defer resp.Body.Close()

	if err != nil {
		log.Print(err)
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Print("Could not download the image file at %s", path)
		return "", fmt.Errorf("Could not download the image file at %s", path)
	}

	tempfile, err := ioutil.TempFile("", "macrobooru-")
	temppath := tempfile.Name()
	defer tempfile.Close()

	if err != nil {
		return "", err
	}

	_, err = io.Copy(tempfile, resp.Body)
	if err != nil {
		return "", err
	}

	return temppath, nil
}

func imageSize(path string) (int, int, error) {
	cmd := exec.Command("identify", "-format", "%[fx:w] %[fx:h]", path)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Split(string(output), " ")

	x, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
	if err != nil {
		return 0, 0, err
	}

	y, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
	if err != nil {
		return 0, 0, err
	}

	return int(x), int(y), nil
}

func createTemporaryFileWithContents(contents string) (string, error) {
	tempfile, err := ioutil.TempFile("", "macrobooru-")
	if err != nil {
		return "", err
	}

	defer tempfile.Close()

	_, err = tempfile.Write([]byte(contents))
	if err != nil {
		return "", nil
	}

	return tempfile.Name(), nil
}

func uploadToHost(cfg Config, path string) (string, error) {
	buffer := bytes.Buffer{}
	w := multipart.NewWriter(&buffer)

	lbl, err := w.CreateFormField("email")
	if err != nil {
		return "", err
	}

	lbl.Write([]byte(cfg.UploaderEmail))

	file, err := w.CreateFormFile("file", path)
	if err != nil {
		return "", err
	}

	fd, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer fd.Close()
	_, err = io.Copy(file, fd)
	if err != nil {
		return "", err
	}

	w.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/upload/curl", cfg.Endpoint), &buffer)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	responseBuffer := bytes.Buffer{}
	io.Copy(&responseBuffer, res.Body)
	res.Body.Close()

	return string(responseBuffer.Bytes()), nil
}

func CreateMacro(cfg Config, imageID string, topCaption string, bottomCaption string) (string, error) {
	image, err := getImage(cfg, imageID)
	if err != nil {
		return "", err
	}

	path, err := downloadImage(cfg, image)
	defer os.Remove(path)

	if err != nil {
		return "", err
	}

	w, h, err := imageSize(path)
	if err != nil {
		return "", err
	}

	//Generate a temporary path.
	tempfile, err := ioutil.TempFile("", "macrobooru-")
	if err != nil {
		return "", err
	}

	defer tempfile.Close()
	temppath := tempfile.Name()
	defer os.Remove(temppath)

	topFile, err := createTemporaryFileWithContents(topCaption)
	if err != nil {
		return "", err
	}

	defer os.Remove(topFile)

	bottomFile, err := createTemporaryFileWithContents(bottomCaption)
	if err != nil {
		return "", err
	}

	defer os.Remove(bottomFile)

	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("convert", path,
		"-background", "transparent",
		"-font", fmt.Sprintf("%s/impact.ttf", pwd),
		"-fill", "white",
		"-stroke", "black",
		"-strokewidth", "6",
		"-size", fmt.Sprintf("%dx%d", w, h/4),
		"-gravity", "center",
		fmt.Sprintf("caption:@%s", topFile),
		"-geometry", "+10+10",
		"-gravity", "north",
		"-composite",
		"-gravity", "center",
		fmt.Sprintf("caption:@%s", bottomFile),
		"-gravity", "south",
		"-geometry", "+10+10",
		"-composite", temppath)

	log.Print(cmd)
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Print(string(out))
		return "", err
	}

	resp, err := uploadToHost(cfg, temppath)
	if err != nil {
		return "", nil
	}

	log.Print(resp)
	return "", nil
}
