package Netpbm

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// Pixel represents a color pixel with red (R), green (G), and blue (B) values.
type Pixel struct {
	R, G, B uint8
}

// PPM represents a Portable PixMap image.
type PPM struct {
	data          [][]Pixel
	width, height int
	magicNumber   string
	max           uint8
}

// Point represents a point in the image.
type Point struct {
	X, Y int
}

// ReadPPM reads a PPM image from a file and returns a struct that represents the image.
func ReadPPM(filename string) (*PPM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read magic number
	magicNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %v", err)
	}
	magicNumber = strings.TrimSpace(magicNumber)
	if magicNumber != "P3" && magicNumber != "P6" {
		return nil, fmt.Errorf("invalid magic number: %s", magicNumber)
	}

	// Read dimensions
	dimensions, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading dimensions: %v", err)
	}
	var width, height int
	_, err = fmt.Sscanf(strings.TrimSpace(dimensions), "%d %d", &width, &height)
	if err != nil {
		return nil, fmt.Errorf("invalid dimensions: %v", err)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width and height must be positive")
	}

	// Read max value
	maxValue, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading max value: %v", err)
	}
	maxValue = strings.TrimSpace(maxValue)
	var max int
	_, err = fmt.Sscanf(maxValue, "%d", &max)
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %v", err)
	}

	// Read image data
	data := make([][]Pixel, height)
	expectedBytesPerPixel := 3

	if magicNumber == "P3" {
		// Read P3 format (ASCII)
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)
			rowData := make([]Pixel, width)
			for x := 0; x < width; x++ {
				if x*3+2 >= len(fields) {
					return nil, fmt.Errorf("index out of range at row %d, column %d", y, x)
				}
				var pixel Pixel
				_, err := fmt.Sscanf(fields[x*3], "%d", &pixel.R)
				if err != nil {
					return nil, fmt.Errorf("error parsing Red value at row %d, column %d: %v", y, x, err)
				}
				_, err = fmt.Sscanf(fields[x*3+1], "%d", &pixel.G)
				if err != nil {
					return nil, fmt.Errorf("error parsing Green value at row %d, column %d: %v", y, x, err)
				}
				_, err = fmt.Sscanf(fields[x*3+2], "%d", &pixel.B)
				if err != nil {
					return nil, fmt.Errorf("error parsing Blue value at row %d, column %d: %v", y, x, err)
				}
				rowData[x] = pixel
			}
			data[y] = rowData
		}
	} else if magicNumber == "P6" {
		// Read P6 format (binary)
		for y := 0; y < height; y++ {
			row := make([]byte, width*expectedBytesPerPixel)
			n, err := reader.Read(row)
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unexpected end of file at row %d", y)
				}
				return nil, fmt.Errorf("error reading pixel data at row %d: %v", y, err)
			}
			if n < width*expectedBytesPerPixel {
				return nil, fmt.Errorf("unexpected end of file at row %d, expected %d bytes, got %d", y, width*expectedBytesPerPixel, n)
			}

			rowData := make([]Pixel, width)
			for x := 0; x < width; x++ {
				pixel := Pixel{R: row[x*expectedBytesPerPixel], G: row[x*expectedBytesPerPixel+1], B: row[x*expectedBytesPerPixel+2]}
				rowData[x] = pixel
			}
			data[y] = rowData
		}
	}

	// Return the PPM struct
	return &PPM{data, width, height, magicNumber, uint8(max)}, nil
}

// Size returns the width and height of the image.
func (ppm *PPM) Size() (int, int) {
	return ppm.width, ppm.height
}

// At returns the value of the pixel at (x, y).
func (ppm *PPM) At(x, y int) Pixel {
	return ppm.data[y][x]
}

// Set sets the value of the pixel at (x, y).
func (ppm *PPM) Set(x, y int, color Pixel) {
	if x >= 0 && x < ppm.width && y >= 0 && y < ppm.height {
		ppm.data[y][x] = color
	}
}

// Save saves the PPM image to a file and returns an error if there was a problem.
func (ppm *PPM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if ppm.magicNumber == "P6" || ppm.magicNumber == "P3" {
		fmt.Fprintf(file, "%s\n%d %d\n%d\n", ppm.magicNumber, ppm.width, ppm.height, ppm.max)
	} else {
		err = fmt.Errorf("magic number error")
		return err
	}

	//bytesPerPixel := 3 // Nombre d'octets par pixel pour P6

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			pixel := ppm.data[y][x]
			if ppm.magicNumber == "P6" {
				// Conversion inverse des pixels
				file.Write([]byte{pixel.R, pixel.G, pixel.B})
			} else if ppm.magicNumber == "P3" {
				// Conversion inverse des pixels
				fmt.Fprintf(file, "%d %d %d ", pixel.R, pixel.G, pixel.B)
			}
		}
		if ppm.magicNumber == "P3" {
			fmt.Fprint(file, "\n")
		}
	}

	return nil
}

