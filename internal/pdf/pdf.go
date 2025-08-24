package pdf

import (
    "context"
    "net/http"
		"log"
    
    "github.com/starwalkn/gotenberg-go-client/v8"
    "github.com/starwalkn/gotenberg-go-client/v8/document"
)

func Generate(r *http.Request, pdfName string, bodyBytes []byte)(*http.Response, error){
	client, err := gotenberg.NewClient("http://gotenberg:3000", http.DefaultClient)
		if err != nil {
			log.Fatalf("Couln't connect to gotenberg container. \nErr: %q \n", err)
		}
		doc, err := document.FromBytes(pdfName, bodyBytes)

    // Create the HTML request.
    req := gotenberg.NewHTMLRequest(doc)

    // Set the document parameters to request (optional).
    req.Margins(gotenberg.NoMargins)
    req.Scale(0.90)
    req.PaperSize(gotenberg.A4)

    // Skips the IDLE events for faster PDF conversion.
    req.SkipNetworkIdleEvent(true)
    // If you wish to redirect the response directly to the browser, you may also use:
    resp, err := client.Send(context.Background(), req)
		if err != nil{
			log.Fatalln(err)
		}
		return resp, nil
}
