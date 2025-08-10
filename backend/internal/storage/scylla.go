package storage

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

// NewScyllaSession creates a gocql session using env or provided parameters.
// Env variables:
// - SCYLLA_HOSTS (comma separated, default: 127.0.0.1)
// - SCYLLA_PORT (default: 9042)
// - SCYLLA_KEYSPACE (default: holidays)
// - SCYLLA_USERNAME, SCYLLA_PASSWORD (optional)
func NewScyllaSession() (*gocql.Session, error) {
	hosts := getEnv("SCYLLA_HOSTS", "127.0.0.1")
	port := getEnvInt("SCYLLA_PORT", 9042)
	keyspace := getEnv("SCYLLA_KEYSPACE", "holidays")
	username := os.Getenv("SCYLLA_USERNAME")
	password := os.Getenv("SCYLLA_PASSWORD")
	consistencyEnv := strings.ToUpper(getEnv("SCYLLA_CONSISTENCY", "QUORUM"))
	dc := getEnv("SCYLLA_LOCAL_DC", "")
	numConns := getEnvInt("SCYLLA_NUM_CONNS", 4)

	cluster := gocql.NewCluster(strings.Split(hosts, ",")...)
	cluster.Port = port
	cluster.Keyspace = keyspace
	cons := map[string]gocql.Consistency{
		"ANY":          gocql.Any,
		"ONE":          gocql.One,
		"TWO":          gocql.Two,
		"THREE":        gocql.Three,
		"QUORUM":       gocql.Quorum,
		"ALL":          gocql.All,
		"LOCAL_QUORUM": gocql.LocalQuorum,
		"EACH_QUORUM":  gocql.EachQuorum,
		"LOCAL_ONE":    gocql.LocalOne,
	}[consistencyEnv]
	if cons == 0 && consistencyEnv != "ANY" { // fallback bei unbekanntem Wert
		cons = gocql.Quorum
	}
	cluster.Consistency = cons
	cluster.ProtoVersion = 4
	cluster.Timeout = 15 * time.Second
	cluster.ConnectTimeout = 15 * time.Second
	cluster.NumConns = numConns
	// Token-aware + optional DC-aware policy keeps requests close to data
	if dc != "" {
		cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.DCAwareRoundRobinPolicy(dc))
	} else {
		cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	}
	// Scylla-specific niceties helpful for containers/Cloud
	cluster.DisableInitialHostLookup = true
	cluster.IgnorePeerAddr = true
	// Hinweis: Shard-aware Port (19042) kann per SCYLLA_PORT konfiguriert werden
	cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{NumRetries: 5, Min: 200 * time.Millisecond, Max: 3 * time.Second}
	if username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{Username: username, Password: password}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("create scylla session: %w", err)
	}
	return session, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return def
}
