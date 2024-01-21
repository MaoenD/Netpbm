package Netpbm

import (
	"bufio"
	"fmt"
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

// ReadPPM reads a PPM image from a file and returns a struct that represents the image.
func ReadPPM(filename string) (*PPM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close() // Open the specified file, return an error if needed and ensures the file will be closed at the end of the function.

	reader := bufio.NewReader(file)

	magicNumber, err := reader.ReadString('\n') // Read the first line to get the magic number P3 or P6.
	if err != nil {
		return nil, err
	}
	magicNumber = strings.TrimSpace(magicNumber) // Trim whitespace.
	if magicNumber != "P3" && magicNumber != "P6" {
		return nil, err // Return an error if the magic number is neither P3 nor P6.
	}

	dimensions, err := reader.ReadString('\n') // Read the next line to get the image dimensions.
	if err != nil {
		return nil, err // Return an error if the read fails.
	}
	var width, height int

	_, err = fmt.Sscanf(strings.TrimSpace(dimensions), "%d %d", &width, &height) // Parse the line to extract width and height.
	if err != nil {
		return nil, err // Return an error if the parsing fails.
	}

	maxValue, err := reader.ReadString('\n') // Read the next line to get the maximum color value.
	if err != nil {
		return nil, err // Return an error if the read fails.
	}
	var max int

	_, err = fmt.Sscanf(strings.TrimSpace(maxValue), "%d", &max) // Parse the line to extract the maximum value.
	if err != nil {
		return nil, err // Return an error if the parsing fails.
	}

	data := make([][]Pixel, height) // Initialize a slice of slices to store the image data.
	expectedBytesPerPixel := 3      // Expected number of bytes per pixel.

	if magicNumber == "P3" {
		// Handle P3 format ASCII.
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, err // Return an error if needed.
			}
			fields := strings.Fields(line) // Split the line into fields.
			rowData := make([]Pixel, width)
			for x := 0; x < width; x++ {
				if x*3+2 >= len(fields) {
					return nil, err // Return an error if the data is incomplete.
				}
				var r, g, b int

				_, err = fmt.Sscanf(fields[x*3], "%d", &r)
				if err != nil {
					return nil, err
				}
				_, err = fmt.Sscanf(fields[x*3+1], "%d", &g)
				if err != nil {
					return nil, err
				}
				_, err = fmt.Sscanf(fields[x*3+2], "%d", &b)
				if err != nil {
					return nil, err
				}
				rowData[x] = Pixel{R: uint8(r), G: uint8(g), B: uint8(b)} // Read the RGB values of each pixel. and store them in the rowData slice. Errorwill appear if needed.
			}
			data[y] = rowData // Add the row of pixels to the image data.
		}
	} else if magicNumber == "P6" {
		// Handle P6 format binary.
		for y := 0; y < height; y++ {
			row := make([]byte, width*expectedBytesPerPixel)
			_, err = reader.Read(row)
			if err != nil {
				return nil, err // Return an error if needed.
			}
			rowData := make([]Pixel, width)
			for x := 0; x < width; x++ {

				rowData[x] = Pixel{R: row[x*expectedBytesPerPixel], G: row[x*expectedBytesPerPixel+1], B: row[x*expectedBytesPerPixel+2]} // Extract the RGB values for each pixel.
			}
			data[y] = rowData // Add the row of pixels to the image data.
		}
	}

	// Create and return a new PPM object with the read data.
	return &PPM{data, width, height, magicNumber, uint8(max)}, nil
}

// Size returns the width and height of the image.
func (ppm *PPM) Size() (int, int) {
	return ppm.width, ppm.height // This line returns the width and height of the PPM. 'ppm.width' and 'ppm.height' are accessing the fields 'width' and 'height' from the PPM struct.
}

// At returns the value of the pixel at (x, y).
func (ppm *PPM) At(x, y int) Pixel {
	return ppm.data[y][x] // This line returns the pixel at the specified coordinates. accesses the y-th row (assuming y is within the range [0, height-1])then accesses the x-th pixel in this row (assuming x is within the range [0, width-1]).
}

// Set sets the value of the pixel at (x, y).
func (ppm *PPM) Set(x, y int, color Pixel) {
	if x >= 0 && x < ppm.width && y >= 0 && y < ppm.height { // Checks if the provided coordinates are within the bounds of the image and 'ppm.width' and 'ppm.height' are used to ensure 'x' and 'y' are valid indices.

		ppm.data[y][x] = color // Sets the pixel at the specified coordinates to the new color and the assignment replaces its color with the provided 'color'.
	}
	// PS: If 'x' or 'y' are out of bounds, the method does nothing.
}

