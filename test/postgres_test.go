package postgres_test

import (
	"authService/internal/config"
	"authService/internal/entities"
	"authService/internal/repository/postgres"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupRepo(t *testing.T) (*postgres.Repository, func()) {
	t.Helper()

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Postgres: config.PostgresConfig{
			Host:     "localhost",
			Port:     "6132",
			User:     "postgres",
			Password: "n164838i",
			DBName:   "db_auth",
			SSLMode:  "allow",
		},
	}

	repo, err := postgres.NewRepository(logger, cfg, context.Background())
	require.NoError(t, err)

	err = repo.OnStart(context.Background())
	require.NoError(t, err)

	tearDown := func() {
		require.NoError(t, repo.OnStop(context.Background()))
	}

	return repo, tearDown
}

func createTestUser(t *testing.T, repo *postgres.Repository) *entities.User {
	t.Helper()
	user := &entities.User{
		TelegramID: 999999999,
		FirstName:  "TestUser",
	}
	created, err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err)
	return created
}

func deleteUserByID(t *testing.T, repo *postgres.Repository, id int32) {
	t.Helper()
	_, err := repo.DB.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	require.NoError(t, err)
}

func TestUserCreate(t *testing.T) {
	repo, tearDown := setupRepo(t)
	defer tearDown()

	user := createTestUser(t, repo)
	defer deleteUserByID(t, repo, user.ID)

	assert.NotZero(t, user.ID)
	assert.Equal(t, int64(999999999), user.TelegramID)
	assert.Equal(t, "TestUser", user.FirstName)
}

func TestUserGetByTelegramId(t *testing.T) {
	repo, tearDown := setupRepo(t)
	defer tearDown()

	user := createTestUser(t, repo)
	defer deleteUserByID(t, repo, user.ID)

	foundID, err := repo.GetUserByTelegramId(context.Background(), user)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundID)
}

func TestUserUpdateTokens(t *testing.T) {
	repo, tearDown := setupRepo(t)
	defer tearDown()

	user := createTestUser(t, repo)
	defer deleteUserByID(t, repo, user.ID)

	user.AccessToken = "test-access-token"
	user.RefreshToken = "test-refresh-token"

	err := repo.UpdateTokens(context.Background(), user)
	require.NoError(t, err)

	updated, err := repo.GetUserById(context.Background(), user.ID)
	require.NoError(t, err)

	assert.Equal(t, "test-access-token", updated.AccessToken)
	assert.Equal(t, "test-refresh-token", updated.RefreshToken)
}

func TestUserGetById(t *testing.T) {
	repo, tearDown := setupRepo(t)
	defer tearDown()

	user := createTestUser(t, repo)
	defer deleteUserByID(t, repo, user.ID)

	foundUser, err := repo.GetUserById(context.Background(), user.ID)
	require.NoError(t, err)

	assert.Equal(t, user.TelegramID, foundUser.TelegramID)
	assert.Equal(t, user.FirstName, foundUser.FirstName)
}
