package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

var httpAddr = flag.String("http", ":8080", "Listen address")
var enablePush = flag.Bool("enablePush", true, "Enable HTTP/2 push")

var indexHTMLTmpl = template.Must(template.New("index").Parse(`<html>
<head>
	<title>Pushing {{.}} files to the browser</title>
</head>
<body>
<p>Total: <span id="total">No results yet</span></p>
<script>
const totalElt = document.getElementById("total");
let total = 0;
for (let i = 0; i < {{.}}; i++) {
	fetch('/file?n=' + i)
		.then(res => res.text())
		.then(body => {
			total += parseInt(body, 10)
			totalElt.innerText = total.toString(10)
		})
}
</script>
</body>
</html>
`))

func main() {
	flag.Parse()

	http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/file" {
			http.NotFound(w, r)
			return
		}

		nFiles := r.URL.Query().Get("n")
		if nFiles == "" {
			http.NotFound(w, r)
			return
		}

		r.Header.Add("Content-Type", "text")
		fmt.Fprintf(w, nFiles)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		nFiles := r.URL.Query().Get("n")
		if nFiles == "" {
			http.NotFound(w, r)
			return
		}

		n, err := strconv.Atoi(nFiles)
		if err != nil {
			log.Printf("invalid nFiles: %s", nFiles)
			http.NotFound(w, r)
			return
		}

		pusher, ok := w.(http.Pusher)
		if ok && *enablePush {
			for i := 0; i < n; i++ {
				assetPath := fmt.Sprintf("/file?n=%d", i)
				err := pusher.Push(assetPath, nil)
				if err != nil {
					log.Printf("Failed to push: %v", err)
				} else {
					log.Printf("Succesfully pushed %s", assetPath)
				}
			}
		}
		indexHTMLTmpl.Execute(w, n)
	})

	log.Printf("listening on https://localhost%s", *httpAddr)
	log.Fatal(http.ListenAndServeTLS(*httpAddr, "localhost.crt", "localhost.key", nil))
}