// Invert inverts the colors of the PPM image.
func (ppm *PPM) Invert() {
	for i := 0; i < ppm.height; i++ {
		for j := 0; j < ppm.width; j++ {
			ppm.data[i][j].R = uint8(ppm.max) - ppm.data[i][j].R
			ppm.data[i][j].G = uint8(ppm.max) - ppm.data[i][j].G
			ppm.data[i][j].B = uint8(ppm.max) - ppm.data[i][j].B
		}
	}
}

// Flip flips the PPM image horizontally.
func (ppm *PPM) Flip() {
	for i := 0; i < ppm.height; i++ {
		for j := 0; j < ppm.width/2; j++ {
			ppm.data[i][j], ppm.data[i][ppm.width-j-1] = ppm.data[i][ppm.width-j-1], ppm.data[i][j]
		}
	}
}

// Flop flops the PPM image vertically.
func (ppm *PPM) Flop() {
	for i := 0; i < ppm.height/2; i++ {
		ppm.data[i], ppm.data[ppm.height-i-1] = ppm.data[ppm.height-i-1], ppm.data[i]
	}
}

// SetMagicNumber sets the magic number of the PPM image.
func (ppm *PPM) SetMagicNumber(magicNumber string) {
	ppm.magicNumber = magicNumber
}

// SetMaxValue sets the max value of the PPM image.
func (ppm *PPM) SetMaxValue(maxValue uint8) {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Scale the RGB values based on the new max value
			ppm.data[y][x].R = uint8(float64(ppm.data[y][x].R) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].G = uint8(float64(ppm.data[y][x].G) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].B = uint8(float64(ppm.data[y][x].B) * float64(maxValue) / float64(ppm.max))
		}
	}

	// Update the max value
	ppm.max = uint8(int(maxValue))
}

// Rotate90CW rotates the PPM image 90° clockwise.
func (ppm *PPM) Rotate90CW() {
	newData := make([][]Pixel, ppm.width)
	for i := 0; i < ppm.width; i++ {
		newData[i] = make([]Pixel, ppm.height)
	}

	for i := 0; i < ppm.height; i++ {
		for j := 0; j < ppm.width; j++ {
			newData[j][ppm.height-i-1] = ppm.data[i][j]
		}
	}

	ppm.data = newData
	ppm.width, ppm.height = ppm.height, ppm.width
}

// ToPGM converts the PPM image to PGM.
func (ppm *PPM) ToPGM() *PGM {
	pgm := &PGM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P2",
		max:         ppm.max,
	}

	pgm.data = make([][]uint8, ppm.height)
	for i := range pgm.data {
		pgm.data[i] = make([]uint8, ppm.width)
	}

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Convert RGB to grayscale
			gray := uint8((int(ppm.data[y][x].R) + int(ppm.data[y][x].G) + int(ppm.data[y][x].B)) / 3)
			pgm.data[y][x] = gray
		}
	}

	return pgm
}

// ToPBM converts the PPM image to PBM.
func (ppm *PPM) ToPBM() *PBM {
	pbm := &PBM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P1",
	}

	pbm.data = make([][]bool, ppm.height)
	for i := range pbm.data {
		pbm.data[i] = make([]bool, ppm.width)
	}

	// Set a threshold for binary conversion
	threshold := uint8(ppm.max / 2)

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Calculate the average intensity of RGB values
			average := (uint16(ppm.data[y][x].R) + uint16(ppm.data[y][x].G) + uint16(ppm.data[y][x].B)) / 3
			// Set the binary value based on the threshold
			pbm.data[y][x] = average < uint16(threshold)
		}
	}
	return pbm
}

// DrawLine draws a line between two points.
func (ppm *PPM) DrawLine(p1, p2 Point, color Pixel) {
	deltaX := p2.X - p1.X
	deltaY := p2.Y - p1.Y

	steep := math.Abs(float64(deltaY)) > math.Abs(float64(deltaX))

	if steep {
		p1.X, p1.Y = p1.Y, p1.X
		p2.X, p2.Y = p2.Y, p2.X
		deltaX, deltaY = deltaY, deltaX
	}

	if p1.X > p2.X {
		p1.X, p2.X = p2.X, p1.X
		p1.Y, p2.Y = p2.Y, p1.Y
		deltaX, deltaY = -deltaX, -deltaY
	}

	deltaErr := math.Abs(float64(deltaY) / float64(deltaX))
	error := 0.0
	y := p1.Y

	for x := p1.X; x <= p2.X; x++ {
		if steep {
			if y >= 0 && y < len(ppm.data) && x >= 0 && x < len(ppm.data[y]) {
				ppm.Set(y, x, color)
			}
		} else {
			if x >= 0 && x < len(ppm.data) && y >= 0 && y < len(ppm.data[x]) {
				ppm.Set(x, y, color)
			}
		}
		error += deltaErr
		if error >= 0.5 {
			if deltaY > 0 {
				y++
			} else {
				y--
			}
			error -= 1.0
		}
	}
}

// DrawRectangle draws a rectangle.
func (ppm *PPM) DrawRectangle(p1 Point, width, height int, color Pixel) {
	// Draw the four sides of the rectangle using DrawLine.
	p2 := Point{p1.X + width, p1.Y}
	p3 := Point{p1.X + width, p1.Y + height}
	p4 := Point{p1.X, p1.Y + height}

	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p4, color)
	ppm.DrawLine(p4, p1, color)
}

