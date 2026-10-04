//go:build integration

package integration

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	integrationUserID    = "60000000-0000-4000-8000-000000000001"
	integrationJWTSecret = "integration-access-secret-with-at-least-32-bytes"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestFrontendFacingTransactionPath(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Fatal("Docker is required for the frontend-path integration test")
	}

	repositoryRoot := findRepositoryRoot(t)
	containerName := fmt.Sprintf("finory-integration-%d-%d", os.Getpid(), time.Now().UnixNano())
	run(t, repositoryRoot, nil, "docker", "run", "--rm", "-d", "--name", containerName,
		"-e", "POSTGRES_USER=finory", "-e", "POSTGRES_PASSWORD=finory", "-e", "POSTGRES_DB=postgres",
		"-p", "127.0.0.1::5432", "postgres:16-alpine")
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
	})

	databasePort := waitForPostgres(t, repositoryRoot, containerName)
	run(t, repositoryRoot, nil, "docker", "exec", containerName, "createdb", "-U", "finory", "auth_db")
	run(t, repositoryRoot, nil, "docker", "exec", containerName, "createdb", "-U", "finory", "ledger_db")
	applyMigrations(t, repositoryRoot, containerName)

	temporaryDirectory := t.TempDir()
	apiBinary := filepath.Join(temporaryDirectory, "finory-api")
	run(t, filepath.Join(repositoryRoot, "backend/service"), nil, "go", "build", "-o", apiBinary, "./cmd/api")

	apiPort := availablePort(t)
	apiURL := "http://127.0.0.1:" + apiPort
	stopAPI := startService(t, filepath.Join(repositoryRoot, "backend/service"), apiBinary, []string{
		"PORT=" + apiPort,
		"AUTH_DATABASE_URL=postgresql://finory:finory@127.0.0.1:" + databasePort + "/auth_db?sslmode=disable",
		"LEDGER_DATABASE_URL=postgresql://finory:finory@127.0.0.1:" + databasePort + "/ledger_db?sslmode=disable",
		"GOOGLE_CLIENT_ID=integration-client-id",
		"GOOGLE_CLIENT_SECRET=integration-client-secret",
		"GOOGLE_REDIRECT_URL=http://localhost:" + apiPort + "/api/v1/auth/google/callback",
		"APP_REDIRECT_URL=http://localhost:3000/dashboard",
		"CORS_ALLOWED_ORIGINS=http://localhost:3000",
		"JWT_ACCESS_SECRET=" + integrationJWTSecret,
	})
	t.Cleanup(stopAPI)
	waitForHealth(t, apiURL+"/health")

	accessToken := signedAccessToken(t, integrationUserID)
	category := postJSON(t, apiURL+"/api/v1/categories", accessToken, map[string]string{
		"name": "Integration Expense", "type": "expense",
	})
	wallet := postJSON(t, apiURL+"/api/v1/wallets", accessToken, map[string]string{
		"name": "Integration Bank", "type": "bank", "balance": "100.00", "currencyCode": "THB",
	})
	transaction := postJSON(t, apiURL+"/api/v1/transactions", accessToken, map[string]string{
		"categoryId": category.ID, "walletId": wallet.ID, "type": "expense",
		"amount": "42.50", "description": "frontend-path-integration", "transactionDate": "2026-09-27",
	})

	request, err := http.NewRequest(http.MethodGet, apiURL+"/api/v1/transactions/"+transaction.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		t.Fatalf("read transaction through API: %v", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !bytes.Contains(responseBody, []byte(`"id":"`+transaction.ID+`"`)) {
		t.Fatalf("API read response = %d %s", response.StatusCode, responseBody)
	}

	if !uuidPattern.MatchString(transaction.ID) {
		t.Fatalf("transaction ID is not a UUID: %q", transaction.ID)
	}
	query := fmt.Sprintf(
		"SELECT user_id::text || '|' || amount::text || '|' || description FROM transactions WHERE id = '%s'::uuid;",
		transaction.ID,
	)
	databaseRow := strings.TrimSpace(run(t, repositoryRoot, nil, "docker", "exec", containerName,
		"psql", "-U", "finory", "-d", "ledger_db", "-tA", "-c", query))
	if databaseRow != integrationUserID+"|42.5000|frontend-path-integration" {
		t.Fatalf("database row = %q", databaseRow)
	}
}

type createdResource struct {
	ID string `json:"id"`
}

func postJSON(t *testing.T, endpoint, accessToken string, payload map[string]string) createdResource {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		t.Fatalf("POST %s: %v", endpoint, err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("POST %s = %d %s", endpoint, response.StatusCode, responseBody)
	}
	var resource createdResource
	if err := json.Unmarshal(responseBody, &resource); err != nil || resource.ID == "" {
		t.Fatalf("decode POST %s response %s: %v", endpoint, responseBody, err)
	}
	return resource
}

func signedAccessToken(t *testing.T, userID string) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{"sub": userID, "exp": time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	signature := hmac.New(sha256.New, []byte(integrationJWTSecret))
	_, _ = signature.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil))
}

func findRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate integration test")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../.."))
}

func waitForPostgres(t *testing.T, workingDirectory, containerName string) string {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	var port string
	for time.Now().Before(deadline) {
		output, err := command(workingDirectory, nil, "docker", "port", containerName, "5432/tcp")
		if err == nil {
			_, candidate, splitErr := net.SplitHostPort(strings.TrimSpace(output))
			if splitErr == nil {
				port = candidate
			}
		}
		if port != "" {
			if _, err := command(workingDirectory, nil, "docker", "exec", containerName, "pg_isready", "-U", "finory", "-d", "postgres"); err == nil {
				return port
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatal("PostgreSQL did not become ready")
	return ""
}

func applyMigrations(t *testing.T, repositoryRoot, containerName string) {
	t.Helper()
	sets := []struct {
		database  string
		directory string
		files     []string
	}{
		{database: "auth_db", directory: "auth", files: []string{
			"00001_create_users.sql", "00002_create_refresh_sessions.sql",
		}},
		{database: "ledger_db", directory: "ledger", files: []string{
			"00001_create_categories_and_wallets.sql", "00002_create_transactions.sql", "00003_create_budgets.sql",
		}},
	}
	for _, set := range sets {
		for _, name := range set.files {
			path := filepath.Join(repositoryRoot, "backend/service/migrations", set.directory, name)
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			migration := string(contents)
			upStart := strings.Index(migration, "-- +goose Up")
			downStart := strings.Index(migration, "-- +goose Down")
			if upStart < 0 || downStart < 0 || downStart <= upStart {
				t.Fatalf("migration %s is missing Goose Up/Down markers", name)
			}
			upSQL := migration[upStart+len("-- +goose Up") : downStart]
			cmd := exec.Command("docker", "exec", "-i", containerName, "psql", "-v", "ON_ERROR_STOP=1", "-U", "finory", "-d", set.database)
			cmd.Dir = repositoryRoot
			cmd.Stdin = strings.NewReader(upSQL)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("apply %s migration %s: %v\n%s", set.database, name, err, output)
			}
		}
	}
}

func availablePort(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
}

func startService(t *testing.T, workingDirectory, binary string, environment []string) func() {
	t.Helper()
	cmd := exec.Command(binary)
	cmd.Dir = workingDirectory
	cmd.Env = append(os.Environ(), environment...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %s: %v", binary, err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	return func() {
		if cmd.Process == nil {
			return
		}
		_ = cmd.Process.Signal(os.Interrupt)
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	}
}

func waitForHealth(t *testing.T, endpoint string) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(endpoint)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("service did not become healthy at %s", endpoint)
}

func run(t *testing.T, workingDirectory string, environment []string, name string, arguments ...string) string {
	t.Helper()
	output, err := command(workingDirectory, environment, name, arguments...)
	if err != nil {
		t.Fatalf("run %s %s: %v\n%s", name, strings.Join(arguments, " "), err, output)
	}
	return output
}

func command(workingDirectory string, environment []string, name string, arguments ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arguments...)
	cmd.Dir = workingDirectory
	cmd.Env = append(os.Environ(), environment...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
