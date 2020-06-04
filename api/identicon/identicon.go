package identiconservice

import (
	"crypto/sha512"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"os"

	identiconpb "github.com/murryIsDeveloping/identiconGRPC/api/identicon/proto"
)

// IdenticonService acts as a router for all identicon related services
type IdenticonService struct{}

// GetIdenticon is a grpc stream that takes a name, size and width of image in pixels and returns a identicon image as a stream
func (service *IdenticonService) GetIdenticon(req *identiconpb.Request, responseStream identiconpb.IdenticonService_GetIdenticonServer) error {
	identicon := &Identicon{}
	identicon.SetName(req.GetFileName())
	identicon.SetSize(int(req.GetSize()))
	file, err := identicon.DrawImg(int(req.GetPixelsize()))

	if err != nil {
		panic(err)
	}
	defer file.Close()

	bufferSize := 64 * 1024 //64KiB, tweak this as desired
	buff := make([]byte, bufferSize)

	for {
		bytesRead, err := file.Read(buff)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		resp := &identiconpb.Response{
			FileChunk: buff[:bytesRead],
		}

		err = responseStream.Send(resp)
		if err != nil {
			log.Println("error while sending chunk:", err)
			return err
		}
	}
	return nil
}

func reverse(bytes []byte) []byte {
	if len(bytes) == 0 {
		return bytes
	}
	return append(reverse(bytes[1:]), bytes[0])
}

// Identicon controls the creation of identicons
type Identicon struct {
	size  int
	width int
	name  string
	grid  [][]byte
	hash  []byte
	color color.RGBA
}

// SetName sets the name of the identicon
func (identicon *Identicon) SetName(name string) {
	identicon.name = name
}

// SetSize sets the number of cells across and high of the identicon must be a number between 3 and 10 anything greater or smaller will default to 3 or 10
func (identicon *Identicon) SetSize(size int) {
	if size < 3 {
		size = 3
	}

	if size > 10 {
		size = 10
	}

	identicon.size = size
	identicon.width = int(math.Ceil(float64(size) / 2.0))
}

func (identicon *Identicon) setColor() {
	rgb := identicon.hash[len(identicon.hash)-3:]
	identicon.color = color.RGBA{uint8(rgb[0]), uint8(rgb[1]), uint8(rgb[2]), 255}
}

// DrawImg takes an image width which should be equal to the height and width of the image in pixels and returns an image
func (identicon *Identicon) DrawImg(imageWidth int) (*os.File, error) {
	filename := fmt.Sprintf("/tmp/%v%vx%v-%v.png", identicon.name, identicon.size, identicon.size, imageWidth)
	existingfile, err := os.Open(filename)
	if existingfile != nil {
		return existingfile, nil
	}

	identicon.generateGrid()

	cellSize := int(math.Ceil(float64(imageWidth) / float64(identicon.size)))

	generatedImage := image.NewRGBA(image.Rect(0, 0, imageWidth, imageWidth))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(generatedImage, generatedImage.Bounds(), &image.Uniform{identicon.color}, image.Point{}, draw.Src)

	for h, row := range identicon.grid {
		for v, cell := range row {
			if int(cell)%2 == 0 {
				horStart := h * cellSize
				verStart := v * cellSize
				rect := image.Rect(verStart, horStart, verStart+cellSize, horStart+cellSize)
				draw.Draw(generatedImage, rect, &image.Uniform{white}, image.Point{}, draw.Src)
			}
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	encodingErr := png.Encode(file, generatedImage)

	if encodingErr != nil {
		return nil, encodingErr
	}

	return file, nil
}

func (identicon *Identicon) generateHash() {
	hash := sha512.New()
	hash.Write([]byte(identicon.name))
	identicon.hash = hash.Sum(nil)[:identicon.size*identicon.width]
}

func (identicon *Identicon) generateGrid() {
	identicon.generateHash()
	identicon.setColor()

	isOdd := identicon.size%2 == 1

	grid := [][]byte{}
	chunk := []byte{}

	for i := 0; i < identicon.width*identicon.size; i++ {
		chunk = append(chunk, identicon.hash[i])
		if len(chunk) == identicon.width {
			if isOdd {
				row := append([]byte{}, chunk...)
				grid = append(grid, append(row, reverse(chunk[:identicon.width-1])...))
			} else {
				grid = append(grid, append(chunk, reverse(chunk)...))
			}
			chunk = []byte{}
		}
	}

	identicon.grid = grid
}
