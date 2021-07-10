package main

import "fmt"

var usage = `Usage: qc-vaxx-check [options...] <qr_code file>

qc-vaxx-check is a tool for checking if a vaccination status QR code is valid.

Example:

1. Save the QR code as an image file.

2. Run the command on the QR code image file:
    $ qc-vaxx-check qr_code.png
    This QR code has been signed by the Quebec Ministry of Health and Social Services.
    Jean Biche received 2 doses of the COVID-19 vaccine.
`

func main() {
	fmt.Println("Hello, world!")
	// TODO: take filepath arg to QR code image
	// TODO: Decode QR
	// TODO: Check signature
	// TODO: Get dose count
	// TODO: Output verification validity
	// TODO: Output dose count
}
