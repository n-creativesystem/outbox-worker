package outbox

import (
	"testing"

	"github.com/n-creativesystem/outbox-worker/pkg/internal/tests"
	"github.com/stretchr/testify/require"
)

// testProviders は全てのテスト対象DBプロバイダーを返す
func testProviders() []tests.DBProvider {
	providers := []tests.DBProvider{}

	// if mysqlProvider != nil {
	// 	providers = append(providers, mysqlProvider)
	// }

	if postgresProvider != nil {
		providers = append(providers, postgresProvider)
	}

	return providers
}

// newTestClientForProvider は指定されたDBプロバイダーでクライアントを作成します
func newTestClientForProvider(t *testing.T, provider tests.DBProvider) *Client {
	t.Helper()

	// DBプロバイダーからテスト用DBと設定を取得
	db, cfg := provider.CreateTestDB(t)

	// スキーマを作成
	provider.CreateSchema(t, db)

	ctx := t.Context()
	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)

	return client
}

// newTestClient はデフォルト（MySQL）のテストクライアントを作成します
// 後方互換性のために残していますが、新しいテストではtestProvidersを使用してください
func newTestClient(t *testing.T) *Client {
	t.Helper()
	return newTestClientForProvider(t, mysqlProvider)
}

// insertTestData はテストデータを挿入するヘルパー関数
// rows: 各行のデータ（カラム順: id, aggregate_type, aggregate_id, event, payload, retry_at, retry_count, sent_at）
func insertTestData(t *testing.T, provider tests.DBProvider, client *Client, rows [][]interface{}) {
	t.Helper()
	provider.InsertTestData(t, client.db.DB, rows)
}
