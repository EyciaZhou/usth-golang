package usth
import (
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
)

type DbScore struct{}
var DBScore = &DbScore {}

const _QIU_API_ADDR = "http://www.usth.applinzi.com/api"

var (
	TIME_LOGIN = "登录时出错"
)

//Get: simple return reason of interface qiu's api
func (p *DbScore) Get(username string, password string, _type string) (_res []byte, _err error) {
	defer func() {
		err := recover()
		if err != nil {
			_err = newErrorByError(TIME_LOGIN, err.(error))
		}
	}()


	resp, err := http.PostForm(_QIU_API_ADDR, url.Values{
		"username" : []string{username},
		"password" : []string{password},
		"type" : []string{_type},
	})
	defer resp.Body.Close()
	if err != nil {
		return ([]byte)("Service Error"), newErrorByError(TIME_LOGIN, err)
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ([]byte)("Service Error"), newErrorByError(TIME_LOGIN, err)
	}

	_res = raw

	kv := map[string]interface{}{}
	err = json.Unmarshal(raw, &kv)
	if err != nil || kv["status"] != "ok" {
		return raw, newErrorByError(TIME_LOGIN, err) //password err, qiu's server err, etc
	}

	stu_id := username
	name := kv["_name"].(string)
	pwd := password

	DBInfo.Update(stu_id, pwd, name)
	return raw, nil
}

func (p *DbScore) GetFail(username string, password string)  (_res []byte, _err error) {
	return p.Get(username, password, "fail")
}

func (p *DbScore) GetPassing(username string, password string)  (_res []byte, _err error) {
	return p.Get(username, password, "passing")
}

func (p *DbScore) GetSemester(username string, password string)  (_res []byte, _err error) {
	return p.Get(username, password, "semester")
}
