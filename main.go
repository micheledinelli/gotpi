package main

import (
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"os"

	"github.com/akamensky/argparse"
	"github.com/kevin-cantwell/dotmatrix"
	"github.com/nfnt/resize"
)

func main() {
	cli := argparse.NewParser("One-Time Pad image encryptor", "encrypt and decrypt images using OTP images as keys")
	keyGen := cli.NewCommand("key-gen", "Generate a new OTP key image")
	keyF := keyGen.String("o", "out", &argparse.Options{Required: false, Help: "Path to store the generated otp key", Default: "otp-key.png"})
	keyW := keyGen.Int("w", "width", &argparse.Options{Required: false, Help: "Width (same as height) of the generated otp key image", Default: 256})

	enc := cli.NewCommand("enc", "Encrypt an image using an OTP key image")
	imgF := enc.String("f", "file", &argparse.Options{Required: true, Help: "Path of the image to encrypt"})
	key := enc.String("k", "key", &argparse.Options{Required: true, Help: "Path of the key image to use for encryption"})
	outEnc := enc.String("o", "out", &argparse.Options{Required: false, Help: "Path to save the encrypted image", Default: "enc.png"})

	dec := cli.NewCommand("dec", "Decrypt an image using an OTP key image")
	decImgF := dec.String("f", "file", &argparse.Options{Required: true, Help: "Path of the image to decrypt"})
	decKey := dec.String("k", "key", &argparse.Options{Required: true, Help: "Path of the key image to use for decryption"})
	outDec := dec.String("o", "out", &argparse.Options{Required: false, Help: "Path to save the decrypted image", Default: "dec.png"})

	xor := cli.NewCommand("xor", "XOR two images together")
	xorImg1 := xor.String("a", "img1", &argparse.Options{Required: true, Help: "Path of the first image"})
	xorImg2 := xor.String("b", "img2", &argparse.Options{Required: true, Help: "Path of the second image"})
	outXor := xor.String("o", "out", &argparse.Options{Required: false, Help: "Path to save the XORed image", Default: "xor.png"})

	verbose := cli.Flag("v", "verbose", &argparse.Options{Required: false, Help: "Print the encrypted image to terminal", Default: false})
	rgb := cli.Flag("c", "rgb", &argparse.Options{Required: false, Help: "use RGB mode instead of black and white", Default: false})

	err := cli.Parse(os.Args)
	if err != nil {
		fmt.Print(cli.Usage(err))
		return
	}

	if keyGen.Happened() {
		k := KeyGen(keyF, keyW, *rgb)
		save(*keyF, k)
		if *verbose {
			termPrint(k)
			fmt.Printf("otp key written to %s\n", *keyF)
		}
	}

	if enc.Happened() {
		img := imgOpen(*imgF)
		keyImg := imgOpen(*key)

		out := Encrypt(img, keyImg, *rgb)
		save(*outEnc, out)

		if verbose != nil && *verbose {
			fmt.Printf("encrypting %s", *imgF)
			termPrint(img)
			fmt.Printf("with key %s", *key)
			termPrint(keyImg)
			fmt.Printf("file saved to %s", *outEnc)
			termPrint(out)
		}
	}

	if dec.Happened() {
		decImg := imgOpen(*decImgF)
		decKeyImg := imgOpen(*decKey)

		out := Decrypt(decImg, decKeyImg, *rgb)
		save(*outDec, out)

		if verbose != nil && *verbose {
			fmt.Printf("decrypting %s", *decImgF)
			termPrint(decImg)
			fmt.Printf("with key %s", *decKey)
			termPrint(decKeyImg)
			fmt.Printf("file saved to %s", *outDec)
			termPrint(out)
		}
	}

	if xor.Happened() {
		a := imgOpen(*xorImg1)
		b := imgOpen(*xorImg2)

		out := Encrypt(a, b, *rgb)
		save(*outXor, out)

		if verbose != nil && *verbose {
			fmt.Printf("XORing %s with %s", *xorImg1, *xorImg2)
			termPrint(a)
			termPrint(b)
			fmt.Printf("file saved to %s", *outXor)
			termPrint(out)
		}
	}
}

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

// KeyGen generates a new OTP key image. If the rgb flag is true,
// it generates an RGB key image otherwise, it generates a black
// and white key image. The key image is always of size kw x kw.
// A rgb otp key can encrypt coloured images, while a black and white
// otp key can only encrypt black and white images. Attempting to ecnrtypt
// a coloured image with a black and white key will result in loss of
// colour information and unsuccessful decryption.
func KeyGen(keyFile *string, kw *int, rgb bool) image.Image {
	var k image.Image
	if rgb {
		k = keyGenRGB(*kw, *kw)
	} else {
		k = keyGenBW(*kw, *kw)
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

func termPrint(img image.Image) error {
	fmt.Printf("\n")
	return dotmatrix.Print(os.Stdout, resize.Resize(128, 0, img, resize.Lanczos3))
}

func save(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func imgOpen(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}
