package simplenote

// api docs: https://simperium.com/docs/reference/http/
// for even newer api: https://github.com/Simperium/simperium-protocol/blob/master/SYNCING.md#index-i

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	simpleNoteAppID  = "chalk-bump-f49"
	authURL2         = "https://auth.simperium.com/1/"
	apiURL2          = "https://api.simperium.com/1/"
	tokenHeaderName  = "X-Simperium-Token"
	apiKeyHeaderName = "X-Simperium-API-Key"
	bucketName       = "Note"
)

// RawLogger is for logging raw HTTP interactions
type RawLogger interface {
	Log(s string)
}

// Note describes a single note
type Note struct {
	ID               string
	Version          int
	Tags             []string `json:",omitempty"`
	Deleted          bool
	Content          string
	SystemTags       []string `json:",omitempty"`
	ModificationDate time.Time
	CreationDate     time.Time
}

// NoteID describes a version of a note
type NoteID struct {
	ID      string     `json:"id"`
	Version int        `json:"v"`
	Note    *iResponse `json:"d"`
}

type indexResponse struct {
	Current string   `json:"current"`
	Mark    string   `json:"mark"`
	Index   []NoteID `json:"index"`
}

type iResponse struct {
	Tags             []string    `json:"tags"`
	Deleted          interface{} `json:"deleted"`
	ShareURL         string      `json:"shareURL"`
	PublishURL       string      `json:"publishURL"`
	Content          string      `json:"content"`
	SystemTags       []string    `json:"systemTags"`
	ModificationDate float64     `json:"modificationDate"`
	CreationDate     float64     `json:"creationDate"`
}

type loginResponse struct {
	UserName    string `json:"username"`
	AccessToken string `json:"access_token"`
	UserID      string `json:"userid"`
}

// Client describes SimpleNote client
type Client struct {
	user           string
	pwd            string
	simperiumToken string
	appID          string
	login          *loginResponse
	Logger         RawLogger
}

func timeToStr(t time.Time) string {
	f := float64(t.UnixNano()) / 1000000000
	return fmt.Sprintf("%.9f", f)
}

func httpGet2(uri string) (int, []byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return 0, nil, err
	}
	return httpReadResponse(resp)
}

func httpReadResponse(resp *http.Response) (int, []byte, error) {
	defer resp.Body.Close()
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, d, nil
}

func httpReadReq(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	statusCode, d, err := httpReadResponse(resp)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("status code %d (not 200) for %s", statusCode, req.URL)
	}
	return d, nil
}

func httpGet(uri string) ([]byte, error) {
	statusCode, d, err := httpGet2(uri)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("GET: status code %d (not 200) for %q", statusCode, uri)
	}
	return d, nil
}

func httpGetRetry(uri string) ([]byte, error) {
	statusCode, d, err := httpGet2(uri)
	if err != nil {
		return nil, err
	}
	if statusCode == 500 {
		//fmt.Printf("Retrying %q\n", u)
		// unfortunately, simplenote requires this ridiculosly long backoff
		// time aftar a failing request
		time.Sleep(time.Second * 30)
		statusCode, d, err = httpGet2(uri)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("GET: status code %d (not 200) for %q", statusCode, uri)
	}
	return d, nil
}

func httpPost(uri string, body string) ([]byte, error) {
	r := strings.NewReader(body)
	req, err := http.NewRequest("POST", uri, r)
	if err != nil {
		return nil, err
	}
	return httpReadReq(req)
}

func httpDelete2(uri string) (int, error) {
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}

func httpDelete(uri string) error {
	//fmt.Printf("%#v\n", resp)
	statusCode, err := httpDelete2(uri)
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return fmt.Errorf("DELETE status code %d (not 200) url: %q", statusCode, uri)
	}
	return nil
}

func httpDeleteRetry(uri string) error {
	statusCode, err := httpDelete2(uri)
	if err != nil {
		return err
	}
	if statusCode == 500 {
		// unfortunately, simplenote requires this ridiculosly long backoff
		// time aftar a failing request
		time.Sleep(time.Second * 30)
		statusCode, err = httpDelete2(uri)
	}
	if statusCode != 200 {
		return fmt.Errorf("DELETE status code %d (not 200) url: %q", statusCode, uri)
	}
	return nil
}

// NewClient creates a new SimpleNote client
func NewClient(simperiumToken, user, pwd string) *Client {
	return &Client{
		simperiumToken: simperiumToken,
		user:           user,
		pwd:            pwd,
		appID:          simpleNoteAppID,
	}
}

func (c *Client) logRaw(format string, args ...interface{}) {
	if c.Logger != nil {
		s := fmt.Sprintf(format, args...)
		c.Logger.Log(s)
	}
}

