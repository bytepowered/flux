package flux

import (
	"github.com/spf13/viper"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
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
			config:   NewConfigurationOfNS("empty"),
			lookup:   "user",
			expected: nil,
		},
		{
			config: func() *Configuration {
				c := NewConfigurationOfNS("empty")
				c.Set("user", "chen")
				return c
			}(),
			lookup:   "user",
			expected: "chen",
		},
		{
			config: func() *Configuration {
				c := NewConfigurationOfNS("empty")
				c.Set("user", "${username}")
				return c
			}(),
			lookup:   "user",
			expected: "chen",
		},
		{
			config: func() *Configuration {
				c := NewConfigurationOfNS("empty")
				c.Set("user", "${usernameX:haha}")
				return c
			}(),
			lookup:   "user",
			expected: "haha",
		},
		{
			config: func() *Configuration {
				c := NewConfigurationOfNS("empty")
				c.Set("user", "${user.year}")
				return c
			}(),
			lookup:   "user",
			expected: 2020,
		},
	}
	assert := assert2.New(t)
	for _, tcase := range cases {
		assert.Equal(tcase.expected, tcase.config.Get(tcase.lookup))
	}
}
