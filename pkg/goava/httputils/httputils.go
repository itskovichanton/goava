package httputils

import (
	"encoding/base64"
	"fmt"
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
	r.Header.Add("Authorization", CalcBasicAuth(username, password))
}

func CalcBasicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func ParseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
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