// Save saves the PPM image to a file and returns an error if there was a problem.
func (ppm *PPM) Save(filename string) error {

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close() // Create or overwrite a file with the specified filename, secure that the file is closed when the function exits and return an error if file creation fails.

	// Check if the magic number is either P3 or P6, which are valid PPM formats.
	if ppm.magicNumber == "P6" || ppm.magicNumber == "P3" {
		fmt.Fprintf(file, "%s\n%d %d\n%d\n", ppm.magicNumber, ppm.width, ppm.height, ppm.max) // Write the header information to the file.
	} else {
		return fmt.Errorf("magic number error")
	}

	for y := 0; y < ppm.height; y++ { // Iterate over each pixel in the image.
		for x := 0; x < ppm.width; x++ {
			pixel := ppm.data[y][x] // Get the pixel at coordinates (x, y).

			if ppm.magicNumber == "P6" { // If the format is P6 (binary), write the pixel data as binary.
				file.Write([]byte{pixel.R, pixel.G, pixel.B}) // Write pixel colors as bytes.

			} else if ppm.magicNumber == "P3" { // If the format is P3 (ASCII), write the pixel data as text.
				fmt.Fprintf(file, "%d %d %d ", pixel.R, pixel.G, pixel.B) // it allows to write pixel colors as integers.
			}
		}
		if ppm.magicNumber == "P3" {
			fmt.Fprint(file, "\n") // Add a newline after each row if the format is P3.
		}
	}

	return nil // Return nil to indicate success.
}

// Invert inverts the colors of the PPM image.
func (ppm *PPM) Invert() {
	for i := 0; i < ppm.height; i++ { // Iterate over each row of the image.

		for j := 0; j < ppm.width; j++ { // Iterate over each column in the current row.

			ppm.data[i][j].R = uint8(ppm.max) - ppm.data[i][j].R // Invert the red component of the pixel.
			ppm.data[i][j].G = uint8(ppm.max) - ppm.data[i][j].G // Invert the green component of the pixel.
			ppm.data[i][j].B = uint8(ppm.max) - ppm.data[i][j].B // Invert the blue component of the pixel.
		} // Invert the component of the pixel at (i, j) by subtracting it from the maximum color value. The result is then stored back in the component of the pixel, effectively inverting its RGB value.
	}
}

// Flip flips the PPM image horizontally.
func (ppm *PPM) Flip() {

	for i := 0; i < ppm.height; i++ { // Iterate over each row of the image.
		for j := 0; j < ppm.width/2; j++ { // Iterate over the first half of the columns in the current row.The loop goes up to half the width of the image, as we are swapping pixels symmetrically.

			ppm.data[i][j], ppm.data[i][ppm.width-j-1] = ppm.data[i][ppm.width-j-1], ppm.data[i][j] // Swap the pixel at position j with its counterpart from the other side of the row. ppm.data[i][j] is a pixel on the left side of the row. ppm.data[i][ppm.width-j-1] is the corresponding pixel on the right side.
		}
	}
}

// Flop flops the PPM image vertically.
func (ppm *PPM) Flop() {
	for i := 0; i < ppm.height/2; i++ { // Iterate over the first half of the rows in the image.

		ppm.data[i], ppm.data[ppm.height-i-1] = ppm.data[ppm.height-i-1], ppm.data[i] // Swap the current row with its corresponding row in the bottom half of the image.// ppm.data[i] is the current row in the top half then the swapping is done using Go's tuple assignment, since it's more concise and efficient.
	}
}

// SetMagicNumber sets the magic number of the PPM image.
func (ppm *PPM) SetMagicNumber(magicNumber string) {
	ppm.magicNumber = magicNumber // Set the magic number of the PPM image. The magic number is stored in the variable "magicNumber". The function takes a string as an argument and sets the variable to the value of the argument.
}

// SetMaxValue sets the max value of the PPM image.
func (ppm *PPM) SetMaxValue(maxValue uint8) {
	for y := 0; y < ppm.height; y++ { // Iterate over each row of the image.
		for x := 0; x < ppm.width; x++ { // Iterate over each pixel in the current row.

			ppm.data[y][x].R = uint8(float64(ppm.data[y][x].R) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].G = uint8(float64(ppm.data[y][x].G) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].B = uint8(float64(ppm.data[y][x].B) * float64(maxValue) / float64(ppm.max))
		} // Scale the RGB component of the pixel to the new maximum value. by multiplying the current value by the ratio of the new maximum value to the old maximum value.
	}
	ppm.max = maxValue // Update the max value in the PPM struct to the new maximum value.
}

