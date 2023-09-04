package user

import (
	"app/config"
	"app/modules/models"
	"app/utils"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
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

var RegisterUrl = "/douyin/user/register/"
var LoginUrl = "/douyin/user/login/"
var GetUserUrl = "/douyin/user/"
var db = utils.GetDb()
var jordanId uint
var testToken string

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
	jordanId = jordan.ID
}

// 测试用户登录
func TestLogin(t *testing.T) {
	config.Router.POST(LoginUrl, Login) // 将登录API函数注册给登录URL

	// 测试成功登录
	values := url.Values{}
	values.Add("username", "jordan")
	values.Add("password", "jordan_pass")
	reqURL := LoginUrl + "?" + values.Encode()

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	var responseJson map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &responseJson)
	// 断言的格式是 assert.Equal(t, 期望值, 实际值)
	assert.Equal(t, http.StatusOK, response.Code) // 测试成功登录
	// 测试返回的JSON内容
	assert.Equal(t, 0, int(responseJson["status_code"].(float64)))
	assert.Equal(t, jordanId, uint(responseJson["user_id"].(float64)))
	testToken = responseJson["token"].(string)

	// 测试空密码
	values.Set("password", "")
	reqURL = LoginUrl + "?" + values.Encode()

	req, _ = http.NewRequest("POST", reqURL, nil)
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// 测试不存在的用户
	values.Set("username", "wrong_user")
	values.Set("password", "wrong_pass")
	reqURL = LoginUrl + "?" + values.Encode()
	req, _ = http.NewRequest("POST", reqURL, nil)
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	// 测试正确的用户但是错误的密码
	values.Set("jordan", "wrong_pass")
	reqURL = LoginUrl + "?" + values.Encode()
	req, _ = http.NewRequest("POST", reqURL, nil)
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

// 测试用户注册
func TestUserRegister(t *testing.T) {
	config.Router.POST(RegisterUrl, Register)

	// 测试成功注册
	values := url.Values{}
	values.Add("username", "stephen")
	values.Add("password", "stephen_pass")
	reqURL := RegisterUrl + "?" + values.Encode()
	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusCreated, response.Code) // test successfully registered.

	// 测试钩子是否自动创建了对应的 UserProfile
	userProfile := models.UserProfile{UserID: jordanId}
	result := db.Find(&userProfile)
	assert.Equal(t, int64(1), result.RowsAffected)

	// 测试空用户名
	values.Set("username", "")
	reqURL = RegisterUrl + "?" + values.Encode()
	req, err = http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// 测试空密码
	values.Set("username", "stephen")
	values.Set("password", "")
	reqURL = RegisterUrl + "?" + values.Encode()
	req, err = http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// 测试长用户名
	values.Set("username", "this_is_a_super_long_user_name_with_more_than_25_chars")
	reqURL = RegisterUrl + "?" + values.Encode()
	req, err = http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// 测试长密码
	values.Set("username", "stephen")
	values.Set("password", "this_is_a_super_long_password_with_more_than_25_chars")
	reqURL = RegisterUrl + "?" + values.Encode()
	req, err = http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

// TODO: Test GetUser()
// 1. Test user not found. 2. test invalid token (a. invalid token. b. valid token but expired)
// 3. Test get a valid user info
func TestGetUser(t *testing.T) {
	config.Router.GET(GetUserUrl, GetUser)

	// 测试成功获取用户信息
	values := url.Values{}
	values.Add("user_id", strconv.Itoa(int(jordanId)))
	values.Add("token", testToken)
	reqURL := GetUserUrl + "?" + values.Encode()
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)

	// 测试一个非法用户ID
	values.Set("user_id", "invalid_id")
	reqURL = GetUserUrl + "?" + values.Encode()
	req, err = http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// 测试查询一个不存在的用户
	values.Set("user_id", "987654321")
	reqURL = GetUserUrl + "?" + values.Encode()
	req, err = http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusNotFound, response.Code)
}
