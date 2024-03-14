package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/estesp/manifest-tool/v2/pkg/registry"
	"github.com/estesp/manifest-tool/v2/pkg/store"
	"github.com/estesp/manifest-tool/v2/pkg/types"
	"github.com/estesp/manifest-tool/v2/pkg/util"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	cacheTimeout = time.Hour
)

var (
	tableName  = "imagecache"
	dynaClient dynamodbiface.DynamoDBAPI
)

// ErrorBody is used to encapsulate error responses to the client
type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

// Image represents the cached JSON metadata about the image
type Image struct {
	CacheTS   int64              `json:"cachets"`
	IsList    bool               `json:"islist"`
	ImageName string             `json:"imagename"`
	Digest    string             `json:"digest"`
	MediaType string             `json:"mediatype"`
	ArchList  []ocispec.Platform `json:"archlist"`
}

func main() {
	region := os.Getenv("AWS_REGION")
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return
	}
	dynaClient = dynamodb.New(awsSession)
	lambda.Start(handleRequest)
}

func inspectImage(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	imageName := req.QueryStringParameters["image"]
	if len(imageName) == 0 {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String("No image name provided")})
	}
	// check cache
	if image, err := checkCache(imageName); err == nil {
		return apiResponse(http.StatusOK, image)
	}
	// inspect image
	image, err := queryRegistry(imageName)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String(fmt.Sprintf("Error querying image: %s", err))})
	}
	if err = cacheImage(image); err != nil {
		log.Printf("WARN: unable to cache image: %v", err)
	}

	return apiResponse(http.StatusOK, image)
}

func handleRequest(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return inspectImage(req)
	}
	return apiResponse(http.StatusMethodNotAllowed, "method not allowed")
}

func apiResponse(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	resp.StatusCode = status

	stringBody, _ := json.Marshal(body)
	resp.Body = string(stringBody)
	return &resp, nil
}

func checkCache(imageName string) (*Image, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"imagename": {
				S: aws.String(imageName),
			},
		},
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.GetItem(input)
	if err != nil {
		return nil, errors.New("failed to find image")
	}

	item := new(Image)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, errors.New("failed to unmarshal image cached details")
	}
	cacheTime := time.Unix(item.CacheTS, 0)
	if cacheTime.Add(cacheTimeout).Before(time.Now()) {
		// invalidate
		deleteCache(imageName)
		return nil, errors.New("Cache expired image " + imageName)
	}
	return item, nil
}

func deleteCache(imageName string) {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"imagename": {
				S: aws.String(imageName),
			},
		},
		TableName: aws.String(tableName),
	}

	if _, err := dynaClient.DeleteItem(input); err != nil {
		log.Printf("error deleting cache item for %s: %v", imageName, err)
	}
}

func cacheImage(image *Image) error {
	av, err := dynamodbattribute.MarshalMap(image)
	if err != nil {
		return errors.New("could not marshal image data")
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		return errors.New("could not write to dynamoDB")
	}
	return nil
}

func queryRegistry(name string) (*Image, error) {
	var image *Image
	imageRef, err := util.ParseName(name)
	if err != nil {
		return nil, err
	}

	memoryStore := store.NewMemoryStore()
	if err = util.CreateRegistryHost(imageRef, "", "", false, false, "", false); err != nil {
		return nil, err
	}

	descriptor, err := registry.FetchDescriptor(util.GetResolver(), memoryStore, imageRef)
	if err != nil {
		return nil, err
	}

	_, db, _ := memoryStore.Get(descriptor)
	switch descriptor.MediaType {
	case ocispec.MediaTypeImageIndex, types.MediaTypeDockerSchema2ManifestList:
		// this is a multi-platform image descriptor; marshal to Index type
		var idx ocispec.Index
		if err := json.Unmarshal(db, &idx); err != nil {
			return nil, err
		}
		image = generateImage(name, memoryStore, descriptor, idx, ocispec.Image{})
	case ocispec.MediaTypeImageManifest, types.MediaTypeDockerSchema2Manifest:
		var man ocispec.Manifest
		if err := json.Unmarshal(db, &man); err != nil {
			return nil, err
		}
		_, cb, _ := memoryStore.Get(man.Config)
		var conf ocispec.Image
		if err := json.Unmarshal(cb, &conf); err != nil {
			return nil, err
		}
		image = generateImage(name, memoryStore, descriptor, ocispec.Index{}, conf)
		return image, nil
	default:
		return nil, errors.New("Unknown descriptor type: " + descriptor.MediaType)
	}
	return image, nil
}

func generateImage(name string, cs *store.MemoryStore, desc ocispec.Descriptor, index ocispec.Index, imgConfig ocispec.Image) *Image {
	image := new(Image)
	image.Digest = desc.Digest.String()
	image.MediaType = desc.MediaType
	image.ImageName = name
	image.CacheTS = time.Now().Unix()
	switch desc.MediaType {
	case ocispec.MediaTypeImageIndex, types.MediaTypeDockerSchema2ManifestList:
		image.IsList = true
		for _, img := range index.Manifests {
			image.ArchList = append(image.ArchList, *img.Platform)
		}
	default:
		image.ArchList = []ocispec.Platform{{
			OS:           imgConfig.OS,
			Architecture: imgConfig.Architecture,
		}}
	}
	return image
}
