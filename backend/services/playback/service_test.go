package playback_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"novastream/config"
	"novastream/internal/database"
	"novastream/internal/integration"
	"novastream/internal/pool"
	"novastream/services/playback"
)

type stubMetadataService struct {
	files   map[string][]string
	subdirs map[string][]string
}

func newStubMetadataService() *stubMetadataService {
	return &stubMetadataService{
		files:   make(map[string][]string),
		subdirs: make(map[string][]string),
	}
}

func (s *stubMetadataService) ListDirectory(virtualPath string) ([]string, error) {
	return s.files[virtualPath], nil
}

func (s *stubMetadataService) ListSubdirectories(virtualPath string) ([]string, error) {
	return s.subdirs[virtualPath], nil
}

func setupPlaybackService(t *testing.T) (*playback.Service, *integration.NzbSystem, *stubMetadataService) {
	t.Helper()

	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.json")
	cfg := config.NewManager(settingsPath)
	if err := cfg.Save(config.DefaultSettings()); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	adapter := config.NewConfigAdapter(cfg)
	poolManager := pool.NewManager()
	nzbCfg := integration.NzbConfig{
		QueueDatabasePath:   filepath.Join(tempDir, "queue.db"),
		MetadataRootPath:    filepath.Join(tempDir, "metadata"),
		Password:            "",
		Salt:                "",
		MaxProcessorWorkers: 1,
		MaxDownloadWorkers:  1,
	}

	nzbSystem, err := integration.NewNzbSystem(nzbCfg, poolManager, adapter.GetConfigGetter())
	if err != nil {
		t.Fatalf("new nzb system: %v", err)
	}
	t.Cleanup(func() {
		_ = nzbSystem.Close()
	})

	if err := nzbSystem.StopService(context.Background()); err != nil {
		t.Fatalf("stop nzb service: %v", err)
	}

	metadataSvc := newStubMetadataService()
	service := playback.NewService(cfg, nil, nzbSystem, metadataSvc)

	return service, nzbSystem, metadataSvc
}

func TestQueueStatusQueued(t *testing.T) {
	service, nzbSystem, _ := setupPlaybackService(t)
	importerSvc := nzbSystem.ImporterService()

	fileSize := int64(1024)
	item := &database.ImportQueueItem{
		NzbPath:    "queued-item.nzb",
		Priority:   database.QueuePriorityNormal,
		Status:     database.QueueStatusPending,
		MaxRetries: 3,
		FileSize:   &fileSize,
	}

	if err := importerSvc.Database().Repository.AddToQueue(item); err != nil {
		t.Fatalf("add to queue: %v", err)
	}

	meta := `{"sourceNzbPath":"queued-item.nzb"}`
	if err := importerSvc.Database().Repository.UpdateMetadata(item.ID, &meta); err != nil {
		t.Fatalf("set metadata: %v", err)
	}

	status, err := service.QueueStatus(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("QueueStatus returned error: %v", err)
	}

	if status.QueueID != item.ID {
		t.Fatalf("expected queueID %d, got %d", item.ID, status.QueueID)
	}
	if status.HealthStatus != "queued" {
		t.Fatalf("expected healthStatus queued, got %q", status.HealthStatus)
	}
	if status.WebDAVPath != "" {
		t.Fatalf("expected empty webdav path, got %q", status.WebDAVPath)
	}
	if status.SourceNZBPath != "queued-item.nzb" {
		t.Fatalf("expected sourceNzbPath queued-item.nzb, got %q", status.SourceNZBPath)
	}
	if status.FileSize != fileSize {
		t.Fatalf("expected fileSize %d, got %d", fileSize, status.FileSize)
	}
}

func TestQueueStatusCompleted(t *testing.T) {
	service, nzbSystem, _ := setupPlaybackService(t)
	importerSvc := nzbSystem.ImporterService()

	fileSize := int64(2048)
	item := &database.ImportQueueItem{
		NzbPath:    "completed-item.nzb",
		Priority:   database.QueuePriorityNormal,
		Status:     database.QueueStatusPending,
		MaxRetries: 3,
		FileSize:   &fileSize,
	}

	if err := importerSvc.Database().Repository.AddToQueue(item); err != nil {
		t.Fatalf("add to queue: %v", err)
	}

	meta := `{"sourceNzbPath":"Series.S01E01.mkv","preflightHealth":"healthy"}`
	if err := importerSvc.Database().Repository.UpdateMetadata(item.ID, &meta); err != nil {
		t.Fatalf("set metadata: %v", err)
	}

	storagePath := "/virtual/Series.S01E01.mkv"
	if err := importerSvc.Database().Repository.AddStoragePath(item.ID, storagePath); err != nil {
		t.Fatalf("add storage path: %v", err)
	}

	if err := importerSvc.Database().Repository.UpdateQueueItemStatus(item.ID, database.QueueStatusCompleted, nil); err != nil {
		t.Fatalf("mark completed: %v", err)
	}

	status, err := service.QueueStatus(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("QueueStatus returned error: %v", err)
	}

	if status.WebDAVPath == "" {
		t.Fatalf("expected non-empty webdav path")
	}
	if status.HealthStatus != "healthy" {
		t.Fatalf("expected healthStatus healthy, got %q", status.HealthStatus)
	}
	if status.SourceNZBPath != "Series.S01E01.mkv" {
		t.Fatalf("expected sourceNzbPath Series.S01E01.mkv, got %q", status.SourceNZBPath)
	}
	if status.FileSize != fileSize {
		t.Fatalf("expected fileSize %d, got %d", fileSize, status.FileSize)
	}
}

