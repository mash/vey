package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/mash/vey"
)

// Client is a HTTP client that consumes Vey's HTTP APIs.
type Client struct {
	*http.Client
	root url.URL
}

// NewClient creates a new HTTP client that consumes Vey's HTTP APIs, given a root URL.
func NewClient(root string) Client {
	u, err := url.Parse(root)
	if err != nil {
		panic(err)
	}
	return Client{
		Client: &http.Client{
			// Client does not follow redirects, to be able to get the Location header in Open().
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		root: *u,
	}
}

// GetKeys calls the GetKeys interface on the Vey server and returns a slice of PublicKeys.
func (c Client) GetKeys(email string) ([]vey.PublicKey, error) {
	res, err := c.Do("/getKeys", Body{Email: email})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var keys []vey.PublicKey
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&keys); err != nil {
		return nil, err
	}
	return keys, nil
}

// BeginDelete calls the BeginDelete interface on the Vey server.
func (c Client) BeginDelete(email string, publicKey vey.PublicKey) error {
	res, err := c.Do("/beginDelete", Body{Email: email, PublicKey: publicKey})
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// CommitDelete calls the CommitDelete interface on the Vey server.
func (c Client) CommitDelete(token []byte) error {
	q := url.Values{}
	q.Add("token", base64.StdEncoding.EncodeToString(token))
	res, err := c.Get("/commitDelete", q)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// BeginPut calls the BeginPut interface on the Vey server.
func (c Client) BeginPut(email string, publicKey vey.PublicKey) error {
	res, err := c.Do("/beginPut", Body{Email: email, PublicKey: publicKey})
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// CommitPut calls the CommitPut interface on the Vey server.
func (c Client) CommitPut(challenge, signature []byte) error {
	res, err := c.Do("/commitPut", Body{
		Challenge: challenge,
		Signature: signature,
	})
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// Open calls the /open path on the Vey server, and returns the redirect response's Location header.
func (c Client) Open(query url.Values) (string, error) {
	// We're expecting a 302 response
	_, err := c.Get("/open", query)
	if err != nil {
		if er, ok := err.(ClientError); ok && er.Res != nil && er.Res.StatusCode == 302 {
			return er.Res.Header.Get("Location"), nil
		}
	}
	return "", err
}

// Do does request preparation and error handling common to all of Vey's HTTP APIs.
// Caller should Close the response.Body when error is nil.
func (c Client) Do(path string, body Body) (*http.Response, error) {
	u := c.root.ResolveReference(&url.URL{Path: path})
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(body); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", u.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		defer res.Body.Close()

		body := Error{}
		dec := json.NewDecoder(res.Body)
		if err := dec.Decode(&body); err != nil {
			return nil, ClientError{Msg: "json decode error", Res: res, Err: err}
		}
		return nil, ClientError{Msg: body.Msg, Res: res, Err: nil}
	}
	return res, nil
}

func (c Client) Get(path string, q url.Values) (*http.Response, error) {
	u := c.root.ResolveReference(&url.URL{Path: path})
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		defer res.Body.Close()

		body := Error{}
		dec := json.NewDecoder(res.Body)
		if err := dec.Decode(&body); err != nil {
			return nil, ClientError{Msg: "json decode error", Res: res, Err: err}
		}
		return nil, ClientError{Msg: body.Msg, Res: res, Err: nil}
	}
	return res, nil
}