// Rotate90CW rotates the PPM image 90° clockwise.
func (ppm *PPM) Rotate90CW() {
	newData := make([][]Pixel, ppm.width) // Create a new slice to hold the rotated image data in the new data's dimensions will be transposed: width becomes height and vice versa.
	for i := 0; i < ppm.width; i++ {
		newData[i] = make([]Pixel, ppm.height) // Initialize newData with dimensions transposed from the original image. new rows equal to original width, new columns equal to original height.
	}
	for i := 0; i < ppm.height; i++ { // Iterate over each pixel of the original image.
		for j := 0; j < ppm.width; j++ {
			newData[j][ppm.height-i-1] = ppm.data[i][j] // Calculate the new position of the current pixel in the rotated image.The pixel at (i, j) in the original image moves to (j, height-i-1) in the rotated image.
		}
	}

	ppm.data = newData                            // Update the PPM instance's data with the new, rotated image data.
	ppm.width, ppm.height = ppm.height, ppm.width // Swap the width and height to reflect the rotation.
}

// ToPGM converts the PPM image to PGM.
func (ppm *PPM) ToPGM() *PGM {

	pgm := &PGM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P2",
		max:         ppm.max,
	} // I created a new PGM struct with the same dimensions and max value as the PPM image and set the magic number to "P2", which represents a plain PGM format.

	pgm.data = make([][]uint8, ppm.height)
	for i := range pgm.data {
		pgm.data[i] = make([]uint8, ppm.width)
	} // Initialize the 2D slice for grayscale data.

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			gray := uint8((int(ppm.data[y][x].R) + int(ppm.data[y][x].G) + int(ppm.data[y][x].B)) / 3)
			pgm.data[y][x] = gray
		} // Convert the RGB values to grayscale using the average method .The average grayscale value is calculated by averaging the R, G, and B values and then assign the calculated grayscale value to the corresponding pixel in the PGM data.
	}
	return pgm // Return the new PGM image.
}

// ToPBM converts the PPM image to PBM.
func (ppm *PPM) ToPBM() *PBM {
	pbm := &PBM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P1", // Initialize a new PBM struct with the same dimensions as the PPM image and then set the magic number to "P1", representing a plain PBM format.
	}

	pbm.data = make([][]bool, ppm.height)
	for i := range pbm.data {
		pbm.data[i] = make([]bool, ppm.width)
	} // Initialize the 2D slice for binary data.

	threshold := uint8(ppm.max / 2) // Set a threshold for the binary conversion if the pixels are brighter than this threshold, will be white and if darker will be black.

	for y := 0; y < ppm.height; y++ { // Iterate over each pixel in the PPM image.
		for x := 0; x < ppm.width; x++ {
			average := (uint16(ppm.data[y][x].R) + uint16(ppm.data[y][x].G) + uint16(ppm.data[y][x].B)) / 3
			pbm.data[y][x] = average < uint16(threshold)
		} // Calculate the average intensity of the RGB values.Determine if the pixel should be black or white based on the threshold, if the average intensity is less than the threshold, it's set to black (true), otherwise white (false).
	}

	return pbm // Return the new PBM image.
}

type Point struct {
	X, Y int
}

// DrawLine draws a line between two points.
func (ppm *PPM) DrawLine(p1, p2 Point, color Pixel) { // Based on Bresenham's line algorithm.

	deltaX := p2.X - p1.X
	deltaY := p2.Y - p1.Y // Calculate the differences in x and y directions between the two points.

	steep := math.Abs(float64(deltaY)) > math.Abs(float64(deltaX)) // Determine if the line is steep.

	if steep {
		// If the line is steep, swap x and y coordinates of both points.
		p1.X, p1.Y = p1.Y, p1.X
		p2.X, p2.Y = p2.Y, p2.X
		deltaX, deltaY = deltaY, deltaX
	}

	if p1.X > p2.X {
		// If the starting point is to the right of the ending point, swap the points.
		p1.X, p2.X = p2.X, p1.X
		p1.Y, p2.Y = p2.Y, p1.Y
		deltaX, deltaY = -deltaX, -deltaY
	}

	deltaErr := math.Abs(float64(deltaY) / float64(deltaX)) // Calculate the error factor.
	error := 0.0                                            // Initialize error.
	y := p1.Y

	for x := p1.X; x <= p2.X; x++ { // Iterate over x-coordinates.
		if steep {
			// Plot the point with swapped coordinates for steep lines.
			if y >= 0 && y < len(ppm.data) && x >= 0 && x < len(ppm.data[y]) {
				ppm.Set(y, x, color)
			}
		} else {
			// Plot the point with original coordinates for non-steep lines.
			if x >= 0 && x < len(ppm.data) && y >= 0 && y < len(ppm.data[x]) {
				ppm.Set(x, y, color)
			}
		}

		error += deltaErr // Increment the error.
		if error >= 0.5 {

			if deltaY > 0 { // Adjust y-coordinate based on the sign of deltaY.
				y++ // Move up if deltaY is positive.
			} else {
				y-- // Move down if deltaY is negative.
			}
			error -= 1.0 // Adjust the error after changing the y-coordinate.
		}
	}
}

