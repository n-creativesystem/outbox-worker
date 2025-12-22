package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResource(t *testing.T) {
	resource := Resource{
		Whitelist:  []string{"a", "b", "c", "d"},
		Blacklist:  []string{"a", "x", "y", "z"},
		SkipEvents: []string{"a", "b", "c", "d"},
	}
	require := require.New(t)
	require.ElementsMatch(resource.GetWhitelist(), []string{"b", "c", "d"})
	require.ElementsMatch(resource.GetBlacklist(), []string{"a", "x", "y", "z"})
	require.ElementsMatch(resource.GetSkipEvents(), []string{"a", "b", "c", "d"})
}

func TestIsIgnoreFetchErrorResponses(t *testing.T) {
	resource := Resource{
		IgnoreFetchErrorResources: []string{"a", "b", "c"},
	}
	require := require.New(t)
	require.True(resource.IsIgnoreFetchErrorResources("a"))
	require.True(resource.IsIgnoreFetchErrorResources("b"))
	require.True(resource.IsIgnoreFetchErrorResources("c"))
	require.False(resource.IsIgnoreFetchErrorResources("d"))
}
