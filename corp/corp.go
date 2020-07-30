package corp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/toolkits/pkg/logger"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const dingTimeOut = time.Second * 1

// Err
type Err struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Client
type Client struct {
	Mobiles   []string
	Chatid    string
	Appid     string
	Appsecret string
	openUrl   string
	IsAtAll   bool
	Token     string
}


// Result 发送消息返回结果
type Token struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	AppAccessToken    string `json:"app_access_token"`
	Expire            int    `json:"expire"`
	TenantAccessToken string `json:"tenant_access_token"`
}

type MobilesOpenId struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Data   `json:"data"`
}

type Data struct {
	MobileUsers map[string][]User `json:"mobile_users"`
}

type User struct {
	UserId string `json:"user_id"`
	OpenId string `json:"open_id"`
}

type Result struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data Message `json:"data"`
}

type Message struct {
	MessageId string `json:"message_id"`
}

func GetToken(appid,appsecret string) (string, error) {
	url := "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal/"
	postData := make(map[string]interface{})
	postData["app_id"] = appid
	postData["app_secret"] = appsecret

	jsonBody, err := encodeJSON(postData)
	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		logger.Info("ding talk new post request err =>", err)
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := getClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("ding talk post request err =>", err)
		fmt.Println(err)
	}

	defer resp.Body.Close()
	resultByte, err :=ioutil.ReadAll(resp.Body);
	if err == nil {
		result := Token{}
		err = json.Unmarshal(resultByte, &result)
		if err != nil {
			fmt.Errorf("parse send api response fail: %v", err)
		}

		if result.Code != 0 {
			err = fmt.Errorf("invoke send api return ErrCode = %d, ErrMsg = %s ", result.Code, result.Msg)
		}
		return "Bearer " + result.TenantAccessToken, nil
	}
	return "Bearer ", err

}

func (c Client) GetOpneIdFromMobiles(phones []string) (map[string][]User, error) {
	myurl := "https://open.feishu.cn/open-apis/user/v1/batch_get_id"
	//postData := make(map[string][]string)
	//postData["mobiles"] = phones
	//fmt.Println(phones)

	//jsonBody, err := encodeJSON(postData)
	//if err != nil {
	//	fmt.Println(err)
	//}
	// {"code":0,"msg":"success","data":{"mobile_users":{"18810223379":[{"open_id":"ou_4473776b6762b9ea317407957b3fa22e","user_id":"f6964gdd"}]}}}
	params := url.Values{
		"mobiles":  phones,

	}

	req, err := http.NewRequest("GET", myurl+"?"+params.Encode(),nil)
	if err != nil {
		logger.Info("ding talk new post request err =>", err)
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.Token)

	client := getClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("ding talk post request err =>", err)
		fmt.Println(err)
	}

	defer resp.Body.Close()
	resultByte, err :=ioutil.ReadAll(resp.Body);
	//fmt.Println(string(resultByte))
	result := MobilesOpenId{}
	if err == nil {
		err = json.Unmarshal(resultByte, &result)
		if err != nil {
			fmt.Errorf("parse send api response fail: %v", err)
		}
		if result.Code != 0 {
			err = fmt.Errorf("invoke send api return ErrCode = %d, ErrMsg = %s ", result.Code, result.Msg)
		}
		return result.Data.MobileUsers, nil
	}
	return result.Data.MobileUsers, err

}

// New
func New(chatid string, mobiles []string, isAtAll bool, appid,appsecret string) *Client {
	c := new(Client)
	token, err := GetToken(appid,appsecret)
	if err != nil {
		fmt.Println(err)
	}
	c.openUrl = "https://open.feishu.cn/open-apis/message/v4/send/"
	c.Chatid = chatid
	c.Mobiles = mobiles
	c.IsAtAll = isAtAll
	c.Appid = appid
	c.Appsecret = appsecret
	c.Token = token
	return c
}

func (c Client) GetChatid() string {

	return c.Chatid
}

// Send 发送信息
func (c *Client) Send(chatid string, mobile []string, msg string) error {
	c.Chatid = chatid
	postData := c.generateData(mobile, msg)
	if c.GetChatid() != "" {
		// 配置了token 说明采用配置文件的token
		chatid = c.GetChatid()
	}


	resultByte, err := jsonPost(c.openUrl, postData, c.Token)
	if err != nil {
		return fmt.Errorf("invoke send api fail: %v", err)
	}

	result := Result{}
	err = json.Unmarshal(resultByte, &result)
	if err != nil {
		return fmt.Errorf("parse send api response fail: %v", err)
	}

	if result.Code != 0 || result.Msg != "ok" {
		err = fmt.Errorf("200 invoke send api return ErrCode = %d, ErrMsg = %s ", result.Code, result.Msg)
		token, err := GetToken(c.Appid, c.Appsecret)
		if err != nil {
			fmt.Println(err)
		}
		c.Token = token

	}

	return err
}

func jsonPost(url string, data interface{}, token string) ([]byte, error) {
	jsonBody, err := encodeJSON(data)
	if err != nil {
		return nil, err
	}
	fmt.Println(token)
	fmt.Println(string(jsonBody))
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		logger.Info("ding talk new post request err =>", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	client := getClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("ding talk post request err =>", err)
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func encodeJSON(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}


func (c *Client) generateData(mobile []string, msg string) interface{} {
	postData := make(map[string]interface{})
	postData["msg_type"] = "text"
	postData["chat_id"] = c.Chatid
	sendContext := make(map[string]interface{})

	at := make(map[string]interface{})
	if !c.IsAtAll && len(c.Mobiles) > 0 {
		at["atMobiles"] = c.Mobiles // 根据手机号@指定人
	} else if len(mobile) > 0{
		at["atMobiles"] = mobile // 根据手机号@指定人
	} else {
		c.IsAtAll = true
	}
	if (!c.IsAtAll) {
		data, err := c.GetOpneIdFromMobiles(mobile)
		if err == nil {
			fmt.Println("Success        ")
			for _,v :=range data {
				for i:=0;i<len(v);i++{
					tmp_at := fmt.Sprintf("<at user_id=\"%s\">test</at>",v[i].OpenId)
					msg += tmp_at
				}

			}
		}

	}

	sendContext["text"] = msg
	postData["content"] = sendContext

	return postData
}

func getClient() *http.Client {
	// 通过http.Client 中的 DialContext 可以设置连接超时和数据接受超时 （也可以使用Dial, 不推荐）
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
				conn, err := net.DialTimeout(network, addr, dingTimeOut) // 设置建立链接超时
				if err != nil {
					return nil, err
				}
				_ = conn.SetDeadline(time.Now().Add(dingTimeOut)) // 设置接受数据超时时间
				return conn, nil
			},
			ResponseHeaderTimeout: dingTimeOut, // 设置服务器响应超时时间
		},
	}
}