// e.g. /authorize/
func (c *Client) authURL(path string) string {
	return authURL2 + c.appID + path
}

// path must start with
func (c *Client) apiURL(path string, args ...string) string {
	var urlArgs string
	if len(args) > 0 {
		if len(args)%2 != 0 {
			panic("number of args must be even")
		}
		n := len(args) / 2
		v := url.Values{}
		for i := 0; i < n; i++ {
			v.Add(args[i*2], args[i*2+1])
		}
		urlArgs = "?" + v.Encode()
	}
	uri := apiURL2 + c.appID + "/" + bucketName + path + urlArgs
	//fmt.Printf("uri: '%s'\n", uri)
	return uri
}

func (c *Client) loginJSON() string {
	m := make(map[string]string)
	m["username"] = c.user
	m["password"] = c.pwd
	d, _ := json.Marshal(m)
	return string(d)
}

func (c *Client) loginIfNeeded() error {
	if c.login != nil {
		return nil
	}
	body := c.loginJSON()
	r := strings.NewReader(body)
	uri := c.authURL("/authorize/")
	req, err := http.NewRequest("POST", uri, r)
	if err != nil {
		return err
	}
	req.Header.Add(apiKeyHeaderName, c.simperiumToken)
	req.Header.Add("Content-Type", "application/json")
	d, err := httpReadReq(req)
	if err != nil {
		return err
	}
	c.logRaw("%s\n%s\n\n", uri, string(d))
	//fmt.Printf("auth response: '%s'\n", string(d))
	var rsp loginResponse
	err = json.Unmarshal(d, &rsp)
	if err != nil {
		return err
	}
	//fmt.Printf("%#v\n", rsp)
	c.login = &rsp
	return nil
}

func (c *Client) listRaw(mark string) (*indexResponse, error) {
	err := c.loginIfNeeded()
	if err != nil {
		return nil, err
	}
	args := []string{"limit", "100", "data", "1"}
	if mark != "" {
		args = append(args, "mark")
		args = append(args, mark)
	}
	uri := c.apiURL("/index", args...)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(tokenHeaderName, c.login.AccessToken)
	d, err := httpReadReq(req)
	if err != nil {
		return nil, err
	}
	c.logRaw("%s\n%s\n\n", uri, string(d))
	//fmt.Printf("list response: '%s'\n", string(d))
	var v indexResponse
	err = json.Unmarshal(d, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func timeFromFloat(t float64) time.Time {
	sec := int64(t)
	nsec := int64(t*1e6) % 1e6
	return time.Unix(sec, nsec)
}

func toBool(v interface{}) bool {
	switch v := v.(type) {
	case int:
		if v == 0 {
			return false
		}
		if v == 1 {
			return true
		}
		log.Fatalf("invalid bool int value %d\n", v)
	case float64:
		if int(v) == 0 {
			return false
		}
		if int(v) == 1 {
			return true
		}
		log.Fatalf("invalid float64 value %.2f\n", v)
	case bool:
		return v
	default:
		log.Fatalf("unexpected type %T\n", v)
	}
	return false
}

// List lists notes
func (c *Client) List() ([]*Note, error) {
	var res []*Note
	var mark string
	for {
		curr, err := c.listRaw(mark)
		if err != nil {
			return nil, err
		}
		for _, n := range curr.Index {
			res = append(res, toNote(n.ID, n.Version, n.Note))
		}
		mark = curr.Mark
		if mark == "" {
			break
		}
	}
	return res, nil
}

func toNote(id string, version int, v *iResponse) *Note {
	return &Note{
		ID:               id,
		Version:          version,
		Tags:             v.Tags,
		Deleted:          toBool(v.Deleted),
		Content:          v.Content,
		SystemTags:       v.SystemTags,
		ModificationDate: timeFromFloat(v.ModificationDate),
		CreationDate:     timeFromFloat(v.CreationDate),
	}
}

// GetNote downloads a note
func (c *Client) GetNote(noteID string, version int) (*Note, error) {
	err := c.loginIfNeeded()
	if err != nil {
		return nil, err
	}
	uri := c.apiURL(fmt.Sprintf("/i/%s/v/%d", noteID, version))
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(tokenHeaderName, c.login.AccessToken)
	d, err := httpReadReq(req)
	if err != nil {
		return nil, err
	}
	c.logRaw("%s\n%s\n\n", uri, string(d))
	//fmt.Printf("resp: '%s'\n", string(d))
	var v iResponse
	err = json.Unmarshal(d, &v)
	if err != nil {
		log.Fatalf("failed to unmarshal '%s' with '%s'\n", string(d), err)
		return nil, err
	}

	return toNote(noteID, version, &v), nil
}
