package pdf

import (
    "context"
    "net/http"
		"log"
    
    "github.com/starwalkn/gotenberg-go-client/v8"
    "github.com/starwalkn/gotenberg-go-client/v8/document"
)

func Generate(req *http.Request, pdfName string, bodyBytes []byte)(error){
    client, err := gotenberg.NewClient("gotenberg:3000", http.DefaultClient)
		if err != nil {
			log.Fatalf("Couln't connect to gotenberg container. \nErr: %q \n", err)
		}

    // Creates the Gotenberg documents from a files paths.
		doc, err := document.FromBytes(pdfName, bodyBytes)

    // Create the HTML request.
    req := gotenberg.NewHTMLRequest(doc)

    // Loading style and image from the specified urls. 
    downloads := make(map[string]map[string]string)
    downloads["http://my.style.css"] = nil
    downloads["http://my.img.gif"] = map[string]string{"X-Header": "Foo"}

    req.DownloadFrom(downloads)


    // Set the document parameters to request (optional).
    req.Margins(gotenberg.NoMargins)
    req.Scale(0.90)
    req.PaperSize(gotenberg.A4)

    // Skips the IDLE events for faster PDF conversion.
    req.SkipNetworkIdleEvent(true)

    // If you wish to redirect the response directly to the browser, you may also use:
    resp, err := client.Send(context.Background(), req)
}
