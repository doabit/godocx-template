package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "embed"

	. "github.com/ArFnds/godocx-template"
	"github.com/skip2/go-qrcode"
)

//go:embed dataset.json
var datasetFile []byte

func main() {
	//slog.SetLogLoggerLevel(slog.LevelDebug)
	buffer := bytes.NewBuffer(datasetFile)
	jsonDecoder := json.NewDecoder(buffer)
	jsonData := ReportData{}
	jsonDecoder.Decode(&jsonData)
	outBytes, err := CreateReport(os.Args[1],
		&jsonData,
		CreateReportOptions{
			LiteralXmlDelimiter: "||",
			// Otherwise unused but mandatory options
			FixSmartQuotes:    true,
			ProcessLineBreaks: true,
			Functions: Functions{
				"tile": func(args ...any) VarValue {
					z := args[0].(int64)
					y := args[1].(int64)
					x := args[2].(int64)
					url := fmt.Sprintf("https://tile.thunderforest.com/cycle/%d/%d/%d.png", z, x, y)

					resp, err := http.DefaultClient.Get(url)
					if err != nil {
						panic(err)
					}
					defer resp.Body.Close()
					img, _ := io.ReadAll(resp.Body)
					return &ImagePars{
						Data:      img,
						Width:     3,
						Height:    3,
						Extension: ".png",
					}
				},
				"qr": func(args ...any) VarValue {
					if url, ok := args[0].(string); ok {
						png, err := qrcode.Encode(url, qrcode.Medium, 256)
						if err != nil {
							return "Err"
						}
						return &ImagePars{
							Data:      png,
							Extension: ".png",
							Width:     6,
							Height:    6,
						}
					}
					return ""
				},
			},
		})
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(os.Args[2], outBytes, 0666)
	if err != nil {
		panic(err)
	}

}
