package main

import (
	"fmt"
	"image"
	"image/png"
	_ "image/png"
	"os"

	"github.com/akamensky/argparse"
	"github.com/kevin-cantwell/dotmatrix"
	"github.com/micheledinelli/gotpi"
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
		k := gotpi.KeyGen(keyF, keyW, *rgb)
		save(*keyF, k)
		if *verbose {
			termPrint(k)
			fmt.Printf("otp key written to %s\n", *keyF)
		}
	}

	if enc.Happened() {
		img := imgOpen(*imgF)
		keyImg := imgOpen(*key)

		out := gotpi.Encrypt(img, keyImg, *rgb)
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

		out := gotpi.Decrypt(decImg, decKeyImg, *rgb)
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

		out := gotpi.Encrypt(a, b, *rgb)
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
