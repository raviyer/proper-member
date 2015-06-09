package comms

import (
	"bytes"
	"config"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type client struct {
	ip   string
	port int
}

func (this *client) uri() string {
	return fmt.Sprintf("http://%v:%v", this.ip, this.port)
}

func (this *client) http_method(resource string, args map[string]interface{},
	value config.VersionedConfig, http_function string) (err error) {
	u := fmt.Sprintf("%v/%v", this.uri(), resource)
	var resp *http.Response
	switch {
	case http_function == "GET":
		if len(args) != 0 {
			u += "/?"
		}
		for k, v := range args {
			u += fmt.Sprintf("%v=%v&", k, v)
		}
		resp, err = http.Get(u)
	case http_function == "POST":
		if len(args) != 0 {
			values := url.Values{}
			for k, v := range args {
				values[k][0] = fmt.Sprintf("%v", v)
			}
			resp, err = http.PostForm(u, values)
		} else {
			buf := bytes.Buffer{}
			config.EncodeObject(&buf, value)
			resp, err = http.Post(u, "", &buf)
		}
	}
	if err != nil {
		log.Print(err)
		return
	}

	if http_function == "GET" {
		defer resp.Body.Close()
		err = config.DecodeObject(resp.Body, value)
		if err != nil {
			log.Print(err)
		}
	}
	return
}

func (this *client) Get(resource string, args map[string]interface{},
	value config.VersionedConfig) (err error) {
	return this.http_method(resource, args, value, "GET")
}
func (this *client) Set(resource string, args map[string]interface{},
	value config.VersionedConfig) (err error) {
	return this.http_method(resource, args, value, "POST")
}

func NewClient(ip string, port int) (c *client) {
	c = new(client)
	c.ip = ip
	c.port = port
	return
}
