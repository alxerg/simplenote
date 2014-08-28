package simplenote

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	AUTH_URL       = "https://simple-note.appspot.com/api/login"
	DATA_URL       = "https://simple-note.appspot.com/api2/data"
	INDEX_URL      = "https://simple-note.appspot.com/api2/index?"
	NOTES_TO_FETCH = 100
)

type Api struct {
	user  string
	pwd   string
	token string
}

type apiNoteInfo struct {
	ModifyDate string   `json:"modifydate"`
	Tags       []string `json:"tags"`
	Deleted    int      `json:"deleted"`
	CreateDate string   `json:"createdate"`
	SystemTags []string `json:"systemtags"`
	Version    int      `json:"version"`
	SyncNum    int      `json:"syncnum"`
	Key        string   `json:"key"`
	MinVersion int      `json:"minversion"`
}

type NoteInfo struct {
	Key        string
	CreateDate time.Time
	ModifyDate time.Time
	Tags       []string
	IsDeleted  bool
	SystemTags []string
	Version    int
	SyncNum    int
}

func intToBool(n int) bool {
	if n == 0 {
		return false
	}
	return true
}

func floatToTime(ft float64) time.Time {
	// TODO: write me
	return time.Now()
}

func strToTime(st string) time.Time {
	// TODO: write me
	return time.Now()
}

func (a *apiNoteInfo) ToNoteInfo() NoteInfo {
	return NoteInfo{
		Key:        a.Key,
		CreateDate: strToTime(a.CreateDate),
		ModifyDate: strToTime(a.ModifyDate),
		Tags:       a.Tags,
		IsDeleted:  intToBool(a.Deleted),
		SystemTags: a.SystemTags,
		Version:    a.Version,
		SyncNum:    a.SyncNum,
	}
}

type apiNoteListResponse struct {
	Count int            `json:"count"`
	Data  []*apiNoteInfo `json:"data"`
	Time  string         `json:"time"`
	Mark  string         `json:"mark"`
}

func New(user, pwd string) *Api {
	return &Api{
		user: user,
		pwd:  pwd,
	}
}

func httpGet(u string) ([]byte, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d (not 200)", resp.StatusCode)
	}
	return d, nil
}

func httpPostWithBody(u string, body string) ([]byte, error) {
	r := strings.NewReader(body)
	req, err := http.NewRequest("POST", u, r)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	d, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d (not 200), msg: %q", resp.StatusCode, string(d))
	}
	return d, nil
}

func (s *Api) getToken() (string, error) {
	if s.token != "" {
		return s.token, nil
	}

	auth_params := fmt.Sprintf("email=%s&password=%s", s.user, s.pwd)

	// TODO: use base64.URLEncoding ?
	values := base64.StdEncoding.EncodeToString([]byte(auth_params))
	token, err := httpPostWithBody(AUTH_URL, values)
	if err != nil {
		fmt.Printf("getToken: httpGetWithBody(%q,%q) failed with %q\n", AUTH_URL, values, err)
		return "", err
	}
	fmt.Printf("token: %q\n", string(token))
	s.token = string(token)
	return s.token, nil
}

func (s *Api) getNoteListRaw(mark string, since time.Time) (*apiNoteListResponse, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}
	params := fmt.Sprintf("auth=%s&email=%s&length=%d", url.QueryEscape(token), url.QueryEscape(s.user), NOTES_TO_FETCH)
	if !since.IsZero() {
		params += "&since=" + since.Format("2006-01-02")
	}
	if mark != "" {
		params += fmt.Sprintf("&mark=%s", url.QueryEscape(mark))
	}

	body, err := httpGet(INDEX_URL + params)
	if err != nil {
		return nil, err
	}
	var res apiNoteListResponse
	fmt.Printf("resp: \n%s\n", string(body))
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *Api) GetNoteList() ([]*NoteInfo, error) {
	var zeroTime time.Time
	res := make([]*NoteInfo, 0)
	mark := ""
	for {
		rsp, err := s.getNoteListRaw(mark, zeroTime)
		if err != nil {
			// TODO: return as much as we got?
			return nil, err
		}
		for _, ani := range rsp.Data {
			ni := ani.ToNoteInfo()
			res = append(res, &ni)
		}
		mark = rsp.Mark
		if mark == "" {
			break
		}
		// TODO: also break if len(rsp.Data) == 0 ?
	}
	return res, nil
}
