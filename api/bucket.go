package api

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	pathlib "path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
)

//
// BucketHandler provides bucket management.
type BucketHandler struct {
}

func (h *BucketHandler) serveBucketGet(ctx *gin.Context, owner *model.BucketOwner) {
	if ctx.Request.Header.Get(Directory) == DirectoryArchive {
		h.getDirArchive(ctx, path.Join(owner.Bucket, ctx.Param(Wildcard)))
	} else {
		h.content(ctx, owner)
	}
}

func (h *BucketHandler) serveBucketUpload(ctx *gin.Context, owner *model.BucketOwner) {
	if ctx.Request.Header.Get(Directory) == DirectoryExpand {
		h.uploadDirArchive(ctx, path.Join(owner.Bucket, ctx.Param(Wildcard)))
	} else {
		h.upload(ctx, owner)
	}
}

func (h *BucketHandler) uploadDirArchive(ctx *gin.Context, dir string) {
	// Prepare to uncompress the uploaded data, report 4xx errors
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	fileReader, err := file.Open()
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	defer fileReader.Close()

	ungzReader, err := gzip.NewReader(fileReader)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	defer ungzReader.Close()

	// Report 5xx errors for extraction process
	defer func() {
		if err != nil {
			log.Error(err, "bucket archive expand action failed")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
	}()

	// Clean and prepare destination directory
	err = os.RemoveAll(dir)
	if err != nil {
		return
	}
	if err = os.MkdirAll(dir, 0777); err != nil {
		return
	}

	// Extract the tar archive
	untarReader := tar.NewReader(ungzReader)
	for {
		hdr, err := untarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return
			}
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path.Join(dir, hdr.Name), 0777); err != nil {
				return
			}
		case tar.TypeReg:
			var file *os.File
			if file, err = os.Create(path.Join(dir, hdr.Name)); err != nil {
				return
			}
			if _, err = io.Copy(file, untarReader); err != nil {
				return
			}
			file.Close()
		default:
			// types that are not files or dirs are skipped
		}
	}

	ctx.Status(http.StatusAccepted)
}

func (h *BucketHandler) getDirArchive(ctx *gin.Context, dir string) {
	dirInfo, err := os.Stat(dir)
	if os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Provided directory path doesn't exist.",
		})
		return
	}
	if !dirInfo.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Provided path is not a directory.",
		})
		return
	}

	var tarOutput bytes.Buffer
	entriesCount := 0
	tarWriter := tar.NewWriter(&tarOutput)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		// Scope path in archive to the given dir
		hdr.Name = strings.Replace(path, dir, "", 1)
		if hdr.Name == "" {
			return nil
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			// Add directory header to the archive (no content)
			if err := tarWriter.WriteHeader(hdr); err != nil {
				return err
			}
			entriesCount += 1
		case tar.TypeReg:
			// Add file with its content to the archive
			if err := tarWriter.WriteHeader(hdr); err != nil {
				return err
			}
			file, _ := os.Open(path)

			if _, err = io.Copy(tarWriter, file); err != nil {
				return err
			}
			entriesCount += 1
		default:
			// Other file types like block/character device or TypeSymlink are skipped.
			// Complete list of types: https://pkg.go.dev/archive/tar#pkg-constants
		}

		return nil
	})

	// Report 5xx errors in archive creation steps
	defer func() {
		if err != nil {
			log.Error(err, "bucket archive get action failed")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
	}()

	if err != nil {
		return
	}

	if err := tarWriter.Close(); err != nil {
		return
	}

	// Return empty response instead of empty tar.gz archive
	if entriesCount < 1 {
		ctx.Status(http.StatusNoContent)
		return
	}

	fromTar := bufio.NewReader(&tarOutput)

	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(dir)+".tar.gz"))

	gzWriter := gzip.NewWriter(ctx.Writer)
	defer gzWriter.Close()

	gzWriter.Name = path.Base(dir) + ".tar.gz"
	gzWriter.Comment = "Tackle 2 bucket data archive"
	if _, err = io.Copy(gzWriter, fromTar); err != nil {
		return
	}
}

//
// content at path.
func (h *BucketHandler) content(ctx *gin.Context, owner *model.BucketOwner) {
	if owner.Bucket == "" {
		ctx.Status(http.StatusNotFound)
		return
	}
	rPath := ctx.Param(Wildcard)
	ctx.File(pathlib.Join(
		owner.Bucket,
		rPath))
}

//
// upload file at path.
func (h *BucketHandler) upload(ctx *gin.Context, owner *model.BucketOwner) {
	if owner.Bucket == "" {
		ctx.Status(http.StatusNotFound)
		return
	}
	rPath := ctx.Param(Wildcard)
	path := pathlib.Join(
		owner.Bucket,
		rPath)
	input, err := ctx.FormFile("file")
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	defer func() {
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}()
	reader, err := input.Open()
	if err != nil {
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	err = os.MkdirAll(pathlib.Dir(path), 0777)
	if err != nil {
		return
	}
	writer, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		_ = writer.Close()
	}()
	_, err = io.Copy(writer, reader)
	if err != nil {
		return
	}
	err = os.Chmod(path, 0666)
	if err != nil {
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Delete from the bucket at path.
func (h *BucketHandler) delete(ctx *gin.Context, owner *model.BucketOwner) {
	if owner.Bucket == "" {
		ctx.Status(http.StatusNotFound)
		return
	}
	rPath := ctx.Param(Wildcard)
	path := pathlib.Join(
		owner.Bucket,
		rPath)
	err := os.RemoveAll(path)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusNoContent)
}