// DrawRectangle draws a rectangle.
func (ppm *PPM) DrawRectangle(p1 Point, width, height int, color Pixel) {

	p2 := Point{p1.X + width, p1.Y}
	p3 := Point{p1.X + width, p1.Y + height}
	p4 := Point{p1.X, p1.Y + height}
	// Define the other three corners of the rectangle based on p1, width, and height.Top-right cornerBottom-right cornerBottom-left corner.

	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p4, color)
	ppm.DrawLine(p4, p1, color)
} // Draw the four sides of the rectangle using the DrawLine method in order top side right side bottom side left side.

// DrawFilledRectangle draws a filled rectangle on a PPM image.
func (ppm *PPM) DrawFilledRectangle(p1 Point, width, height int, color Pixel) {
	maxX := min(p1.X+width, ppm.width)
	maxY := min(p1.Y+height, ppm.height)
	// Calculate the bounds of the rectangle, ensuring it doesn't exceed the image dimensions.

	for x := p1.X; x <= maxX; x++ { // Iterate over the rectangle's area to set each pixel's color.
		for y := p1.Y; y <= maxY; y++ { // Include maxY in the loop.
			if x >= 0 && y >= 0 && x < ppm.width && y < ppm.height {
				ppm.Set(x, y, color)
			} // Check if the current coordinates are within the image boundaries and sets the color of the pixel at (x, y).
		}
	}
}

// DrawCircle draws a circle.
func (ppm *PPM) DrawCircle(center Point, radius int, color Pixel) {

	for x := 0; x < ppm.height; x++ { // Iterate over each pixel in the image.
		for y := 0; y < ppm.width; y++ {

			dx := float64(x) - float64(center.X)
			dy := float64(y) - float64(center.Y)
			distance := math.Sqrt(dx*dx + dy*dy)
			// Calculate the distance from the current pixel to the center of the circle.

			if math.Abs(distance-float64(radius)) < 1.0 && distance < float64(radius) {
				ppm.Set(x, y, color)
			} // Check if the pixel lies on the circumference of the circle.
		}
	}
	ppm.Set(center.X-(radius-1), center.Y, color)
	ppm.Set(center.X+(radius-1), center.Y, color)
	ppm.Set(center.X, center.Y+(radius-1), color)
	ppm.Set(center.X, center.Y-(radius-1), color)
} // Draw additional points to ensure the circle is properly formed.

// DrawFilledCircle draws a filled circle.
func (ppm *PPM) DrawFilledCircle(center Point, radius int, color Pixel) {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			dx := float64(x) - float64(center.X)
			dy := float64(y) - float64(center.Y)
			distanceSquared := dx*dx + dy*dy

			if distanceSquared < float64(radius*radius) {
				ppm.Set(x, y, color)
			}
		}
	}
}

// DrawTriangle draws a triangle.
func (ppm *PPM) DrawTriangle(p1, p2, p3 Point, color Pixel) {
	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p1, color)
} // Draw lines between the three vertices to form a triangle, from p1 to p2, p2 to p3 and p3 to p1.

// DrawFilledTriangle draws a filled triangle.
func (ppm *PPM) DrawFilledTriangle(p1, p2, p3 Point, color Pixel) {

	minX := min(min(p1.X, p2.X), p3.X)
	minY := min(min(p1.Y, p2.Y), p3.Y)
	maxX := max(max(p1.X, p2.X), p3.X)
	maxY := max(max(p1.Y, p2.Y), p3.Y)
	// Find the minimum and maximum x and y coordinates to form the bounding box of the triangle.

	for x := minX; x <= maxX; x++ { // Iterate over each pixel in the bounding box.
		for y := minY; y <= maxY; y++ {
			p := Point{x, y} // Current point being considered.

			if isInsideTriangle(p, p1, p2, p3) { // Check if the current point is inside the triangle.
				ppm.Set(x, y, color)
			} // If the point is inside the triangle, set its color.
		}
	}
}

