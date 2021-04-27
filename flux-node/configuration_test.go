package flux

import (
	"fmt"
	"github.com/spf13/viper"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseDynamicKey(t *testing.T) {
	cases := []struct {
		pattern      string
		expectedKey  string
		expectedDef  string
		expectedType int
	}{
		{
			pattern:     "",
			expectedKey: "",
			expectedDef: "",
		},
		{
			pattern:     "${}",
			expectedKey: "${}",
			expectedDef: "",
		},
		{
			pattern:     "${",
			expectedKey: "${",
			expectedDef: "",
		},
		{
			pattern:     "username",
			expectedKey: "username",
			expectedDef: "",
		},
		{
			pattern:      "${user}",
			expectedKey:  "user",
			expectedDef:  "",
			expectedType: DynamicTypeConfig,
		},
		{
			pattern:      "${   user }",
			expectedKey:  "user",
			expectedDef:  "",
			expectedType: DynamicTypeConfig,
		},
		{
			pattern:      "${user:yongjia}",
			expectedKey:  "user",
			expectedDef:  "yongjia",
			expectedType: DynamicTypeConfig,
		},
		{
			pattern:      "${     user:yongjia  }",
			expectedKey:  "user",
			expectedDef:  "yongjia",
			expectedType: DynamicTypeConfig,
		},
		{
			pattern:      "${     address:http://bytepowered.net:8080  }",
			expectedKey:  "address",
			expectedDef:  "http://bytepowered.net:8080",
			expectedType: DynamicTypeConfig,
		},
		{
			pattern:      "${     host:  }",
			expectedKey:  "host",
			expectedDef:  "",
			expectedType: DynamicTypeConfig,
		},
		{
			pattern:      "#{     DEPLOY_ENV:  }",
			expectedKey:  "DEPLOY_ENV",
			expectedDef:  "",
			expectedType: DynamicTypeEnv,
		},
		{
			pattern:      "#{     DEPLOY_ENV:2020  }",
			expectedKey:  "DEPLOY_ENV",
			expectedDef:  "2020",
			expectedType: DynamicTypeEnv,
		},
	}
	assert := assert2.New(t)
	for _, tcase := range cases {
		key, def, typ := ParseDynamicKey(tcase.pattern)
		assert.Equal(tcase.expectedKey, key)
		assert.Equal(tcase.expectedDef, def)
		assert.Equal(tcase.expectedType, typ)
	}
}

func TestConfiguration_GetDynamic(t *testing.T) {
	viper.Set("username", "chen")
	viper.Set("user.year", 2020)
	cases := []struct {
		config   *Configuration
		lookup   string
		expected interface{}
	}{
		{
			config:   NewGlobalConfig(),
			lookup:   "myuserid",
			expected: nil,
		},
		{
			config: func() *Configuration {
				c := NewGlobalConfig()
				c.Set("userE", "chen")
				return c
			}(),
			lookup:   "userE",
			expected: "chen",
		},
		{
			config: func() *Configuration {
				c := NewGlobalConfig()
				c.Set("userX", "${username}")
				return c
			}(),
			lookup:   "userX",
			expected: "chen",
		},
		{
			config: func() *Configuration {
				c := NewGlobalConfig()
				c.Set("usernameA", "${usernameX:haha}")
				return c
			}(),
			lookup:   "usernameA",
			expected: "haha",
		},
		{
			config: func() *Configuration {
				c := NewGlobalConfig()
				c.Set("userA", "${user.year}")
				return c
			}(),
			lookup:   "userA",
			expected: 2020,
		},
	}
	assert := assert2.New(t)
	for _, tcase := range cases {
		assert.Equal(tcase.expected, tcase.config.Get(tcase.lookup))
	}
}

func TestConfiguration_GetStruct(t *testing.T) {
	viper.Set("app.user.name", "chen")
	viper.Set("app.user.year", 2020)
	viper.Set("app.user.id", "yongjiapro")
	viper.Set("app.profile", "yongjiapro")
	app := NewConfiguration("app")
	user := struct {
		Name string `json:"name"`
		Year int    `json:"year"`
		Id   string `json:"id"`
	}{}
	assert := assert2.New(t)
	err := app.GetStruct("user", &user)
	assert.Nil(err)
	assert.Equal(user.Name, "chen")
	assert.Equal(user.Year, 2020)
	assert.Equal(user.Id, "yongjiapro")
}

func TestConfiguration_Keys(t *testing.T) {
	viper.Set("app.year", 2020)
	viper.Set("app.profile", "yongjiapro")
	viper.Set("app.id", "yongjiapro")
	app := NewConfiguration("app")
	assert := assert2.New(t)
	keys := app.Keys()
	assert.Contains(keys, "year")
	assert.Contains(keys, "profile")
}

func TestConfiguration_CircleKey(t *testing.T) {
	viper.Set("app.year", "${app.year:2020}")
	app := NewConfiguration("app")
	assert := assert2.New(t)
	assert.Equal("2020", app.GetString("year"))
}

func TestConfiguration_WatchKey(t *testing.T) {
	viper.Set("app.year", "2020")
	app := NewConfiguration("app")
	app.StartWatch(func(key string, value interface{}) {
		fmt.Printf("key=%s, value=%d \n", key, value)
	})
	go func() {
		for i := 0; i < 10; i++ {
			<-time.After(time.Second)
			app.Set("year", i)
		}
	}()
	<-time.After(time.Second * 10)
	app.StopWatch()
}

func NewGlobalConfig() *Configuration {
	return &Configuration{nspath: "", registry: viper.GetViper()}
}
