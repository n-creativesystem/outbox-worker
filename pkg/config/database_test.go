package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectDB(t *testing.T) {
	conf := Database{
		URI: "mysql://admin:pass1234@tcp(db:3306)/outbox?parseTime=true",
	}
	dsn, err := conf.Build()
	require.Nil(t, err)
	require.Equal(t, "admin:pass1234@tcp(db:3306)/outbox?parseTime=true", dsn)
}

// func TestDatabaseValidation(t *testing.T) {
// 	conf := Database{}
// 	err := Validate(&conf)
// 	require.Equal(t,
// 		"host: cannot be blank; name: cannot be blank; password: cannot be blank; port: cannot be blank; username: cannot be blank.",
// 		err.Error(),
// 	)
// }

func TestTLS(t *testing.T) {
	tls := TLS{
		InsecureSkipVerify: false,
		SeverName:          "",
	}
	conf, err := tls.Config()
	require.NoError(t, err)
	require.False(t, conf.InsecureSkipVerify)
	require.Empty(t, conf.ServerName)
	tls = TLS{
		InsecureSkipVerify: true,
		SeverName:          "aaa",
	}
	conf, err = tls.Config()
	require.NoError(t, err)
	require.True(t, conf.InsecureSkipVerify)
	require.Equal(t, conf.ServerName, "aaa")
}
