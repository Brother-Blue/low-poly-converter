package internal

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"

	"github.com/fogleman/delaunay"
)

// One poly point per X pixels
const density = 500

// Minimum of 10 triangulation points
const minPoints = 10

/*
ApplyLowPoly applies a low-poly effect to the given image.
The intensity parameter controls the number of triangulation points.
*/
func ApplyLowPoly(img image.Image, intensity int) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	polyPoints := (w * h / density) * intensity / 100
	polyPoints = max(polyPoints, minPoints)

	// Generate random points
	// Add + 4 to account for the corners
	points := make([][2]float64, polyPoints+4)
	for i := 0; i < polyPoints; i++ {
		x := rand.Float64() * float64(w)
		y := rand.Float64() * float64(h)
		points = append(points, [2]float64{x, y})
	}

	// Add corners of the image
	points[polyPoints] = [2]float64{0, 0}
	points[polyPoints+1] = [2]float64{float64(w - 1), 0}
	points[polyPoints+2] = [2]float64{0, float64(h - 1)}
	points[polyPoints+3] = [2]float64{float64(w - 1), float64(h - 1)}

	// Delaunay triangulation
	delPoints := make([]delaunay.Point, len(points))
	for i, p := range points {
		delPoints[i] = delaunay.Point{X: p[0], Y: p[1]}
	}
	tris, _ := delaunay.Triangulate(delPoints)

	out := image.NewRGBA(bounds)
	draw.Draw(out, bounds, img, bounds.Min, draw.Src)

	for i := 0; i < len(tris.Triangles); i += 3 {
		ia := tris.Triangles[i]
		ib := tris.Triangles[i+1]
		ic := tris.Triangles[i+2]

		ax, ay := tris.Points[ia].X, tris.Points[ia].Y
		bx, by := tris.Points[ib].X, tris.Points[ib].Y
		cx, cy := tris.Points[ic].X, tris.Points[ic].Y

		processTriangle(img, out, ax, ay, bx, by, cx, cy)
	}

	return out
}

/*
isPointInTriangle checks if a point (px, py) is inside the triangle defined by points (ax, ay), (bx, by), and (cx, cy).
It uses the barycentric coordinates method to determine if the point is within the triangle.
*/
func isPointInTriangle(px, py, ax, ay, bx, by, cx, cy float64) bool {
	p1x, p1y := cx-ax, cy-ay
	p2x, p2y := bx-ax, by-ay
	p3x, p3y := px-ax, py-ay

	dot0 := p1x*p1x + p1y*p1y
	dot1 := p1x*p2x + p1y*p2y
	dot2 := p1x*p3x + p1y*p3y
	dot3 := p2x*p2x + p2y*p2y
	dot4 := p2x*p3x + p2y*p3y

	denominator := (dot0 * dot3) - (dot1 * dot1)
	if denominator == 0 {
		return false
	}

	inverted := 1 / denominator
	u := ((dot3 * dot2) - (dot1 * dot4)) * inverted
	v := ((dot0 * dot4) - (dot1 * dot2)) * inverted
	return (u >= 0) && (v >= 0) && (u+v <= 1)
}

/*
processTriangle processes a triangle defined by points (ax, ay), (bx, by), and (cx, cy) in the input image imgIn.
It calculates the average color of the pixels inside the triangle and fills the triangle in the output image
imgOut with that average color.
It iterates over the bounding box of the triangle to find all pixels that are inside it, computes their average color,
and sets that color for all pixels inside the triangle in the output image.
*/
func processTriangle(imgIn image.Image, imgOut *image.RGBA, ax, ay, bx, by, cx, cy float64) {
	var r, g, b, a, count uint32
	points := make([][2]float64, 0)
	minX := int(min(ax, min(bx, cx)))
	maxX := int(max(ax, max(bx, cx)))
	minY := int(min(ay, min(by, cy)))
	maxY := int(max(ay, max(by, cy)))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if isPointInTriangle(float64(x), float64(y), ax, ay, bx, by, cx, cy) {
				points = append(points, [2]float64{float64(x), float64(y)})
				cr, cg, cb, ca := imgIn.At(x, y).RGBA()
				r += cr
				g += cg
				b += cb
				a += ca
				count++
			}
		}
	}
	if count != 0 {
		col := color.RGBA{
			uint8(r / count >> 8),
			uint8(g / count >> 8),
			uint8(b / count >> 8),
			uint8(a / count >> 8),
		}
		for i := 0; i < len(points); i++ {
			x, y := int(points[i][0]), int(points[i][1])
			if x < 0 || y < 0 || x >= imgIn.Bounds().Dx() || y >= imgIn.Bounds().Dy() {
				continue
			}
			imgOut.Set(x, y, col)
		}
	}
}
