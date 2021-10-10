package flagconfig

import (
	"flag"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldsAsFlags(t *testing.T) {
	type Config struct {
		Width, Height int
	}
	config := createAndParse(t, new(Config), "-width=20 -height 30").(*Config)
	assert.Equal(t, 20, config.Width)
	assert.Equal(t, 30, config.Height)
}

func TestFieldsAsFlags_Defaults(t *testing.T) {
	type Config struct {
		Width  int `default:"25"`
		Height int `default:"21"`
	}
	config := createAndParse(t, new(Config), "").(*Config)
	assert.Equal(t, 25, config.Width)
	assert.Equal(t, 21, config.Height)
}

func TestTextMarshallerSupported(t *testing.T) {
	type Config struct {
		Time time.Time
	}
	config := createAndParse(t, new(Config), "-time 2006-01-02T15:04:00-07:00").(*Config)
	expected, err := time.Parse(time.RFC822Z, "02 Jan 06 15:04 -0700")
	require.NoError(t, err)
	assert.Equal(t, expected, config.Time)
}

func TestBinaryMarshallerSupported(t *testing.T) {
	type Config struct {
		URL *url.URL
	}
	rawURL := "http://localhost/path/to"
	config := createAndParse(t, new(Config), "-url "+rawURL).(*Config)
	expected, err := url.Parse(rawURL)
	require.NoError(t, err)
	assert.Equal(t, expected, config.URL)
}

func TestNameOverriding(t *testing.T) {
	type Config struct {
		Name string `name:"alter"`
	}
	expected := "flag_value"
	config := createAndParse(t, new(Config), "-alter "+expected).(*Config)
	assert.Equal(t, expected, config.Name)
}

func TestIgnoreField(t *testing.T) {
	type Config struct {
		Tag     string `ignored:"true"`
		private string
	}
	require.NotPanics(t, func() {
		flags, err := MakeFlags(new(Config), "", flag.ContinueOnError)
		assert.NoError(t, err)
		assert.Nil(t, flags.Lookup("tag"))
		assert.Nil(t, flags.Lookup("private"))
	})
}

func TestLateBinding(t *testing.T) {
	type Config struct {
		Sub interface{}
	}
	t.Run("struct", func(t *testing.T) {
		type SubConfig struct {
			Field int
		}
		subConfig := new(SubConfig)
		config := &Config{subConfig}
		flags, err := MakeFlags(config, "", flag.ContinueOnError)
		require.NoError(t, err)
		require.NotNil(t, flags.Lookup("sub.field"))
		require.NoError(t, parse(flags, "-sub.field 42"))
		assert.Equal(t, 42, subConfig.Field)
	})
	t.Run("int", func(t *testing.T) {
		subConfig := new(int)
		config := &Config{subConfig}
		flags, err := MakeFlags(config, "", flag.ContinueOnError)
		require.NoError(t, err)
		require.NotNil(t, flags.Lookup("sub"))
		require.NoError(t, parse(flags, "-sub 42"))
		assert.Equal(t, 42, *subConfig)
	})
}

func createAndParse(t *testing.T, config interface{}, args string) interface{} {
	flags, err := MakeFlags(config, "", flag.ContinueOnError)
	require.NoError(t, err)
	require.NoError(t, parse(flags, args))
	return config
}

func parse(flags *flag.FlagSet, raw string) error {
	return flags.Parse(strings.Split(raw, " "))
}
