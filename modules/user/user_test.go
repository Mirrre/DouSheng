package user

import (
	"app/config"
	"app/modules/models"
	"app/util"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

// Golang的测试会检测到每个包的TestMain函数，首先执行它
func TestMain(m *testing.M) {
	util.Setup()
	postSetup()
	code := m.Run()
	util.Teardown()
	os.Exit(code)
}

var RegisterUrl = "/douyin/user/register/"
var LoginUrl = "/douyin/user/login/"
var db = util.GetDb()

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

// 测试用户登录
func TestLogin(t *testing.T) {
	config.Router.POST(LoginUrl, Login) // 将登录API函数注册给登录URL
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

	// 断言的格式是 assert.Equal(t, 期望值, 实际值)
	assert.Equal(t, http.StatusOK, response.Code) // 测试成功登录

	// TODO: 1. Test empty password. 2. Test invalid username. 3. test invalid password.
	// 4. test all JSON field returned as expected.
}

// 测试用户注册
func TestUserRegister(t *testing.T) {
	config.Router.POST(RegisterUrl, Register)
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

	// TODO: 1. Test empty username or password. 2. Test username or password length not satisafied.
}

// TODO: Test GetUser()
// 1. Test user not found. 2. test invalid token (a. invalid token. b. valid token but expired)
// 3. Test get a valid user info
