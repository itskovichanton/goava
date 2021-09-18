package httputils

import (
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func AddFile(field string, filename string, writer *multipart.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	return nil
}

func GetUrl(r *http.Request) string {
	return r.Host + r.RequestURI
}

func ListenAndServe(srv *http.Server, tls bool, protocol, network, certFile, keyFile string) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":" + protocol
	}

	ln, err := net.Listen(network, addr)
	if err != nil {
		return err
	}

	defer ln.Close()
	if tls {
		return srv.ServeTLS(ln, certFile, keyFile)
	}
	return srv.Serve(ln)
}

func BasicAuth(r *http.Request, username, password string) {
	r.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
}

func ToValues(p map[string]string) url.Values {
	r := url.Values{}
	for k, v := range p {
		r.Add(k, fmt.Sprintf("%v", v))
	}
	return r
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func GETPOST(e *echo.Echo, path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	e.Add(http.MethodGet, path, h, m...)
	return e.Add(http.MethodPost, path, h, m...)
}

func NewHttpRequest(method string, callUrl string, params map[string]interface{}) (*http.Request, error) {
	p := url.Values{}
	for k, v := range params {
		p.Add(k, cast.ToString(v))
	}

	if strings.EqualFold(method, "POST") {
		r, err := http.NewRequest(method, callUrl, strings.NewReader(p.Encode()))
		if err != nil {
			return nil, err
		}
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		return r, nil
	}

	u, err := url.Parse(callUrl)
	if err != nil {
		return nil, err
	}

	u.RawQuery = p.Encode()

	r, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}
