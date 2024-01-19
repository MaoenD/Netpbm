package Netpbm

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
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

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	var magicNumber string
	if scanner.Scan() {
		magicNumber = scanner.Text()
	} else {
		return nil, fmt.Errorf("unable to read magic number")
	}

	if magicNumber != "P3" && magicNumber != "P6" {
		return nil, fmt.Errorf("unsupported PPM format: %s", magicNumber)
	}

	var width, height int
	if scanner.Scan() {
		_, err := fmt.Sscanf(scanner.Text(), "%d", &width)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unable to read width")
	}

	if scanner.Scan() {
		_, err := fmt.Sscanf(scanner.Text(), "%d", &height)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unable to read height")
	}

	var maxVal int
	if scanner.Scan() {
		_, err := fmt.Sscanf(scanner.Text(), "%d", &maxVal)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unable to read max value")
	}

	data := make([][]Pixel, height)
	for i := 0; i < height; i++ {
		data[i] = make([]Pixel, width)
		for j := 0; j < width; j++ {
			var r, g, b uint8
			if scanner.Scan() {
				_, err := fmt.Sscanf(scanner.Text(), "%d", &r)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("unable to read pixel data")
			}

			if scanner.Scan() {
				_, err := fmt.Sscanf(scanner.Text(), "%d", &g)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("unable to read pixel data")
			}

			if scanner.Scan() {
				_, err := fmt.Sscanf(scanner.Text(), "%d", &b)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("unable to read pixel data")
			}

			data[i][j] = Pixel{R: r, G: g, B: b}
		}
	}

	return &PPM{
		data:        data,
		width:       width,
		height:      height,
		magicNumber: magicNumber,
		max:         uint8(maxVal),
	}, nil
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
func (ppm *PPM) Set(x, y int, value Pixel) {
	ppm.data[y][x] = value
}

// Save saves the PPM image to a file and returns an error if there was a problem.
func (ppm *PPM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	_, err = writer.WriteString(ppm.magicNumber + "\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "%d %d\n", ppm.width, ppm.height)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "%d\n", ppm.max)
	if err != nil {
		return err
	}

	for _, row := range ppm.data {
		for _, pixel := range row {
			_, err = fmt.Fprintf(writer, "%d %d %d ", pixel.R, pixel.G, pixel.B)
			if err != nil {
				return err
			}
		}
		_, err = writer.WriteString("\n")
		if err != nil {
			return err
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
	data := make([][]uint8, ppm.height)
	for i := 0; i < ppm.height; i++ {
		data[i] = make([]uint8, ppm.width)
		for j := 0; j < ppm.width; j++ {
			// Convert RGB to grayscale using luminosity method
			grayValue := uint8(0.2126*float32(ppm.data[i][j].R) + 0.7152*float32(ppm.data[i][j].G) + 0.0722*float32(ppm.data[i][j].B))
			data[i][j] = grayValue
		}
	}

	return &PGM{
		data:        data,
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P5",
		max:         255,
	}
}

// ToPBM converts the PPM image to PBM.
func (ppm *PPM) ToPBM() *PBM {
	data := make([][]bool, ppm.height)
	for i := 0; i < ppm.height; i++ {
		data[i] = make([]bool, ppm.width)
		for j := 0; j < ppm.width; j++ {
			// Convert RGB to binary using a threshold (127)
			grayValue := uint8(0.2126*float32(ppm.data[i][j].R) + 0.7152*float32(ppm.data[i][j].G) + 0.0722*float32(ppm.data[i][j].B))
			data[i][j] = grayValue > 127
		}
	}

	return &PBM{
		data:        data,
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P1",
	}
}

// DrawLine draws a line between two points.
func (ppm *PPM) DrawLine(p1, p2 Point, color Pixel) {
	deltaX := p2.X - p1.X
	deltaY := p2.Y - p1.Y

	if math.Abs(float64(deltaX)) >= math.Abs(float64(deltaY)) {
		if p1.X > p2.X {
			p1, p2 = p2, p1
		}
		for x := p1.X; x <= p2.X; x++ {
			y := p1.Y + deltaY*(x-p1.X)/deltaX
			ppm.Set(x, int(y), color)
		}
	} else {
		if p1.Y > p2.Y {
			p1, p2 = p2, p1
		}
		for y := p1.Y; y <= p2.Y; y++ {
			x := p1.X + deltaX*(y-p1.Y)/deltaY
			ppm.Set(int(x), y, color)
		}
	}
}

// DrawRectangle draws a rectangle.
func (ppm *PPM) DrawRectangle(p1 Point, width, height int, color Pixel) {
	p2 := Point{p1.X + width, p1.Y}
	p3 := Point{p1.X, p1.Y + height}
	p4 := Point{p2.X, p3.Y}

	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p4, color)
	ppm.DrawLine(p4, p3, color)
	ppm.DrawLine(p3, p1, color)
}

// DrawFilledRectangle draws a filled rectangle.
func (ppm *PPM) DrawFilledRectangle(p1 Point, width, height int, color Pixel) {
	for y := p1.Y; y < p1.Y+height; y++ {
		for x := p1.X; x < p1.X+width; x++ {
			ppm.Set(x, y, color)
		}
	}
}

// DrawCircle draws a circle.
func (ppm *PPM) DrawCircle(center Point, radius int, color Pixel) {
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if x*x+y*y <= radius*radius {
				ppm.Set(center.X+x, center.Y+y, color)
			}
		}
	}
}

// DrawFilledCircle draws a filled circle.
func (ppm *PPM) DrawFilledCircle(center Point, radius int, color Pixel) {
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if x*x+y*y <= radius*radius {
				ppm.Set(center.X+x, center.Y+y, color)
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
	for i := 0; i < len(points)-1; i++ {
		ppm.DrawLine(points[i], points[i+1], color)
	}
	ppm.DrawLine(points[len(points)-1], points[0], color)
}

// DrawFilledPolygon draws a filled polygon.
func (ppm *PPM) DrawFilledPolygon(points []Point, color Pixel) {
	minY := points[0].Y
	maxY := points[0].Y

	for _, point := range points {
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	for y := minY; y <= maxY; y++ {
		intersections := []int{}

		for i := 0; i < len(points); i++ {
			j := (i + 1) % len(points)
			if (points[i].Y < y && points[j].Y >= y) || (points[j].Y < y && points[i].Y >= y) {
				x := int(float64(points[i].X) + (float64(y-points[i].Y)/float64(points[j].Y-points[i].Y))*float64(points[j].X-points[i].X))
				intersections = append(intersections, x)
			}
		}

		sort.Ints(intersections)

		for i := 0; i < len(intersections); i += 2 {
			for x := intersections[i]; x <= intersections[i+1]; x++ {
				ppm.Set(x, y, color)
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
