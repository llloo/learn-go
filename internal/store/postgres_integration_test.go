//go:build integration

package store

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// pgStore 是整个包共享的 store 实例，由 TestMain 初始化
var pgStore *PostgresStore

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 启动 postgres 容器
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer testcontainers.TerminateContainer(container)

	// 获取连接 URL
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	// 初始化 store（同时验证连接）
	pgStore, err = NewPostgresStore(dsn)
	if err != nil {
		panic("failed to create postgres store: " + err.Error())
	}

	// 手动建表（不用 golang-migrate，保持测试简单）
	_, err = pgStore.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			id         SERIAL PRIMARY KEY,
			title      TEXT NOT NULL,
			completed  BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		panic("failed to create table: " + err.Error())
	}

	// 运行所有 Test* 函数
	code := m.Run()
	os.Exit(code)
}

// cleanTable 在每个测试前清空表，保证测试隔离
func cleanTable(t *testing.T) {
	t.Helper()
	_, err := pgStore.db.ExecContext(context.Background(), "TRUNCATE TABLE tasks RESTART IDENTITY")
	if err != nil {
		t.Fatalf("failed to clean table: %v", err)
	}
}

func TestPostgresCreate(t *testing.T) {
	cleanTable(t)
	ctx := context.Background()

	task, err := pgStore.Create(ctx, "learn Go")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if task.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if task.Title != "learn Go" {
		t.Errorf("expected title 'learn Go', got %q", task.Title)
	}
	if task.Completed {
		t.Error("expected Completed to be false")
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestPostgresGetByID(t *testing.T) {
	cleanTable(t)
	ctx := context.Background()

	created, err := pgStore.Create(ctx, "test task")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := pgStore.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, got.ID)
	}
	if got.Title != created.Title {
		t.Errorf("expected title %q, got %q", created.Title, got.Title)
	}
}

func TestPostgresGetByID_NotFound(t *testing.T) {
	cleanTable(t)
	ctx := context.Background()

	_, err := pgStore.GetByID(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresPartialUpdate(t *testing.T) {
	cleanTable(t)
	ctx := context.Background()

	created, _ := pgStore.Create(ctx, "original title")

	newTitle := "updated title"
	updated, err := pgStore.PartialUpdate(ctx, created.ID, &newTitle, nil)
	if err != nil {
		t.Fatalf("PartialUpdate failed: %v", err)
	}

	// title 应该更新
	if updated.Title != "updated title" {
		t.Errorf("expected title 'updated title', got %q", updated.Title)
	}
	// completed 未传 nil，应该保持原值 false
	if updated.Completed {
		t.Error("expected Completed to remain false")
	}
}

func TestPostgresDelete(t *testing.T) {
	cleanTable(t)
	ctx := context.Background()

	created, _ := pgStore.Create(ctx, "to be deleted")

	if err := pgStore.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 删除后应该找不到
	_, err := pgStore.GetByID(ctx, created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
