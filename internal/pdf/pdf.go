package pdf

import (
	"log"
	"strings"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

// Writes html into a pdf using wkhtmltopdf.
// Note that, the binary (plus dependencies) of wkhtmltopdf 
// must be present in the system path
func Generate(pdfName string, bodyBytes []byte)([]byte, error){
		body := strings.NewReader(string(bodyBytes))
		pdfg, err :=  wkhtml.NewPDFGenerator()
		if err != nil{
			log.Fatalf("problem with pdf generator: %q", err)
			return nil, err
		}
		// Add options for wkhtmltopdf
    pdfg.PageSize.Set(wkhtml.PageSizeLetter)
    page := wkhtml.NewPageReader(body)
    page.Zoom.Set(0.95)
		pdfg.AddPage(page)
		if err = pdfg.Create(); err != nil {
			return nil, err
		}
		return pdfg.Bytes(), nil
}