func isInsideTriangle(p, p1, p2, p3 Point) bool {
	area := 0.5 * (float64(-p2.Y)*float64(p3.X) + float64(p1.Y)*(float64(-p2.X)+float64(p3.X)) + float64(p1.X)*(float64(p2.Y)-float64(p3.Y)) + float64(p2.X)*float64(p3.Y)) // Calculate the area of the triangle using the determinant method.

	if area == 0 {
		return false
	} // If the area is 0, the points are collinear and no triangle is formed.

	s := 1 / (2 * area) * (float64(p1.Y)*float64(p3.X) - float64(p1.X)*float64(p3.Y) + (float64(p3.Y)-float64(p1.Y))*float64(p.X) + (float64(p1.X)-float64(p3.X))*float64(p.Y))
	t := 1 / (2 * area) * (float64(p1.X)*float64(p2.Y) - float64(p1.Y)*float64(p2.X) + (float64(p1.Y)-float64(p2.Y))*float64(p.X) + (float64(p2.X)-float64(p1.X))*float64(p.Y))
	// Calculate the barycentric coordinates s and t for the point p.

	return s >= 0 && t >= 0 && (s+t) <= 1 // Check if the point p lies inside the triangle.The point is inside the triangle if s and t are greater than or equal to 0, and the sum of s and t is less than or equal to 1.
}

// DrawPolygon draws a polygon.
func (ppm *PPM) DrawPolygon(points []Point, color Pixel) {
	numPoints := len(points)
	if numPoints < 3 {
		// A polygon must have at least 3 vertices.
		return
	}

	// Draw lines between consecutive points to form the polygon.
	for i := 0; i < numPoints-1; i++ {
		ppm.DrawLine(points[i], points[i+1], color)
	}

	// Draw the last line connecting the last and first points to close the polygon.
	ppm.DrawLine(points[numPoints-1], points[0], color)
}

// DrawFilledPolygon draws a filled polygon.
func (ppm *PPM) DrawFilledPolygon(points []Point, color Pixel) {

	ppm.DrawPolygon(points, color) // First, draw the outline of the polygon.

	for i := 0; i < ppm.height; i++ { // Iterate over each row of the image.
		var positions []int   // To store the x-positions where the polygon's edge is found.
		var number_points int // Count of points found on this row.

		for j := 0; j < ppm.width; j++ { // Check each pixel in the row.
			if ppm.data[i][j] == color {
				number_points += 1
				positions = append(positions, j)
			} // If a pixel is part of the polygon's edge, record its position.
		}

		// If more than one edge point is found on the row, fill the space between them.
		if number_points > 1 {
			for k := positions[0] + 1; k < positions[len(positions)-1]; k++ {
				ppm.data[i][k] = color // Fill the pixels between the first and last edge points.
			}
		}
	}
}

func (ppm *PPM) DrawKochSnowflake(n int, start Point, width int, color Pixel) { //It doesn't work but i let it there since it was hard to come up with it.
	height := width * int(math.Sqrt(3)) / 2
	p1 := start
	p2 := Point{start.X + width, start.Y}
	p3 := Point{start.X + width/2, start.Y - height}

	// Recursively draw the three sides of the triangle.
	ppm.drawKochLine(n, p1, p2, color)
	ppm.drawKochLine(n, p2, p3, color)
	ppm.drawKochLine(n, p3, p1, color)
}

func (ppm *PPM) drawKochLine(n int, p1, p2 Point, color Pixel) {
	if n == 0 {
		ppm.DrawLine(p1, p2, color)
	} else {
		dx, dy := p2.X-p1.X, p2.Y-p1.Y
		a := Point{p1.X + dx/3, p1.Y + dy/3}
		b := Point{p1.X + 2*dx/3, p1.Y + 2*dy/3}

		theta := math.Pi / 3
		sinTheta, cosTheta := math.Sin(theta), math.Cos(theta)
		px := float64(b.X-a.X)*cosTheta - float64(b.Y-a.Y)*sinTheta + float64(a.X)
		py := float64(b.X-a.X)*sinTheta + float64(b.Y-a.Y)*cosTheta + float64(a.Y)
		c := Point{int(px), int(py)}

		// Recursively draw the four line segments.
		ppm.drawKochLine(n-1, p1, a, color)
		ppm.drawKochLine(n-1, a, c, color)
		ppm.drawKochLine(n-1, c, b, color)
		ppm.drawKochLine(n-1, b, p2, color)
	}
}
