// Package ocibundle provides OCI bundle distribution functionality.
// This file contains the tar.gz implementation of the Archiver interface.
package ocibundle

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TarGzArchiver implements the Archiver interface using tar.gz format.
// It provides secure, streaming archive and extraction operations with
// comprehensive validation and progress reporting capabilities.
type TarGzArchiver struct{}

// NewTarGzArchiver creates a new TarGzArchiver instance.
// The archiver uses standard tar.gz format compatible with common tools
// and implements security validation during extraction.
func NewTarGzArchiver() *TarGzArchiver {
	return &TarGzArchiver{}
}

// Archive creates a tar.gz archive from the specified source directory.
// The archive is written to the provided output writer in a streaming fashion
// to minimize memory usage even with large directories.
//
// Parameters:
//   - ctx: Context for cancellation
//   - sourceDir: Directory to archive (must exist and be readable)
//   - output: Writer to receive the compressed archive data
//
// Returns an error if the source directory doesn't exist, is not readable,
// or if writing to the output fails.
func (a *TarGzArchiver) Archive(ctx context.Context, sourceDir string, output io.Writer) error {
	return a.ArchiveWithProgress(ctx, sourceDir, output, nil)
}

// ArchiveWithProgress creates a tar.gz archive from the specified source directory with progress reporting.
// The archive is written to the provided output writer in a streaming fashion
// to minimize memory usage even with large directories.
//
// Parameters:
//   - ctx: Context for cancellation
//   - sourceDir: Directory to archive (must exist and be readable)
//   - output: Writer to receive the compressed archive data
//   - progress: Optional callback for progress reporting (current, total bytes)
//
// Returns an error if the source directory doesn't exist, is not readable,
// or if writing to the output fails.
func (a *TarGzArchiver) ArchiveWithProgress(ctx context.Context, sourceDir string, output io.Writer, progress func(current, total int64)) error {
	if sourceDir == "" {
		return fmt.Errorf("source directory cannot be empty")
	}

	if output == nil {
		return fmt.Errorf("output writer cannot be nil")
	}

	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// If progress callback is provided, calculate total size first
	var totalSize int64
	if progress != nil {
		err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode().IsRegular() {
				totalSize += info.Size()
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to calculate total size: %w", err)
		}
	}

	gzipWriter := gzip.NewWriter(output)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	var currentSize int64
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", path, err)
		}

		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", path, err)
		}

		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()

			// Copy with progress reporting if callback provided
			if progress != nil {
				written, err := a.copyWithProgress(tarWriter, file, func(written int64) {
					currentSize += written
					progress(currentSize, totalSize)
				})
				if err != nil {
					return fmt.Errorf("failed to write file content for %s: %w", path, err)
				}
				_ = written // written is used for progress tracking
			} else {
				if _, err := io.Copy(tarWriter, file); err != nil {
					return fmt.Errorf("failed to write file content for %s: %w", path, err)
				}
			}
		}

		return nil
	})
}

// copyWithProgress copies data from src to dst while reporting progress
func (a *TarGzArchiver) copyWithProgress(dst io.Writer, src io.Reader, progress func(int64)) (int64, error) {
	buf := make([]byte, 32*1024) // 32KB buffer
	var total int64

	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return total, writeErr
			}
			total += int64(n)
			if progress != nil {
				progress(int64(n))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// Extract expands a tar.gz archive to the specified target directory.
// The extraction process includes security validation to prevent common
// archive-based attacks such as path traversal and resource exhaustion.
//
// Parameters:
//   - ctx: Context for cancellation
//   - input: Reader providing the compressed archive data
//   - targetDir: Directory to extract files to (created if it doesn't exist)
//   - opts: Extraction options controlling security limits and behavior
//
// Returns an error if the archive is corrupted, contains security violations,
// exceeds configured limits, or if file operations fail.
func (a *TarGzArchiver) Extract(ctx context.Context, input io.Reader, targetDir string, opts ExtractOptions) error {
	if input == nil {
		return fmt.Errorf("input reader cannot be nil")
	}

	if targetDir == "" {
		return fmt.Errorf("target directory cannot be empty")
	}

	gzipReader, err := gzip.NewReader(input)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	validators := NewValidatorChain(
		NewSizeValidator(opts.MaxFileSize, opts.MaxSize),
		NewFileCountValidator(opts.MaxFiles),
		NewPermissionSanitizer(),
	)

	totalSize := int64(0)
	fileCount := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		fileCount++

		if err := validators.ValidatePath(header.Name); err != nil {
			return NewBundleError("extract", header.Name, err)
		}

		filePath := header.Name
		if opts.StripPrefix != "" && strings.HasPrefix(filePath, opts.StripPrefix) {
			filePath = strings.TrimPrefix(filePath, opts.StripPrefix)
			filePath = strings.TrimPrefix(filePath, "/")
		}

		fullPath := filepath.Join(targetDir, filePath)

		if !strings.HasPrefix(fullPath, targetDir) {
			return NewBundleError("extract", header.Name, ErrSecurityViolation)
		}

		fileInfo := FileInfo{
			Name: header.Name,
			Size: header.Size,
			Mode: uint32(header.Mode),
		}

		if err := validators.ValidateFile(fileInfo); err != nil {
			return NewBundleError("extract", header.Name, err)
		}

		if opts.MaxFiles > 0 && fileCount > opts.MaxFiles {
			return NewBundleError("extract", header.Name, ErrSecurityViolation)
		}

		totalSize += header.Size

		archiveStats := ArchiveStats{
			TotalFiles: fileCount,
			TotalSize:  totalSize,
		}
		if err := validators.ValidateArchive(archiveStats); err != nil {
			return NewBundleError("extract", header.Name, err)
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", fullPath, err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
			}

		case tar.TypeReg:
			file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", fullPath, err)
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return fmt.Errorf("failed to write file content for %s: %w", fullPath, err)
			}

		case tar.TypeSymlink:
			linkTarget := header.Linkname

			if err := validators.ValidatePath(linkTarget); err != nil {
				return NewBundleError("extract", header.Name, err)
			}

			if err := os.Symlink(linkTarget, fullPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", fullPath, linkTarget, err)
			}

		default:
			continue
		}

		if !opts.PreservePerms && header.Typeflag == tar.TypeReg {
			if err := os.Chmod(fullPath, 0644); err != nil {
				return fmt.Errorf("failed to set permissions for %s: %w", fullPath, err)
			}
		}
	}

	return nil
}

// MediaType returns the OCI media type for tar.gz archives.
// This is used when pushing bundles to OCI registries to identify
// the archive format to registry clients.
func (a *TarGzArchiver) MediaType() string {
	return "application/vnd.oci.image.layer.v1.tar+gzip"
}
