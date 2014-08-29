package simplenote

// based on https://github.com/mrtazz/simplenote.py

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
	AUTH_URL  = "https://simple-note.appspot.com/api/login"
	DATA_URL  = "https://simple-note.appspot.com/api2/data"
	INDEX_URL = "https://simple-note.appspot.com/api2/index?"
)

var (
	NotesPerRequestsCount int = 100
)

type Api struct {
	user  string
	pwd   string
	token string
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

type apiNewNote struct {
	Key        string   `json:"key,omitempty"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags,omitempty"`
	ModifyDate string   `json:"modifydate,omitempty"`
}

type apiNoteInfo struct {
	ModifyDate string   `json:"modifydate"`
	Tags       []string `json:"tags,omitempty"`
	Deleted    int      `json:"deleted"`
	CreateDate string   `json:"createdate"`
	SystemTags []string `json:"systemtags,omitempty"`
	Version    int      `json:"version"`
	SyncNum    int      `json:"syncnum"`
	Key        string   `json:"key"`
	MinVersion int      `json:"minversion"`
}

type Note struct {
	ModifyDate time.Time
	Tags       []string
	IsDeleted  bool
	CreateDate time.Time
	SystemTags []string
	Content    string
	Version    int
	SyncNum    int
	Key        string
	MinVersion int
}

type apiNote struct {
	ModifyDate string   `json:"modifydate"`
	Tags       []string `json:"tags"`
	Deleted    int      `json:"deleted"`
	CreateDate string   `json:"createdate"`
	SystemTags []string `json:"systemtags"`
	Content    string   `json:"content"`
	Version    int      `json:"version"`
	SyncNum    int      `json:"syncnum"`
	Key        string   `json:"key"`
	MinVersion int      `json:"minversion"`
}

func (n *apiNote) toNote() *Note {
	return &Note{
		ModifyDate: strToTime(n.ModifyDate),
		Tags:       n.Tags,
		IsDeleted:  intToBool(n.Deleted),
		CreateDate: strToTime(n.CreateDate),
		SystemTags: n.SystemTags,
		Content:    n.Content,
		Version:    n.Version,
		SyncNum:    n.SyncNum,
		Key:        n.Key,
		MinVersion: n.MinVersion,
	}
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

func timeToStr(t time.Time) string {
	f := float64(t.UnixNano()) / 1000000000
	return fmt.Sprintf("%.6f", f)
}

func strToTime(st string) time.Time {
	// TODO: write me
	return time.Now()
}

func (a *apiNoteInfo) toNoteInfo() *NoteInfo {
	return &NoteInfo{
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

func httpPost(u string, body string) ([]byte, error) {
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
		fmt.Printf("%#v\n", resp)
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
	token, err := httpPost(AUTH_URL, values)
	if err != nil {
		//fmt.Printf("getToken: httpGetWithBody(%q,%q) failed with %q\n", AUTH_URL, values, err)
		return "", err
	}
	//fmt.Printf("token: %q\n", string(token))
	s.token = string(token)
	return s.token, nil
}

func (api *Api) getAuthUrlParams() (string, error) {
	token, err := api.getToken()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("auth=%s&email=%s", url.QueryEscape(token), url.QueryEscape(api.user)), nil
}

func (api *Api) getNoteListRaw(mark string, since time.Time) (*apiNoteListResponse, error) {
	authParam, err := api.getAuthUrlParams()
	if err != nil {
		return nil, err
	}
	params := fmt.Sprintf("%s&length=%d", authParam, NotesPerRequestsCount)
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
	//fmt.Printf("resp: \n%s\n", string(body))
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func reachedLimit(n, limit int) bool {
	if limit == -1 {
		return false
	}
	return n > limit
}

// if limit is -1, no limit
func (api *Api) GetNoteListWithLimit(limit int) ([]*NoteInfo, error) {
	var zeroTime time.Time
	res := make([]*NoteInfo, 0)
	mark := ""
	for {
		rsp, err := api.getNoteListRaw(mark, zeroTime)
		if err != nil {
			// TODO: return as much as we got?
			return nil, err
		}
		for _, ani := range rsp.Data {
			res = append(res, ani.toNoteInfo())
			if reachedLimit(len(res), limit) {
				return res, nil
			}
		}
		mark = rsp.Mark
		if mark == "" {
			break
		}
		// TODO: also break if len(rsp.Data) == 0 ?
	}
	return res, nil
}

func (api *Api) GetNoteList() ([]*NoteInfo, error) {
	return api.GetNoteListWithLimit(-1)
}

// if version is -1, return latest version
func (api *Api) getNoteRaw(key string, version int) (*apiNote, error) {
	authParam, err := api.getAuthUrlParams()
	if err != nil {
		return nil, err
	}
	ver := ""
	if version != -1 {
		ver = fmt.Sprintf("/%d", version)
	}
	params := fmt.Sprintf("/%s%s?%s", key, ver, authParam)
	body, err := httpGet(DATA_URL + params)
	if err != nil {
		return nil, err
	}
	var res apiNote
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (api *Api) GetNote(key string, version int) (*Note, error) {
	note, err := api.getNoteRaw(key, version)
	if err != nil {
		return nil, err
	}
	return note.toNote(), nil
}

func (api *Api) GetNoteLatestVersion(key string) (*Note, error) {
	return api.GetNote(key, -1)
}

func (api *Api) addUpdateNoteRaw(note *apiNewNote) (*Note, error) {
	authParam, err := api.getAuthUrlParams()
	if err != nil {
		return nil, err
	}
	var urlStr string
	if note.Key != "" {
		// this is update, so set modifydate
		timeStr := timeToStr(time.Now())
		note.ModifyDate = timeStr
		urlStr = DATA_URL + "/" + note.Key + "?" + authParam
	} else {
		urlStr = DATA_URL + "?" + authParam
	}

	js, err := json.Marshal(note)
	if err != nil {
		return nil, err
	}
	fmt.Printf("url: %q\n js:\n%s\n", urlStr, string(js))
	// Note: python version does urllib.quote(values)
	s := url.QueryEscape(string(js))
	d, err := httpPost(urlStr, s)
	if err != nil {
		//fmt.Printf("getToken: httpGetWithBody(%q,%q) failed with %q\n", AUTH_URL, values, err)
		return nil, err
	}
	//fmt.Printf("%s\n", string(d))
	// TODO: write a function for that?
	var res apiNote
	err = json.Unmarshal(d, &res)
	if err != nil {
		return nil, err
	}
	// returned json response doesn't return the content, so set it to
	// what we've sent
	res.Content = note.Content
	return res.toNote(), nil
}

func (api *Api) AddNote(content string, tags []string) (*Note, error) {
	n := &apiNewNote{
		Content: content,
		Tags:    tags,
	}
	return api.addUpdateNoteRaw(n)
}

func (api *Api) DeleteNote(key string) error {
	// TODO: implement me
	return nil
}
