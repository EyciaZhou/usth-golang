package usth

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/wendal/errors"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"
	"sort"
)

type DbScore struct{}

var DBScore = &DbScore{}

var (
	_BLOCK_NAME_OF_SEMESTER = "学期成绩"

	_FIELDS_OF_SEMESTER = []string{
		"Id", "No", "Name", "EnglishName",
		"Credit", "Type", "Score", "NotPassReason",
	}

	_FIELDS_OF_XJXX = map[int]string{
		0:  "Stu_id",
		1:  "Name",
		29: "Class",
	}

	_BLOCKS_OF_NOT_PASS = []string{
		"尚未通过", "曾未通过",
	}

	_FIELDS_OF_NOT_PASS = []string{
		"Id", "No", "Name", "EnglishName",
		"Credit", "Type", "Score", "TestTime", "NotPassReason",
	}

	_FIELDS_OF_ALL = []string{
		"Id", "No", "Name", "EnglishName",
		"Credit", "Type", "Score",
	}
)

const (
	_HOST = "http://60.219.165.24"

	_LOGIN_DO = "/loginAction.do"
	_SEMESTER = "/bxqcjcxAction.do"
	_FAIL     = "/gradeLnAllAction.do?type=ln&oper=bjg"
	_ALL      = "/gradeLnAllAction.do?type=ln&oper=qbinfo"
	_XJXX     = "/xjInfoAction.do?oper=xjxx"
	_IMG      = "/xjInfoAction.do?oper=img"

	_CONTENT_TYPE          = "Content-Type"
	_USER_AGENT            = "User-Agent"
	_X_WWW_FORM_URLENCODED = "application/x-www-form-urlencoded"
	_UA                    = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.87 Safari/537.36"
)

type SchoolRollInfo struct {
	Stu_id string `json:"stu_id"`
	Name   string `json:"name"`
	Class  string `json:"class"`
}

type Course struct {
	Id            string `json:"id"`              //课程号
	No            string `json:"no"`              //课序号
	Name          string `json:"name"`            //课程名
	EnglishName   string `json:"en_name"`         //英文课程名
	Credit        string `json:"credit"`          //学分
	Type          string `json:"type"`            //课程属性
	Score         string `json:"score"`           //成绩
	TestTime      string `json:"test_time"`       //测试时间
	NotPassReason string `json:"not_pass_reason"` //未通过原因
}

type Block struct {
	BlockName string    `json:"block_name"`
	Data      []*Course `json:"courses"`
}

func TransToBlock(mp map[string][]*Course) []*Block {
	lp := len(mp)

	res := make([]*Block, lp)
	keys := make([]string, 0, lp)
	for k := range mp {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	p := 0
	for i := lp-1; i >= 0; i-- {
		res[p] = &Block{keys[i], mp[keys[i]]}
		p++
	}

	return res
}

type Fetcher struct {
	Client *http.Client
}

func (p *Fetcher) newReqForm(url string, par url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(par.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set(_CONTENT_TYPE, _X_WWW_FORM_URLENCODED)
	req.Header.Set(_USER_AGENT, _UA)

	return p.Client.Do(req)
}

func (p *Fetcher) newReqGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(_USER_AGENT, _UA)
	return p.Client.Do(req)
}

func (p *Fetcher) ensureGBKDecodeAndDocumentParse(resp *http.Response, err error) (_doc *goquery.Document, _err error) {
	defer func() {
		err := recover()
		if err != nil {
			_doc, _err = nil, err.(error)
		}
	}()

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return goquery.NewDocumentFromReader(
		transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder()),
	)
}

func (p *Fetcher) getDocument(URL string) (_doc *goquery.Document, _err error) {
	return p.ensureGBKDecodeAndDocumentParse(p.newReqGet(URL))
}

func (p *Fetcher) postDocument(URL string, par url.Values) (_doc *goquery.Document, _err error) {
	return p.ensureGBKDecodeAndDocumentParse(p.newReqForm(URL, par))
}

func (p *Fetcher) Login(username string, pwd string) error {
	doc, err := p.postDocument(_HOST+_LOGIN_DO, url.Values{
		"zjh": []string{username},
		"mm":  []string{pwd},
	})
	if err != nil {
		return err
	}

	if s := strings.TrimSpace(doc.Find("td.errorTop").Text()); s != "" {
		return errors.New(s)
	}

	return nil
}