func TestQueueStatusCompleted_SelectsEpisodeMatch(t *testing.T) {
	service, nzbSystem, metadataSvc := setupPlaybackService(t)
	importerSvc := nzbSystem.ImporterService()

	releaseDir := "/virtual/Another.Show.Release"
	metadataSvc.files[releaseDir] = []string{
		"Another.Show.S01E01.mkv",
		"Another.Show.S01E02.mkv",
	}

	fileSize := int64(4096)
	item := &database.ImportQueueItem{
		NzbPath:    "Another.Show.S01E02.nzb",
		Priority:   database.QueuePriorityNormal,
		Status:     database.QueueStatusPending,
		MaxRetries: 3,
		FileSize:   &fileSize,
	}

	if err := importerSvc.Database().Repository.AddToQueue(item); err != nil {
		t.Fatalf("add to queue: %v", err)
	}

	meta := `{"sourceNzbPath":"Another.Show.S01E02.2160p.WEB-DL.mkv","preflightHealth":"healthy"}`
	if err := importerSvc.Database().Repository.UpdateMetadata(item.ID, &meta); err != nil {
		t.Fatalf("set metadata: %v", err)
	}

	if err := importerSvc.Database().Repository.AddStoragePath(item.ID, releaseDir); err != nil {
		t.Fatalf("add storage path: %v", err)
	}

	if err := importerSvc.Database().Repository.UpdateQueueItemStatus(item.ID, database.QueueStatusCompleted, nil); err != nil {
		t.Fatalf("mark completed: %v", err)
	}

	status, err := service.QueueStatus(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("QueueStatus returned error: %v", err)
	}

	if !strings.HasSuffix(status.WebDAVPath, "/Another.Show.S01E02.mkv") {
		t.Fatalf("expected webdav path to end with S01E02 file, got %q", status.WebDAVPath)
	}
}

func TestQueueStatusCompleted_PrefersTitleSimilarity(t *testing.T) {
	service, nzbSystem, metadataSvc := setupPlaybackService(t)
	importerSvc := nzbSystem.ImporterService()

	releaseDir := "/virtual/Movie.Collection"
	metadataSvc.files[releaseDir] = []string{
		"Bonus.Featurette.mkv",
		"Movie.Title.2023.2160p.BluRay.x265.mkv",
	}

	fileSize := int64(8192)
	item := &database.ImportQueueItem{
		NzbPath:    "Movie.Title.2023.2160p.BluRay.x265-GROUP.nzb",
		Priority:   database.QueuePriorityNormal,
		Status:     database.QueueStatusPending,
		MaxRetries: 3,
		FileSize:   &fileSize,
	}

	if err := importerSvc.Database().Repository.AddToQueue(item); err != nil {
		t.Fatalf("add to queue: %v", err)
	}

	meta := `{"sourceNzbPath":"Movie.Title.2023.2160p.BluRay.x265-GROUP.nzb","preflightHealth":"healthy"}`
	if err := importerSvc.Database().Repository.UpdateMetadata(item.ID, &meta); err != nil {
		t.Fatalf("set metadata: %v", err)
	}

	if err := importerSvc.Database().Repository.AddStoragePath(item.ID, releaseDir); err != nil {
		t.Fatalf("add storage path: %v", err)
	}

	if err := importerSvc.Database().Repository.UpdateQueueItemStatus(item.ID, database.QueueStatusCompleted, nil); err != nil {
		t.Fatalf("mark completed: %v", err)
	}

	status, err := service.QueueStatus(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("QueueStatus returned error: %v", err)
	}

	if !strings.HasSuffix(status.WebDAVPath, "/Movie.Title.2023.2160p.BluRay.x265.mkv") {
		t.Fatalf("expected movie file to be selected, got %q", status.WebDAVPath)
	}
}
