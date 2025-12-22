package outbox

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/n-creativesystem/outbox-worker/pkg/internal/tests"
)

var (
	mysqlProvider    tests.DBProvider
	postgresProvider tests.DBProvider
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pc, err := tests.RunPostgreSQLContainer(ctx)
	if err != nil {
		fmt.Printf("failed setup postgres container: %v\n", err)
		os.Exit(1)
	}

	mc, err := tests.RunMySQLContainer(ctx)
	if err != nil {
		fmt.Printf("failed setup mysql container: %v\n", err)
		os.Exit(1)
	}

	mysqlProvider = tests.NewMySQLProvider(mc)
	postgresProvider = tests.NewPostgresProvider(pc)

	exitCode := m.Run()

	if err := mc.Terminate(ctx); err != nil {
		fmt.Printf("failed to terminate container: %v\n", err)
	}
	if err := pc.Terminate(ctx); err != nil {
		fmt.Printf("failed to terminate container: %v\n", err)
	}

	os.Exit(exitCode)
}