func (p *Fetcher) SchoolRollInfo() (*SchoolRollInfo, error) {
	doc, err := p.getDocument(_HOST + _XJXX)
	if err != nil {
		return nil, err
	}

	res := &SchoolRollInfo{}

	doc.Find("[width='275']").Each(func(i int, s *goquery.Selection) {
		if _, ok := _FIELDS_OF_XJXX[i]; ok {
			reflect.ValueOf(res).Elem().FieldByName(_FIELDS_OF_XJXX[i]).SetString(strings.TrimSpace(s.Text()))
		}
	})

	return res, nil
}

func (p *Fetcher) Semester() ([]*Block, error) {
	doc, err := p.getDocument(_HOST + _SEMESTER)
	if err != nil {
		return nil, err
	}
	cs := []*Course{}
	doc.Find("tr.odd").Each(func(i int, s *goquery.Selection) {
		c := &Course{}
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			reflect.ValueOf(c).Elem().FieldByName(_FIELDS_OF_SEMESTER[i]).SetString(strings.TrimSpace(s.Text()))
		})
		cs = append(cs, c)
	})

	blocks := []*Block{&Block{
		_BLOCK_NAME_OF_SEMESTER,
		cs,
	}}

	return blocks, nil
}

func (p *Fetcher) NotPass() ([]*Block, error) {
	doc, err := p.getDocument(_HOST + _FAIL)
	if err != nil {
		return nil, err
	}

	res := map[string][]*Course{}

	doc.Find("table.titleTop2").Each(func(i int, s *goquery.Selection) {
		cs := []*Course{}
		s.Find("tr.odd").Each(func(i int, s *goquery.Selection) {
			c := &Course{}
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				reflect.ValueOf(c).Elem().FieldByName(_FIELDS_OF_NOT_PASS[i]).SetString(strings.TrimSpace(s.Text()))
			})
			cs = append(cs, c)
		})
		res[_BLOCKS_OF_NOT_PASS[i]] = cs
	})

	return TransToBlock(res), nil
}

func (p *Fetcher) All() ([]*Block, error) {
	doc, err := p.getDocument(_HOST + _ALL)
	if err != nil {
		return nil, err
	}

	res := map[string][]*Course{}
	mp := map[int]string{}

	doc.Find("table.title").Each(func(i int, s *goquery.Selection) {
		mp[i] = strings.TrimSpace(s.Text())
	})

	doc.Find("table.titleTop2").Each(func(i int, s *goquery.Selection) {
		cs := []*Course{}
		s.Find("tr.odd").Each(func(i int, s *goquery.Selection) {
			c := &Course{}
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				reflect.ValueOf(c).Elem().FieldByName(_FIELDS_OF_ALL[i]).SetString(strings.TrimSpace(s.Text()))
			})
			cs = append(cs, c)
		})

		if res[mp[i]] == nil {
			res[mp[i]] = cs
		} else {
			res[mp[i]] = append(res[mp[i]], cs...)
		}
	})

	return TransToBlock(res), nil
}

func NewFetcher() *Fetcher {
	jar, _ := cookiejar.New(nil)

	return &Fetcher{
		Client: &http.Client{
			Jar:     jar,
			Timeout: 3e9,
		},
	}
}

func (p *DbScore) LoginAndGetSchoolRollInfo(f *Fetcher, username string, password string) (map[string]interface{}, error) {
	if err := f.Login(username, password); err != nil {
		return nil, err
	}

	res := map[string]interface{}{}

	school_roll_info, err := f.SchoolRollInfo()
	if err != nil {
		return nil, err
	}

	res["school_roll_info"] = school_roll_info

	DBInfo.Update(school_roll_info.Stu_id, password, school_roll_info.Name, school_roll_info.Class)

	return res, nil
}

func (p *DbScore) Get(username string, password string, _type string) (map[string]interface{}, error) {
	f := NewFetcher()

	res, err := p.LoginAndGetSchoolRollInfo(f, username, password)
	if err != nil {
		return nil, err
	}

	typs := strings.Split(_type, "|")

	var bsall []*Block

	for _, typ := range typs {
		var bs []*Block

		if typ == "fail" {
			bs, err = f.NotPass()
		} else if typ == "passing" {
			bs, err = f.All()
		} else if typ == "semester" {
			bs, err = f.Semester()
		} else {
			return nil, errors.New("不支持的type")
		}

		if err != nil {
			return nil, err
		}

		if bsall == nil {
			bsall = bs
		} else {
			bsall = append(bsall, bs...)
		}
	}

	res["info"] = bsall

	return res, nil
}
