package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ReqType int

const (
	REQ_POST ReqType = iota
	REQ_GET
	REQ_DELETE
	REQ_PATCH
	REQ_PUT
)

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Body)
}

type ApiCallerProps[Tbody any] struct {
	BaseURL string
	Paths   []string
	Params  map[string]string
	ReqType ReqType
	Body    Tbody
	Headers map[string]string
}

func ApiCaller[TBody any, TResp any](args ApiCallerProps[TBody]) (TResp, error) {
	var resp TResp

	u, err := url.Parse(args.BaseURL)
	if err != nil {
		return resp, err
	}

	u.Path = strings.Join(append([]string{u.Path}, args.Paths...), "/")

	q := u.Query()
	for k, v := range args.Params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	var body io.Reader
	if args.ReqType != REQ_GET {
		b, err := json.Marshal(args.Body)
		if err != nil {
			return resp, err
		}
		body = strings.NewReader(string(b))
	}

	method := map[ReqType]string{
		REQ_GET:    http.MethodGet,
		REQ_POST:   http.MethodPost,
		REQ_PATCH:  http.MethodPatch,
		REQ_PUT:    http.MethodPut,
		REQ_DELETE: http.MethodDelete,
	}[args.ReqType]
	log.Println("url for getting football api response :: ", u)
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return resp, err
	}
	for k, v := range args.Headers {
		req.Header.Add(k, v)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return resp, &HTTPError{
			StatusCode: res.StatusCode,
			Body:       string(b),
		}
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}

	if err := json.Unmarshal(b, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func FootBallOrgAPICaller[TBody any, TResp any](args ApiCallerProps[TBody], key *string, keys []string) (TResp, error) {

	if key != nil {
		resp, err := ApiCaller[TBody, TResp](ApiCallerProps[TBody]{
			BaseURL: args.BaseURL,
			Params:  args.Params,
			Paths:   args.Paths,
			ReqType: args.ReqType,
			Headers: map[string]string{
				"X-Auth-Token": *key,
			},
		})
		if err != nil {
			var httpErr *HTTPError
			if errors.As(err, &httpErr) {
				if httpErr.StatusCode == http.StatusTooManyRequests {
					var retryKeys []string
					for _, k := range keys {
						if k != *key {
							retryKeys = append(retryKeys, k)
						}
					}
					if len(retryKeys) == 0 {
						return resp, err
					}
					resp, err = retryWithKeys[TBody, TResp](args, retryKeys)
					if err != nil {
						return resp, err
					}
					return resp, nil
				}
			}
			return resp, err
		}
	}

	return retryWithKeys[TBody, TResp](args, keys)

}

func retryWithKeys[TBody, TResp any](args ApiCallerProps[TBody], keys []string) (TResp, error) {
	var resp TResp
	if len(keys) == 0 {
		return resp, errors.New("no api keys provided")
	}
	var finalErr error
	for _, val := range keys {
		resp, finalErr = ApiCaller[TBody, TResp](ApiCallerProps[TBody]{
			BaseURL: args.BaseURL,
			Params:  args.Params,
			Paths:   args.Paths,
			ReqType: args.ReqType,
			Headers: map[string]string{
				"X-Auth-Token": val,
			},
		})
		if finalErr == nil {
			return resp, nil
		}

	}
	if finalErr != nil {
		return resp, finalErr
	}
	return resp, nil

}
