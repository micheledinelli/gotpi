package gotpi

import (
	"crypto/rand"
	"image"
	"image/color"
	_ "image/png"

	"github.com/nfnt/resize"
)

// Encrypt encrypts an input image using a key image.
// The input image is first resized to match the key image dimensions.
// If rgb is true, encryption is performed per RGB channel using XOR.
// If rgb is false, the image is converted to monochrome and encrypted
// using a black-and-white XOR-style operation.
// The resulting encrypted image is returned.
func Encrypt(img image.Image, keyImg image.Image, rgb bool) image.Image {
	bounds := keyImg.Bounds()
	out := image.NewRGBA(keyImg.Bounds())
	img = resize.Resize(uint(bounds.Dx()), uint(bounds.Dy()), img, resize.Lanczos3)
	if rgb {
		encRGB(img, keyImg, out)
	} else {
		encBW(img, keyImg, out)
	}
	return out
}

// Decrypt decrypts an input image using a key image.
// The input image is first resized to match the key image dimensions.
// If rgb is true, decryption is performed per RGB channel using XOR.
// If rgb is false, the image is converted to monochrome and decrypted
// using a black-and-white XOR-style operation.
// The resulting decrypted image is returned.
func Decrypt(img image.Image, keyImg image.Image, rgb bool) image.Image {
	return Encrypt(img, keyImg, rgb)
}

// encBW encrypts an image using black-and-white (monochrome) encryption.
// Both the source image and key image are converted to monochrome.
// Each output pixel is white if the source and key pixels are equal,
// and black otherwise.
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

// encRGB encrypts an image using RGB channel-wise XOR encryption.
// Each color channel (R, G, B) of the source image is XORed with the
// corresponding channel of the key image.
// The alpha channel is set to fully opaque
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

// KeyGen generates a new one-time pad (OTP) key image.
// The key image is always square with dimensions (kw × kw) which
// stands for key width.
// If rgb is true, a full-color RGB key is generated.
// If rgb is false, a black-and-white (monochrome) key is generated.
//
// An RGB key can encrypt both colored and grayscale images.
// A black-and-white key should only be used with monochrome images;
// using it to encrypt a colored image will discard color information
// and prevent successful decryption.
func KeyGen(kw int, rgb bool) image.Image {
	var k image.Image
	if rgb {
		k = keyGenRGB(kw, kw)
	} else {
		k = keyGenBW(kw, kw)
	}
	return k
}

// keyGenBW generates a black-and-white OTP key image with the given
// width and height. Each pixel is randomly assigned either black or white
// using a cryptographically secure random source.
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

// keyGenRGB generates an RGB OTP key image with the given width and height.
// Each pixel is assigned a random 24-bit RGB color using a cryptographically
// secure random source. The alpha channel is always set to fully opaque.
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

// The MonochromeModel converts colors to black or white based on their luminance.
// Colors with luminance below the midpoint are converted to black,
// while those above are converted to white.
// More at https://github.com/ev3go/ev3dev/blob/a5fda5c6a492269e01b184046ed42dc4a1dfe8c9/fb/mono.go#L104
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
