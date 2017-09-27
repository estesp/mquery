package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dghubble/sling"
)

const baseURL = "https://openwhisk.ng.bluemix.net/api/v1/web/estesp%40us.ibm.com_dev/default/archList.json"

// QueryParams defines the parameters sent; in our case only "image" is needed
type QueryParams struct {
	Image string `url:"image"`
}

// ImageDataResponse holds the payload response on success
type ImageDataResponse struct {
	ImageData Payload `json:"payload,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// Payload contains the JSON struct we get from the web action
type Payload struct {
	ManifestList string   `json:"manifestList,omitempty"`
	Tag          string   `json:"tag,omitempty"`
	ID           string   `json:"_id,omitempty"`
	RepoTags     []string `json:"repoTags,omitempty"`
	ArchList     []string `json:"archList,omitempty"`
	Platform     string   `json:"platform,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("ERROR: Must provide an image name as a command line parameter.\n")
		os.Exit(1)
	}
	qparam := &QueryParams{
		Image: os.Args[1],
	}
	response := new(ImageDataResponse)
	resp, err := sling.New().Base(baseURL).QueryStruct(qparam).ReceiveSuccess(response)
	if err != nil {
		fmt.Printf("ERROR: failed to query backend: %v\n", err)
		os.Exit(1)
	}
	os.Exit(processResponse(resp, os.Args[1], response))
}

func processResponse(resp *http.Response, imageName string, response *ImageDataResponse) int {
	if resp.StatusCode != 200 {
		// non-success RC from our http request
		fmt.Printf("ERROR: Failure code from our HTTP request: %d\n", resp.StatusCode)
		return 1
	}
	if response.Error != "" {
		// print out error
		fmt.Printf("ERROR: %s\n", response.Error)
		return 1
	}
	printManifestInfo(imageName, response)
	return 0
}

func printManifestInfo(imageName string, response *ImageDataResponse) {
	fmt.Printf("Image: %s\n", imageName)
	fmt.Printf(" * Manifest List: %s\n", response.ImageData.ManifestList)
	if strings.Compare(response.ImageData.ManifestList, "Yes") == 0 {
		fmt.Println(" * Supported platforms:")
		for _, archosPair := range response.ImageData.ArchList {
			fmt.Printf("   - %s\n", archosPair)
		}
	} else {
		fmt.Printf(" * Supports: %s\n", response.ImageData.Platform)
	}
	fmt.Println("")
}
