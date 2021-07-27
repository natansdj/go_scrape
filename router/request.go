package router

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/logx"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//ResponseClose Close response but check errors, used for defer statement
func ResponseClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		logx.LogError.Fatal(err)
	}
}

//args[0] : qs url.Values
func RequestInit(cfg config.ConfYaml, method string, endpoint string, body io.Reader, args ...interface{}) (req *http.Request, err error) {
	urlStr := cfg.Source.BaseURI + endpoint

	req, err = http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	//Add shipper API_KEY
	q := req.URL.Query()

	//Add queryString
	if len(args) >= 1 {
		if v, ok := args[0].(url.Values); ok && len(v) > 0 {
			for i := range v {
				q.Add(i, v.Get(i))
			}
		}
	}

	req.URL.RawQuery = q.Encode()

	//Add Header
	req.Header.Add("User-Agent", "Shipper/")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

//args[0] methodName string
func RequestDo(req *http.Request, args ...interface{}) (body []byte, err error) {
	if req == nil {
		return body, errors.New("empty request")
	}

	//Args
	var methodName string
	if len(args) >= 1 {
		if v, ok := args[0].(string); ok {
			methodName = v
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := config.ScrNetClient.Do(req.WithContext(ctx))
	if err != nil {
		logx.LogError.Error(methodName, err)
		return nil, err
	}

	defer ResponseClose(res.Body)

	body, err = ioutil.ReadAll(res.Body)

	//DEBUG
	urlStr := ""
	if req.URL != nil {
		urlStr = req.URL.String()
	}
	logx.LogAccess.Info(fmt.Sprintf("\n URL : %v \n RESP : %v", urlStr, res.Status))

	return body, err
}

type Response struct {
	Headers map[string][]string
	Body    *JSONReader
	Status  int
}

type JSONReader struct {
	*bytes.Reader
}

func NewJSONReader(outBytes []byte) *JSONReader {
	jr := new(JSONReader)
	jr.Reader = bytes.NewReader(outBytes)
	return jr
}

func (js JSONReader) MarshalJSON() ([]byte, error) {
	data, err := ioutil.ReadAll(js.Reader)
	if err != nil {
		return nil, err
	}
	data = []byte(`"` + string(data) + `"`)
	return data, nil
}

// UnmarshalJSON sets *jr to a copy of data.
func (jr *JSONReader) UnmarshalJSON(data []byte) error {
	if jr == nil {
		return errors.New("json.JSONReader: UnmarshalJSON on nil pointer")
	}
	if data == nil {
		return nil
	}
	data = []byte(strings.Trim(string(data), "\""))
	jr.Reader = bytes.NewReader(data)
	return nil
}
