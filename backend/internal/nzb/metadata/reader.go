package metadata

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	metapb "novastream/internal/nzb/metadata/proto"
)

// MetadataReader provides read operations for the virtual filesystem.
type MetadataReader struct {
	service *MetadataService
}

// NewMetadataReader creates a new metadata reader bound to a metadata service.
func NewMetadataReader(service *MetadataService) *MetadataReader {
	return &MetadataReader{service: service}
}

// ListDirectoryContents lists directory entries beneath a virtual path.
// Directories are returned as fs.FileInfo; files are returned as FileMetadata.
func (mr *MetadataReader) ListDirectoryContents(virtualPath string) ([]fs.FileInfo, []*metapb.FileMetadata, error) {
	virtualPath = filepath.Clean(virtualPath)
	if virtualPath == "." {
		virtualPath = "/"
	}

	metadataDir := mr.service.GetMetadataDirectoryPath(virtualPath)

	entries, err := os.ReadDir(metadataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []fs.FileInfo{}, []*metapb.FileMetadata{}, nil
		}
		return nil, nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var dirs []fs.FileInfo
	var files []*metapb.FileMetadata

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				dirs = append(dirs, info)
			}
			continue
		}

		if filepath.Ext(entry.Name()) != ".meta" {
			continue
		}

		virtualName := entry.Name()[:len(entry.Name())-5]
		virtualFilePath := filepath.Join(virtualPath, virtualName)

		fileMeta, err := mr.service.ReadFileMetadata(virtualFilePath)
		if err != nil || fileMeta == nil {
			continue
		}
		files = append(files, fileMeta)
	}

	return dirs, files, nil
}

// GetDirectoryInfo returns filesystem info about a virtual directory.
func (mr *MetadataReader) GetDirectoryInfo(virtualPath string) (fs.FileInfo, error) {
	metadataPath := mr.service.GetMetadataDirectoryPath(virtualPath)
	info, err := os.Stat(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("directory not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", virtualPath)
	}
	return info, nil
}

// GetFileMetadata returns the metadata stored for a virtual file.
func (mr *MetadataReader) GetFileMetadata(virtualPath string) (*metapb.FileMetadata, error) {
	return mr.service.ReadFileMetadata(virtualPath)
}

// PathExists reports whether the virtual path resolves to either a directory or file.
func (mr *MetadataReader) PathExists(virtualPath string) (bool, error) {
	if mr.service.DirectoryExists(virtualPath) {
		return true, nil
	}
	if mr.service.FileExists(virtualPath) {
		return true, nil
	}
	return false, nil
}

// IsDirectory reports whether the virtual path refers to a directory.
func (mr *MetadataReader) IsDirectory(virtualPath string) (bool, error) {
	if mr.service.DirectoryExists(virtualPath) {
		return true, nil
	}
	if mr.service.FileExists(virtualPath) {
		return false, nil
	}
	return false, fmt.Errorf("path does not exist: %s", virtualPath)
}

// GetFileSegments returns Usenet segments backing the virtual file.
func (mr *MetadataReader) GetFileSegments(virtualPath string) ([]*metapb.SegmentData, error) {
	fileMeta, err := mr.service.ReadFileMetadata(virtualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file metadata: %w", err)
	}
	if fileMeta == nil {
		return nil, fmt.Errorf("file not found: %s", virtualPath)
	}
	return fileMeta.SegmentData, nil
}

// ListDirectory lists all files in a virtual directory (returns filenames without .meta extension)
func (mr *MetadataReader) ListDirectory(virtualPath string) ([]string, error) {
	return mr.service.ListDirectory(virtualPath)
}

// ListSubdirectories lists all subdirectories in a virtual directory
func (mr *MetadataReader) ListSubdirectories(virtualPath string) ([]string, error) {
	return mr.service.ListSubdirectories(virtualPath)
}