// DrawFilledRectangle draws a filled rectangle on a PPM image.
func (ppm *PPM) DrawFilledRectangle(p1 Point, width, height int, color Pixel) {
	// Ensuring the rectangle doesn't exceed the image bounds
	maxX := min(p1.X+width, ppm.width)
	maxY := min(p1.Y+height, ppm.height)

	for x := p1.X; x <= maxX; x++ { // Include maxX in the loop
		for y := p1.Y; y <= maxY; y++ { // Include maxY in the loop
			if x >= 0 && y >= 0 && x < ppm.width && y < ppm.height {
				ppm.Set(x, y, color)
			}
		}
	}
}

// DrawCircle draws a circle.
func (ppm *PPM) DrawCircle(center Point, radius int, color Pixel) {

	for x := 0; x < ppm.height; x++ {
		for y := 0; y < ppm.width; y++ {
			dx := float64(x) - float64(center.X)
			dy := float64(y) - float64(center.Y)
			distance := math.Sqrt(dx*dx + dy*dy)

			if math.Abs(distance-float64(radius)) < 1.0 && distance < float64(radius) {
				ppm.Set(x, y, color)
			}
		}
	}
	ppm.Set(center.X-(radius-1), center.Y, color)
	ppm.Set(center.X+(radius-1), center.Y, color)
	ppm.Set(center.X, center.Y+(radius-1), color)
	ppm.Set(center.X, center.Y-(radius-1), color)
}

// DrawFilledCircle draws a filled circle.
func (ppm *PPM) DrawFilledCircle(center Point, radius int, color Pixel) {
	ppm.DrawCircle(center, radius, color)

	for i := 0; i < ppm.height; i++ {
		var positions []int
		var number_points int
		for j := 0; j < ppm.width; j++ {
			if ppm.data[i][j] == color {
				number_points += 1
				positions = append(positions, j)
			}
		}
		if number_points > 1 {
			for k := positions[0] + 1; k < positions[len(positions)-1]; k++ {
				ppm.data[i][k] = color

			}
		}
	}
}

// DrawTriangle draws a triangle.
func (ppm *PPM) DrawTriangle(p1, p2, p3 Point, color Pixel) {
	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p1, color)
}

// DrawFilledTriangle draws a filled triangle.
func (ppm *PPM) DrawFilledTriangle(p1, p2, p3 Point, color Pixel) {
	minX := min(min(p1.X, p2.X), p3.X)
	minY := min(min(p1.Y, p2.Y), p3.Y)
	maxX := max(max(p1.X, p2.X), p3.X)
	maxY := max(max(p1.Y, p2.Y), p3.Y)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			p := Point{x, y}
			if isInsideTriangle(p, p1, p2, p3) {
				ppm.Set(x, y, color)
			}
		}
	}
}

// DrawPolygon draws a polygon.
func (ppm *PPM) DrawPolygon(points []Point, color Pixel) {
	numPoints := len(points)
	if numPoints < 3 {
		// A polygon must have at least 3 vertices
		return
	}

	// Draw lines between consecutive points to form the polygon
	for i := 0; i < numPoints-1; i++ {
		ppm.DrawLine(points[i], points[i+1], color)
	}

	// Draw the last line connecting the last and first points to close the polygon
	ppm.DrawLine(points[numPoints-1], points[0], color)
}

// DrawFilledPolygon draws a filled polygon.
func (ppm *PPM) DrawFilledPolygon(points []Point, color Pixel) {
	ppm.DrawPolygon(points, color)
	for i := 0; i < ppm.height; i++ {
		var positions []int
		var number_points int
		for j := 0; j < ppm.width; j++ {
			if ppm.data[i][j] == color {
				number_points += 1
				positions = append(positions, j)
			}
		}
		if number_points > 1 {
			for k := positions[0] + 1; k < positions[len(positions)-1]; k++ {
				ppm.data[i][k] = color

			}
		}
	}
}

func isInsideTriangle(p, p1, p2, p3 Point) bool {
	area := 0.5 * (float64(-p2.Y)*float64(p3.X) + float64(p1.Y)*(float64(-p2.X)+float64(p3.X)) + float64(p1.X)*(float64(p2.Y)-float64(p3.Y)) + float64(p2.X)*float64(p3.Y))
	if area == 0 {
		return false
	}

	s := 1 / (2 * area) * (float64(p1.Y)*float64(p3.X) - float64(p1.X)*float64(p3.Y) + (float64(p3.Y)-float64(p1.Y))*float64(p.X) + (float64(p1.X)-float64(p3.X))*float64(p.Y))
	t := 1 / (2 * area) * (float64(p1.X)*float64(p2.Y) - float64(p1.Y)*float64(p2.X) + (float64(p1.Y)-float64(p2.Y))*float64(p.X) + (float64(p2.X)-float64(p1.X))*float64(p.Y))

	return s >= 0 && t >= 0 && (s+t) <= 1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
