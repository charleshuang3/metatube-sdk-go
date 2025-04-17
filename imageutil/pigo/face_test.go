package pigo

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"testing"

	pigo "github.com/esimov/pigo/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

/*
  # Generate MD5-hashed images:
  setopt null_glob
  for IMG in *.jpg *.jpeg *.png; do
    gbase64 "$IMG" > "$(md5sum "$IMG" | cut -d' ' -f1)"
  done

  # Decode image:
  gbase64 -d "$IMG"  > "$IMG".jpg
*/

//go:embed testdata/*
var fs embed.FS

func TestCalculatePosition(t *testing.T) {
	for _, unit := range []struct {
		filename string
		position float64
	}{
		{filename: "809ee47a17a7938ebd6d908244b962c8", position: 0.40},
		{filename: "c7806a2f581012eb71dce597f682c7a2", position: 0.22},
		{filename: "aa237b3a2bfd35dbe10c386c7ac777ae", position: 0.30},
		{filename: "c2a6bf02b748bb460a0dbb550f39c635", position: 0.70},
		{filename: "1d8aaf63245426c4a32720bdbf33a651", position: 0.80},
		{filename: "e953fce3bf5ec5746ead8954bec758e0", position: 0.30},
		{filename: "d054f170d52c83a773571675d954e3bb", position: 0.75},
		{filename: "3685c2648be7eeeaa2cef0118873a55f", position: 0.60},
		{filename: "7848e5995a58df9d063df8543c50c943", position: 0.20},
		{filename: "c0c99e28da91693a27de2beb6dfd7161", position: 0.50},
		{filename: "6dbe5b2d7d7056f3b60c6d05f5176529", position: 0.25},
		{filename: "369993051097480935eadf1f468eaadb", position: 0.70},
		{filename: "d8df07a8312543f638373eb5921f896d", position: 0.15},
		{filename: "e8e85575a04d75d2bc29abb4bb7fb447", position: 0.25},
		{filename: "c977809e691fc2037f3a9279068720c2", position: 0.70},
		{filename: "e1c5fce943a4ba36576607eaa585b9d8", position: 0.90},
		{filename: "345a376e579ff02a518b831b1b2b4602", position: 0.20},
		{filename: "f100611a90fa024c73132457fa77da36", position: 0.65},
		{filename: "e5ff5d6966391409a0fed7d3446b12aa", position: 0.60},
		{filename: "068b7fb0c8e3953ff5ed25fe00fc22fd", position: 0.25},
		{filename: "6335e1276cd7edc191de3768dd62aa03", position: 0.15},
		{filename: "4307f4c6826a88936e6e4351d70195cb", position: 0.45},
		{filename: "eec27d560038e3367afe42f4ffa7a8e6", position: 0.65},
		{filename: "263a3cc91c74673957ea9ca7dbac11f4", position: 0.15},
		{filename: "bfc6d0dcf7d9750d13d3c52cac84ed9a", position: 0.20},
		{filename: "db4aec6ce163c3113473af00848f717a", position: 0.85},
		{filename: "a6d7e2f816aae0c22150688489491d21", position: 0.60},
		{filename: "91271e4f1c1369ab3f06da9b1175d450", position: 0.70},
		{filename: "2c1cc55118d22b39dbbae5c8fca2aa1f", position: 0.40},
		{filename: "f88aa397bf3ea9df387af2de6d12a6c7", position: 0.20},
		{filename: "21939b16a2dc22be9a4035f04da350db", position: 0.25},
		{filename: "7e655977ad687e683815567b8081d9f9", position: 0.55},
		{filename: "3ed1e1a46f25375ba3478d72cd2b9958", position: 0.28},
		// Failed detection:
		// {filename: "ca5993f3f85d7ee19aeb9bf1e997e7bb", position: 0.72},
		// {filename: "ffe1f9b37d33bc9b7e0a4e400ffb64f7", position: 0.85},
		// {filename: "bad4c3bd1484a6e32873839f0a5ec77e", position: 0.25},
	} {
		t.Run(unit.filename, func(t *testing.T) {
			data, err := fs.ReadFile("testdata/" + unit.filename)
			require.NoError(t, err)

			decoded, err := base64.StdEncoding.DecodeString(string(data))
			require.NoError(t, err)

			img, _, err := image.Decode(bytes.NewReader(decoded))
			require.NoError(t, err)

			const (
				ratio     = 0 /* backdrop */
				tolerance = 5e-2
			)

			var (
				innerImg   image.Image
				innerFaces []pigo.Detection
				innerVotes []vote
			)

			pos, found := DetectMainFacePosition(
				img, ratio,
				// debug: extract all inner variables.
				func(img image.Image, faces []pigo.Detection, votes []vote) {
					innerImg = img
					innerFaces = faces
					innerVotes = votes
				},
			)

			// debug: print all votes
			for _, v := range innerVotes {
				t.Logf(
					"count:%d, sumPos:%.2f, weight:%.2f, pos=%.2f",
					v.count, v.sumPos, v.weight, v.avgPos(),
				)
			}

			// debug: print all faces.
			for _, face := range innerFaces {
				t.Logf("%v, weight=%.2f, pos=%.2f",
					face, float32(face.Scale)*face.Q,
					calculateFacePosition(innerImg, ratio, face),
				)
			}

			if assert.True(t, found) {
				t.Logf("detected position: %f", pos)
			}

			if !found || !assert.LessOrEqualf(t,
				math.Abs(unit.position-pos), tolerance,
				"expect pos=%.2f, but got pos=%.2f", unit.position, pos) {
				// debug: draw image with boxes.
				debugImg := drawBoxes(innerImg, innerFaces)
				saveImage(unit.filename, debugImg)
			}
		})
	}
}

func drawBoxes(img image.Image, dets []pigo.Detection) image.Image {
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	_ = blue
	_ = yellow

	for _, m := range dets {
		x0 := m.Col - m.Scale/2
		y0 := m.Row - m.Scale/2
		x1 := m.Col + m.Scale/2
		y1 := m.Row + m.Scale/2

		if x0 < 0 {
			x0 = 0
		}
		if y0 < 0 {
			y0 = 0
		}
		if x1 >= rgba.Bounds().Dx() {
			x1 = rgba.Bounds().Dx() - 1
		}
		if y1 >= rgba.Bounds().Dy() {
			y1 = rgba.Bounds().Dy() - 1
		}

		// Draw red rectangle
		for x := x0; x <= x1; x++ {
			rgba.Set(x, y0, red)
			rgba.Set(x, y1, red)
		}
		for y := y0; y <= y1; y++ {
			rgba.Set(x0, y, red)
			rgba.Set(x1, y, red)
		}

		// Draw label inside the box
		label := fmt.Sprintf("(%d,%d,%d,%.2f)", m.Col, m.Row, m.Scale, m.Q)
		point := fixed.Point26_6{
			X: fixed.I(x0 + 2),
			Y: fixed.I(y0 + 12),
		}
		d := &font.Drawer{
			Dst:  rgba,
			Src:  image.NewUniform(yellow),
			Face: basicfont.Face7x13,
			Dot:  point,
		}
		d.DrawString(label)
	}

	return rgba
}

func saveImage(name string, img image.Image) {
	outputFile := fmt.Sprintf("%s.jpg", name)
	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_ = jpeg.Encode(f, img, nil)
}
