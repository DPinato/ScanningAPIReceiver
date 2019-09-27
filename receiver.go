// possible input arguments:
// --port <n>: TCP port web server will listen on. Defaults to 8080
// --validator <str>: validator string that will be returned to the Meraki cloud
// --logfile <path>: path where log file with all the requests received is stored. Logging will be sent to STDOUT by default

package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gorilla/mux"
)

var validatorStr = ""
var logfile *os.File
var port = "8080"

func main() {
	// read input arguments and store them to variables
	args := os.Args[1:]
	argMap := map[string]string{}
	if len(args)%2 != 0 || len(args) == 0 {
		log.Fatalln("Bad number of input arguments")
	}

	for i := 0; i < len(args); i += 2 {
		argMap[args[i]] = args[i+1]
	}
	log.Println(argMap)

	validatorStr = argMap["--validator"]
	if validatorStr == "" {
		log.Fatalln("Enter validator string with --validator <string>")
	}

	var err error
	if val, ok := argMap["--logfile"]; ok {
		logfile, err = os.OpenFile(val, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	} else {
		logfile, err = os.OpenFile("receiver.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
	if err != nil {
		log.Fatalln("Failed to open log file, " + err.Error())
	}
	defer logfile.Close()

	multi := io.MultiWriter(logfile, os.Stdout)
	log.SetOutput(multi)

	if val, ok := argMap["--port"]; ok {
		port = val
	}

	// define HTTP methods to respond to and start HTTP server
	router := mux.NewRouter()
	router.HandleFunc("/", validatorHandler).Methods("GET")
	router.HandleFunc("/", dataHandler).Methods("POST")
	router.HandleFunc("/{*}", otherHandler)

	log.Fatalln(http.ListenAndServe(":"+port, router)) // if the HTTP server errors out, just exit
	log.Printf("\n\n")
}

func validatorHandler(w http.ResponseWriter, r *http.Request) {
	// dump request received
	prettyRequest, err := prettyHTTPRequestDump(r)
	if err != nil {
		log.Println(err)
	}
	log.Println(prettyRequest)

	// reply to the endpoint with the validator string
	n, err := w.Write([]byte(validatorStr))
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Sent validator string, bytes: %d", n)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	// just dump request received to STDOUT and logfile
	prettyRequest, err := prettyHTTPRequestDump(r)
	if err != nil {
		log.Println(err)
	}
	log.Println(prettyRequest)
}

func otherHandler(w http.ResponseWriter, r *http.Request) {
	// log any other http requests, this should not really happen, it will return a code 404
	prettyRequest, err := prettyHTTPRequestDump(r)
	if err != nil {
		log.Println(err)
	}
	log.Println(prettyRequest)
	http.Error(w, "WHY?", http.StatusNotFound)
}

func prettyHTTPRequestDump(r *http.Request) (string, error) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return "", err
	}
	return string(requestDump), nil
}
