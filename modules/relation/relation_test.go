package relation

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

func TestMain(m *testing.M) {
	utils.Setup()
	postSetup()
	code := m.Run()
	utils.Teardown()
	os.Exit(code)
}

var ActionUrl = "/douyin/relation/action/"              //关注操作
var FollowingListUrl = "/douyin/relation/follow/list/"  //关注列表
var FollowerListUrl = "/douyin/relation/follower/list/" //粉丝列表
var FriendListRul = "/douyin/relation/friend/list/"     //好友列表
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

// 测试关注操作
func TestAction(t *testing.T) {
	config.Router.POST(ActionUrl, Action)
	// 关注：发起关注操作的id是 1 ，被关注的id是 2 ，结果返回200
	token, err := utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}
	values := url.Values{}
	values.Add("token", token)
	values.Add("to_user_id", "2")
	values.Add("action_type", "1")
	reqURL := ActionUrl + "?" + values.Encode()

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	// 断言
	assert.Equal(t, http.StatusOK, response.Code)

	// 关注：发起关注操作的id是 1 ，被关注的id是 1 ，错误操作，结果返回400
	token1, err := utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}

	values.Set("token", token1)
	values.Set("to_user_id", "1")
	values.Set("action_type", "1")
	reqURL = ActionUrl + "?" + values.Encode()

	req, err = http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	// 断言
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// 取关：发起取关操作的id是 1 ，被取关的id是 2 ，结果返回200
	token2, err := utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}

	values.Set("token", token2)
	values.Set("to_user_id", "2")
	values.Set("action_type", "2")
	reqURL = ActionUrl + "?" + values.Encode()

	req, err = http.NewRequest("POST", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)

	// 断言
	assert.Equal(t, http.StatusOK, response.Code)
}

// 测试获取关注列表
func TestGetFollowings(t *testing.T) {
	config.Router.GET(FollowingListUrl, GetFollowings)
	// 获取id为1的关注列表，结果返回200
	token, err := utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}
	values := url.Values{}
	values.Add("token", token)
	values.Add("user_id", "1")
	reqURL := FollowingListUrl + "?" + values.Encode()
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)

	// 获取id为0的关注列表，结果返回400
	values.Set("user_id", "0")
	reqURL = FollowingListUrl + "?" + values.Encode()
	req, err = http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

}

// 测试获取粉丝列表
func TestGetFollowers(t *testing.T) {
	config.Router.GET(FollowerListUrl, GetFollowers)
	// 获取id为1的粉丝列表，结果返回200
	token, err := utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}
	values := url.Values{}
	values.Add("token", token)
	values.Add("user_id", "1")
	reqURL := FollowerListUrl + "?" + values.Encode()
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)

	// 获取id为0的粉丝列表，结果返回400
	values.Set("user_id", "0")
	reqURL = FollowerListUrl + "?" + values.Encode()
	req, err = http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

// 测试获取好友列表
func TestGetFriends(t *testing.T) {
	config.Router.GET(FriendListRul, GetFollowers)
	// 获取id为1的好友列表，结果返回200
	token, err := utils.GenerateToken(1)
	if err != nil {
		t.Fatal(err)
	}
	values := url.Values{}
	values.Add("token", token)
	values.Add("user_id", "1")
	reqURL := FriendListRul + "?" + values.Encode()
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusOK, response.Code)

	// 获取id为0的好友列表，结果返回400
	values.Set("user_id", "0")
	reqURL = FriendListRul + "?" + values.Encode()
	req, err = http.NewRequest("GET", reqURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	config.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}
