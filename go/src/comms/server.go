package comms

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

type Request interface {
	GetParameter(name string) string
	GetBody() io.ReadCloser
	Sender() string
	GetMethod() string
}

type http_request struct {
	request *http.Request
}

func (r http_request) GetParameter(name string) string {
	return r.request.FormValue(name)
}
func (r http_request) GetBody() io.ReadCloser {
	return r.request.Body
}
func (r http_request) Sender() string {
	return r.request.Host
}
func (r http_request) GetMethod() string {
	return r.request.Method
}

type HandlerFunction func(io.Writer, Request) (err error)

type Server interface {
	Start() (err error)
	RegisterHandler(resource string, handler HandlerFunction)
}

type http_server struct {
	address string
	port    int
}

func (this *http_server) Start() (err error) {
	url := fmt.Sprintf("%v:%v", this.address, this.port)
	log.Printf("Serving on %v\n", url)
	err = http.ListenAndServe(url, nil)
	return err
}

type httpHandlerFunc func(http.ResponseWriter, *http.Request)

func wrapHandlerFunc(resource string, handler HandlerFunction) (hfn httpHandlerFunc) {
	hfn = func(resw http.ResponseWriter, req *http.Request) {
		resw.Header().Set("Access-Control-Allow-Origin", "*")		
		resw.Header().Set("Access-Control-Allow-Headers", "origin, content-type")
		err := handler(resw, http_request{req})
		if err != nil {
			log.Panicf("Error in %s: %+v", resource, err)
		}
	}
	return
}

func (*http_server) RegisterHandler(resource string, handler HandlerFunction) {
	http.HandleFunc(fmt.Sprintf("/%v", resource), wrapHandlerFunc(resource, handler))
}

func serveFiles(resw http.ResponseWriter, req *http.Request) {
	resw.Header().Set("Access-Control-Allow-Origin", "*")		
	resw.Header().Set("Access-Control-Allow-Headers", "origin, content-type")
	hndlr := http.StripPrefix("/ui", http.FileServer(http.Dir("../html")))
	log.Printf("Received request for %v", req.URL)
	hndlr.ServeHTTP(resw, req)
}

func NewServer(address string, port int) (srv *http_server) {
	srv = &http_server{address, port}
	// Hack, listen for files on "/ui"
	http.HandleFunc("/ui/", serveFiles)
	return
}
