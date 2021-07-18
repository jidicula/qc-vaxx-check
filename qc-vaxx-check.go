package main

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	filepath := flag.Arg(0)
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(2)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	bmp, _ := gozxing.NewBinaryBitmapFromImage(img)

	qrReader := qrcode.NewQRCodeReader()
	result, _ := qrReader.Decode(bmp, nil)
	qrBody := result.String()

	// convert numeric QR into byte slice
	shcJWS := []byte{}
	for i := 5; i < len(qrBody); i += 2 {
		pair := qrBody[i : i+2]
		j, err := strconv.Atoi(pair)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		j += 45
		char := byte(j)
		shcJWS = append(shcJWS, char)
	}

	// Decode portions of JWS
	jwsParts := [][]byte{}
	for _, s := range strings.Split(string(shcJWS), ".") {
		portion, err := decode([]byte(s))
		if err != nil {
			fmt.Fprintln(os.Stderr, s)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(20)
		}
		jwsParts = append(jwsParts, portion)
	}
	fmt.Printf("%s\n", jwsParts)

	// read SHC contents
	b := bytes.NewReader(jwsParts[1])
	z := flate.NewReader(b)
	defer z.Close()
	shc, err := ioutil.ReadAll(z)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(40)
	}
	fmt.Printf("%s\n", shc)
	shc = PatientFamilyNameWorkaround(shc)
	fmt.Printf("%s\n", shc)
	var v VerificationBody
	if err := json.Unmarshal(shc, &v); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Printf("%v\n", v)

	// TODO: take filepath arg to QR code image
	// TODO: Decode QR
	// TODO: Check signature
	// TODO: Get dose count
	// TODO: Output verification validity
	// TODO: Output dose count
}

// decode decodes a base64 byte slice.
func decode(data []byte) ([]byte, error) {
	// Pad to length that's a multiple of 4.
	if missingPadding := len(data) % 4; missingPadding > 0 {
		for i := 0; i < 4-(missingPadding); i++ {
			data = append(data, '=')
		}
	}
	return base64.URLEncoding.DecodeString(string(data))
}

type VerificationBody struct {
	Iss string `json:"iss"`
	Nbf int    `json:"nbf"`
	Vc  struct {
		Type              []string `json:"type"`
		CredentialSubject struct {
			FhirVersion string `json:"fhirVersion"`
			FhirBundle  struct {
				ResourceType string `json:"resourceType"`
				Type         string `json:"type"`
				Entry        []struct {
					FullURL  string `json:"fullUrl"`
					Resource struct {
						ResourceType string `json:"resourceType"`
						Name         []struct {
							Family string   `json:"family"`
							Given  []string `json:"given"`
						} `json:"name"`
						BirthDate string `json:"birthDate"`
					} `json:"resource"`
				} `json:"entry"`
			} `json:"fhirBundle"`
		} `json:"credentialSubject"`
	} `json:"vc"`
}

// PatientFamilyNameWorkaround replaces the string array for name.family in older versions of the FHIR Patient payload.
func PatientFamilyNameWorkaround(data []byte) []byte {
	var re = regexp.MustCompile(`(?m)"family":\[.*\],`)
	familyNameStringArray := re.Find(data)
	if string(familyNameStringArray) == "" {
		return data
	}
	// Handle empty family name scenario
	var reEmptyName = regexp.MustCompile(`(?m)"family":\[\],`)
	data = reEmptyName.ReplaceAll(data, []byte(`"family":"",`))

	// join quotes and commas for multiple family names
	var reQuotesComma = regexp.MustCompile(`","`)
	familyNameString := reQuotesComma.ReplaceAll(familyNameStringArray, []byte(" "))

	// remove brackets
	familyNameString = bytes.Replace(familyNameString, []byte("["), []byte(""), 1)
	familyNameString = bytes.Replace(familyNameString, []byte("]"), []byte(""), 1)

	cleanedData := bytes.Replace(data, familyNameStringArray, familyNameString, 1)
	return cleanedData
}
