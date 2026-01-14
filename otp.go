package gotpi

import (
	"crypto/rand"
	"image"
	"image/color"
	_ "image/png"

	"github.com/nfnt/resize"
)

// Encrypt encrypts an image using a key image. If the rgb flag is true,
// it uses RGB encryption otherwise, it uses black and white encryption.
// The encrypted image is returned.
func Encrypt(img image.Image, keyImg image.Image, rgb bool) image.Image {
	bounds := keyImg.Bounds()
	resizedImg := resize.Resize(uint(bounds.Dx()), uint(bounds.Dy()), img, resize.Lanczos3)
	out := image.NewRGBA(keyImg.Bounds())
	if rgb {
		encRGB(resizedImg, keyImg, out)
	} else {
		encBW(resizedImg, keyImg, out)
	}
	return out
}

// Decrypt decrypts an image using a key image. If the rgb flag is true,
// it uses RGB decryption otherwise, it uses black and white decryption.
// The decrypted image is returned.
func Decrypt(img image.Image, keyImg image.Image, rgb bool) image.Image {
	return Encrypt(img, keyImg, rgb)
}

func encBW(img, k image.Image, out *image.RGBA) {
	bounds := k.Bounds()
	monoResized := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := MonochromeModel.Convert(img.At(x, y))
			monoResized.Set(x, y, p)
		}
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			origPixel := MonochromeModel.Convert(monoResized.At(x, y)).(Pixel)
			keyPixel := MonochromeModel.Convert(k.At(x, y)).(Pixel)

			result := White
			if origPixel != keyPixel {
				result = Black
			}
			out.Set(x, y, result)
		}
	}
}

func encRGB(img, k image.Image, out *image.RGBA) {
	bounds := k.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			kr, kg, kb, _ := k.At(x, y).RGBA()

			resR := uint8(r>>8) ^ uint8(kr>>8)
			resG := uint8(g>>8) ^ uint8(kg>>8)
			resB := uint8(b>>8) ^ uint8(kb>>8)

			out.Set(x, y, color.RGBA{resR, resG, resB, 255})
		}
	}
}

// KeyGen generates a new OTP key image.
// It accepts the keyFile path to save the key image to, the kw size of
// the key image, and a rgb flag. If the rgb flag is true, it generates
// an RGB key image otherwise, it generates a black and white key image.
// The key image is always of size kw x kw.
// A rgb otp key can encrypt coloured images, while a black and white
// otp key can only encrypt black and white images. Attempting to ecnrtypt
// a coloured image with a black and white key will result in loss of
// colour information and unsuccessful decryption.
func KeyGen(keyFile string, kw int, rgb bool) image.Image {
	var k image.Image
	if rgb {
		k = keyGenRGB(kw, kw)
	} else {
		k = keyGenBW(kw, kw)
	}
	return k
}

func keyGenBW(width, height int) image.Image {
	k := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			var b [1]byte
			rand.Read(b[:])
			if b[0]&1 == 0 {
				k.Set(x, y, Black)
			} else {
				k.Set(x, y, White)
			}
		}
	}
	return k
}

func keyGenRGB(width, height int) image.Image {
	k := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			pix := make([]byte, 3)
			rand.Read(pix)
			k.Set(x, y, color.RGBA{pix[0], pix[1], pix[2], 255})
		}
	}
	return k
}

type Pixel bool

const (
	Black Pixel = true
	White Pixel = false
)

// https://github.com/ev3go/ev3dev/blob/a5fda5c6a492269e01b184046ed42dc4a1dfe8c9/fb/mono.go#L104
var MonochromeModel color.Model = color.ModelFunc(monoModel)

func (c Pixel) RGBA() (r, g, b, a uint32) {
	if c == Black {
		return 0, 0, 0, 0xffff
	}
	return 0xffff, 0xffff, 0xffff, 0xffff
}

func monoModel(c color.Color) color.Color {
	if _, ok := c.(Pixel); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	y := (299*r + 587*g + 114*b + 500) / 1000
	return Pixel(uint16(y) < 0x8000)
}
