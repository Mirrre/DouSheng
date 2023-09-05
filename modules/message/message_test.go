package message

import (
	"app/config"
	"app/modules/models"
	"app/utils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

// Golang的测试会检测到每个包的TestMain函数，首先执行它
func TestMain(m *testing.M) {
	utils.Setup()
	postSetup()
	code := m.Run()
	utils.Teardown()
	os.Exit(code)
}

var SendUrl = "/douyin/message/action/"
var GetHistoryUrl = "/douyin/message/chat/"
var db = utils.GetDb()

func postSetup() {
	// Create 2 users
	jordan := models.User{
		Username: "jordan",
		Password: "jordan_pass",
		Profile:  models.UserProfile{Avatar: "jordan.jpg"},
	}
	michael := models.User{
		Username: "michael",
		Password: "michael_pass",
		Profile:  models.UserProfile{Avatar: "michael.jpg"},
	}
	db.Create(&jordan)
	db.Create(&michael)
}

// 测试发送消息
func TestSend(t *testing.T) {
	config.Router.POST(SendUrl, Send)

	// 测试成功发送
	token, err := utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}
	values := url.Values{}
	values.Add("token", token)
	values.Add("to_user_id", "2")
	values.Add("action_type", "1")
	values.Add("content", "I am jordan,are you michael?")
	reqURL := SendUrl + "?" + values.Encode()
	req, _ := http.NewRequest("POST", reqURL, nil)

	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	assert.Equal(t, http.StatusCreated, response.Code)
	// 测试message是否创建
	message := models.Message{FromUserID: 1}
	result := db.First(&message)
	assert.Equal(t, int64(1), result.RowsAffected)

	// 测试失败发送(自己给自己发消息）
	values.Set("to_user_id", "1")
	reqURL = SendUrl + "?" + values.Encode()
	req, _ = http.NewRequest("POST", reqURL, nil)

	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	assert.Equal(t, http.StatusBadRequest, response.Code)

	// 测试失败发送（给不存在的用户发消息）
	values.Set("to_user_id", "0")
	reqURL = SendUrl + "?" + values.Encode()
	req, _ = http.NewRequest("POST", reqURL, nil)

	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestGetHistory(t *testing.T) {
	config.Router.GET(GetHistoryUrl, GetHistory)
	// 创建消息记录
	message := models.Message{
		FromUserID: 1,
		ToUserID:   2,
		Content:    "hello",
	}
	db.Create(&message)
	// 测试成功查询
	token, err := utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}
	values := url.Values{}
	values.Add("token", token)
	values.Add("to_user_id", "2")
	values.Add("pre_msg_time", "1693918697")
	reqURL := GetHistoryUrl + "?" + values.Encode()
	req, _ := http.NewRequest("GET", reqURL, nil)

	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	assert.Equal(t, http.StatusOK, response.Code)
	// 测试失败查询（to_user_id==from_user_id）
	token, err = utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}
	values = url.Values{}
	values.Set("to_user_id", "1")
	reqURL = GetHistoryUrl + "?" + values.Encode()
	req, _ = http.NewRequest("GET", reqURL, nil)

	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	// 测试失败查询（to_user_id==0）
	token, err = utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}
	values = url.Values{}
	values.Set("to_user_id", "0")
	reqURL = GetHistoryUrl + "?" + values.Encode()
	req, _ = http.NewRequest("GET", reqURL, nil)

	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}
