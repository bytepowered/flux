package flux

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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
			expectedType: DynamicTypeLookupConfig,
		},
		{
			pattern:      "${   user }",
			expectedKey:  "user",
			expectedDef:  "",
			expectedType: DynamicTypeLookupConfig,
		},
		{
			pattern:      "${user:yongjia}",
			expectedKey:  "user",
			expectedDef:  "yongjia",
			expectedType: DynamicTypeLookupConfig,
		},
		{
			pattern:      "${     user:yongjia  }",
			expectedKey:  "user",
			expectedDef:  "yongjia",
			expectedType: DynamicTypeLookupConfig,
		},
		{
			pattern:      "${     address:http://bytepowered.net:8080  }",
			expectedKey:  "address",
			expectedDef:  "http://bytepowered.net:8080",
			expectedType: DynamicTypeLookupConfig,
		},
		{
			pattern:      "${     host:  }",
			expectedKey:  "host",
			expectedDef:  "",
			expectedType: DynamicTypeLookupConfig,
		},
		{
			pattern:      "#{     DEPLOY_ENV:  }",
			expectedKey:  "DEPLOY_ENV",
			expectedDef:  "",
			expectedType: DynamicTypeLookupEnv,
		},
		{
			pattern:      "#{     DEPLOY_ENV:2020  }",
			expectedKey:  "DEPLOY_ENV",
			expectedDef:  "2020",
			expectedType: DynamicTypeLookupEnv,
		},
	}
	for _, tcase := range cases {
		key, def, typ := ParseDynamicKey(tcase.pattern)
		assert.Equal(t, tcase.expectedKey, key)
		assert.Equal(t, tcase.expectedDef, def)
		assert.Equal(t, tcase.expectedType, typ)
	}
}

func TestConfiguration_GetDynamic(t *testing.T) {
	viper.Reset()
	viper.Set("username", "chen")
	viper.Set("user.year", 2020)
	cases := []struct {
		config   *Configuration
		lookup   string
		expected interface{}
	}{
		{
			config:   NewRootConfiguration(),
			lookup:   "myuserid",
			expected: nil,
		},
		{
			config: func() *Configuration {
				c := NewRootConfiguration()
				c.Set("userE", "chen")
				return c
			}(),
			lookup:   "userE",
			expected: "chen",
		},
		{
			config: func() *Configuration {
				c := NewRootConfiguration()
				c.Set("userX", "${username}")
				return c
			}(),
			lookup:   "userX",
			expected: "chen",
		},
		{
			config: func() *Configuration {
				c := NewRootConfiguration()
				c.Set("usernameA", "${usernameX:haha}")
				return c
			}(),
			lookup:   "usernameA",
			expected: "haha",
		},
		{
			config: func() *Configuration {
				c := NewRootConfiguration()
				c.Set("userA", "${user.year}")
				return c
			}(),
			lookup:   "userA",
			expected: 2020,
		},
	}
	for _, tcase := range cases {
		assert.Equal(t, tcase.expected, tcase.config.Get(tcase.lookup))
	}
}

func TestConfiguration_GetStruct(t *testing.T) {
	viper.Reset()
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
	err := app.GetStruct("user", &user)
	assert.Nil(t, err)
	assert.Equal(t, user.Name, "chen")
	assert.Equal(t, user.Year, 2020)
	assert.Equal(t, user.Id, "yongjiapro")
}

func TestConfiguration_Keys(t *testing.T) {
	viper.Reset()
	viper.Set("app.year", 2020)
	viper.Set("app.profile", "yongjiapro")
	viper.Set("app.id", "yongjiapro")
	app := NewConfiguration("app")
	keys := app.Keys()
	assert.Contains(t, keys, "year")
	assert.Contains(t, keys, "profile")
}

func TestConfiguration_CircleKey(t *testing.T) {
	viper.Reset()
	viper.Set("app.year", "${app.year:2020}")
	app := NewConfiguration("app")
	assert.Equal(t, "2020", app.GetString("year"))
}

func TestConfiguration_WatchKey(t *testing.T) {
	viper.Reset()
	viper.Set("app.year", "2020")
	app := NewConfiguration("app")
	app.StartWatch(func(key string, value interface{}) {
		fmt.Printf("key=%s, value=%d \n", key, value)
	})
	go func() {
		for i := 0; i < 3; i++ {
			<-time.After(time.Second)
			app.Set("year", i)
		}
	}()
	<-time.After(time.Second * 10)
	app.StopWatch()
}

func TestConfiguration_RootNamespace(t *testing.T) {
	viper.Reset()
	viper.Set("app.year", "${app.year:2020}")
	viper.Set("app.no", "9999")
	root := NewRootConfiguration()
	assert.Equal(t, map[string]interface{}{
		"app": map[string]interface{}{
			"no":   "9999",
			"year": "${app.year:2020}",
		},
	}, root.ToStringMap())
	confs := root.ToConfigurations()
	assert.Equal(t, 1, len(confs))
	assert.Equal(t, "2020", confs[0].GetString("app.year"))
}
