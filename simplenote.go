package simplenote

// based on https://github.com/mrtazz/simplenote.py

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	authUrl  = "https://simple-note.appspot.com/api/login"
	dataUrl  = "https://simple-note.appspot.com/api2/data"
	indexUrl = "https://simple-note.appspot.com/api2/index?"
	zeros    = "0000000000"
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
	Content    string   `json:"content,omitempty""`
	Tags       []string `json:"tags,omitempty"`
	ModifyDate string   `json:"modifydate,omitempty"`
	Deleted    int      `json:"deleted"`
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

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func timeToStr(t time.Time) string {
	f := float64(t.UnixNano()) / 1000000000
	return fmt.Sprintf("%.9f", f)
}

func strToTime(st string) time.Time {
	parts := strings.Split(st, ".")
	if len(parts) != 2 {
		return time.Now()
	}
	secs, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Now()
	}

	nsStr := parts[1]
	n := 9 - len(nsStr)
	if n < 0 {
		return time.Now()
	}
	nsStr += zeros[:n]
	ns, err := strconv.ParseInt(nsStr, 10, 64)
	if err != nil {
		return time.Now()
	}
	return time.Unix(secs, ns)
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

func httpGet2(u string) (int, []byte, error) {
	resp, err := http.Get(u)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, d, nil
}

func httpGet(u string) ([]byte, error) {
	statusCode, d, err := httpGet2(u)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("status code %d (not 200) for %q", statusCode, u)
	}
	return d, nil
}

func httpGetRetry(u string) ([]byte, error) {
	statusCode, d, err := httpGet2(u)
	if err != nil {
		return nil, err
	}
	if statusCode == 500 {
		//fmt.Printf("Retrying %q\n", u)
		// unfortunately, simplenote requires this ridiculosly long backoff
		// time aftar a failing request
		time.Sleep(time.Second * 30)
		statusCode, d, err = httpGet2(u)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("GET: status code %d (not 200) for %q", statusCode, u)
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
		//fmt.Printf("%#v\n", resp)
		return nil, fmt.Errorf("POST status code %d (not 200), url: msg: %q", resp.StatusCode, u, string(d))
	}
	return d, nil
}

func httpDelete(urlStr string) error {
	req, err := http.NewRequest("DELETE", urlStr, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	//fmt.Printf("%#v\n", resp)
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d (not 200)", resp.StatusCode)
	}
	return nil
}

func (s *Api) getToken() (string, error) {
	if s.token != "" {
		return s.token, nil
	}

	auth_params := fmt.Sprintf("email=%s&password=%s", s.user, s.pwd)

	// TODO: use base64.URLEncoding ?
	values := base64.StdEncoding.EncodeToString([]byte(auth_params))
	token, err := httpPost(authUrl, values)
	if err != nil {
		//fmt.Printf("getToken: httpPost(%q,%q) failed with %q\n", authUrl, values, err)
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

	body, err := httpGetRetry(indexUrl + params)
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
	body, err := httpGet(dataUrl + params)
	if err != nil {
		return nil, err
	}
	var res apiNote
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("creation date: %s\n", res.CreateDate)
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

// TODO: change note to map[string]interface{}
func (api *Api) addUpdateNoteRaw(update map[string]interface{}) (*Note, error) {
	var ok bool
	authParam, err := api.getAuthUrlParams()
	if err != nil {
		return nil, err
	}
	var urlStr string
	keyI, hasKey := update["key"]
	key := ""
	if hasKey {
		key, ok = keyI.(string)
		if !ok {
			v := update["key"]
			return nil, fmt.Errorf("%T %v is not string", v, v)
		}
	}
	content, hasContent := update["content"]
	if key != "" {
		// this is update, so set modifydate
		update["modifydate"] = timeToStr(time.Now())
		urlStr = dataUrl + "/" + key + "?" + authParam
	} else {
		urlStr = dataUrl + "?" + authParam
	}

	js, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("url: %q\n js:\n%s\n", urlStr, string(js))
	// Note: python version does urllib.quote(values)
	s := url.QueryEscape(string(js))
	d, err := httpPost(urlStr, s)
	if err != nil {
		//fmt.Printf("getToken: httpPost(%q,%q) failed with %q\n", authUrl, values, err)
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
	if hasContent {
		res.Content = content.(string)
	}
	return res.toNote(), nil
}

func (api *Api) AddNote(content string, tags []string) (*Note, error) {
	update := make(map[string]interface{})
	update["content"] = content
	if len(tags) > 0 {
		update["tags"] = tags
	}
	return api.addUpdateNoteRaw(update)
}

func (api *Api) UpdateContent(key string, content string) error {
	update := make(map[string]interface{})
	update["key"] = key
	update["content"] = content
	_, err := api.addUpdateNoteRaw(update)
	return err
}

func (api *Api) UpdateTags(key string, tags []string) error {
	update := make(map[string]interface{})
	update["key"] = key
	update["content"] = tags
	_, err := api.addUpdateNoteRaw(update)
	return err
}

func (api *Api) TrashNote(key string) (*Note, error) {
	n, err := api.GetNoteLatestVersion(key)
	if err != nil {
		return nil, err
	}
	if n.IsDeleted {
		return n, nil
	}
	update := make(map[string]interface{})
	update["key"] = key
	update["deleted"] = 1
	return api.addUpdateNoteRaw(update)
}

func (api *Api) RestoreNote(key string) (*Note, error) {
	n, err := api.GetNoteLatestVersion(key)
	if err != nil {
		return nil, err
	}
	if !n.IsDeleted {
		return n, nil
	}
	update := make(map[string]interface{})
	update["key"] = key
	update["deleted"] = 0
	return api.addUpdateNoteRaw(update)
}

func (api *Api) DeleteNote(key string) error {
	// according to python version, the note must first be trash
	_, err := api.TrashNote(key)
	if err != nil {
		return err
	}

	authParam, err := api.getAuthUrlParams()
	if err != nil {
		return err
	}

	urlStr := dataUrl + fmt.Sprintf("/%s?%s", key, authParam)
	return httpDelete(urlStr)
}
