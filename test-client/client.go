package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"

	identiconpb "github.com/murryIsDeveloping/identiconGRPC/api/identicon/proto"
	grpc "google.golang.org/grpc"
)

func main() {
	namePtr := flag.String("name", "foo", "name of the identicon will change apperance of identicon")
	sizePtr := flag.Int("size", 5, "A value between 3 and 10")
	widthPtr := flag.Int("width", 500, "the width and height of img in pixels")

	flag.Parse()

	conn, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal("client could connect to grpc service:", err)
	}
	c := identiconpb.NewIdenticonServiceClient(conn)
	fileStreamResponse, err := c.GetIdenticon(context.TODO(), &identiconpb.Request{
		FileName:  *namePtr,
		Size:      int32(*sizePtr),
		Pixelsize: int32(*widthPtr),
	})
	if err != nil {
		log.Println("error downloading:", err)
		return
	}

	imageBytes := []byte{}
	for {
		chunkResponse, err := fileStreamResponse.Recv()
		if err == io.EOF {
			log.Println("received all chunks")
			break
		}
		if err != nil {
			log.Println("err receiving chunk:", err)
			break
		}
		imageBytes = append(imageBytes, chunkResponse.FileChunk...)
	}

	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Create(fmt.Sprintf("./tmp/%v.png", *namePtr))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Println("Error encoding to png")
		log.Fatal(err)
	}
}
