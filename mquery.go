package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dghubble/sling"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const baseURL = "https://2xopp470jc.execute-api.us-east-2.amazonaws.com/mquery"

// QueryParams defines the parameters sent; in our case only "image" is needed
type QueryParams struct {
	Image string `url:"image"`
}

// ErrorResponse holds the payload response on failure HTTP codes
type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// Image contains the JSON struct we get from success
type Image struct {
	CacheTS   int64              `json:"cachets"`
	IsList    bool               `json:"islist"`
	ImageName string             `json:"imagename"`
	Digest    string             `json:"digest"`
	ArchList  []ocispec.Platform `json:"archlist"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("ERROR: Must provide an image name as a command line parameter.\n")
		os.Exit(1)
	}
	qparam := &QueryParams{
		Image: os.Args[1],
	}
	image := new(Image)
	errResp := new(ErrorResponse)
	resp, err := sling.New().Base(baseURL).QueryStruct(qparam).Receive(image, errResp)
	if err != nil {
		fmt.Printf("ERROR: failed to query backend: %v\n", err)
		os.Exit(1)
	}
	os.Exit(processResponse(resp, os.Args[1], errResp, image))
}

func processResponse(resp *http.Response, imageName string, errResp *ErrorResponse, image *Image) int {
	if resp.StatusCode != 200 {
		// non-success RC from our http request
		fmt.Printf("ERROR: %s\n", errResp.Error)
		return 1
	}
	printManifestInfo(imageName, image)
	return 0
}

func printManifestInfo(imageName string, image *Image) {
	fmt.Printf("Image: %s\n", imageName)
	list := "Yes"
	if !image.IsList {
		list = "No"
	}
	fmt.Printf(" * Manifest List: %s\n", list)
	if image.IsList {
		fmt.Println(" * Supported platforms:")
		for _, platform := range image.ArchList {
			platformOutput := parsePlatform(platform)
			fmt.Printf("   - %s\n", platformOutput)
		}
	} else {
		fmt.Printf(" * Supports: %s\n", parsePlatform(image.ArchList[0]))
	}
	fmt.Println("")
}

func parsePlatform(platform ocispec.Platform) string {
	platformStr := fmt.Sprintf("%s/%s", platform.OS, platform.Architecture)
	if len(platform.Variant) > 0 {
		platformStr = platformStr + "/" + platform.Variant
	}
	if platform.OS == "windows" {
		if len(platform.OSVersion) > 0 {
			platformStr = platformStr + ":" + platform.OSVersion
		}
	}
	return platformStr
}
