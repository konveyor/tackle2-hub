package api

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
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
	if ctx.Request.Header.Get(Accept) == TarGzMimetype {
		h.getDirArchive(ctx, owner)
	} else {
		h.content(ctx, owner)
	}
}

func (h *BucketHandler) serveBucketUpload(ctx *gin.Context, owner *model.BucketOwner) {
	if ctx.Request.Header.Get(ContentType) == TarGzMimetype {
		h.uploadDirArchive(ctx, owner)
	} else {
		h.upload(ctx, owner)
	}
}

func (h *BucketHandler) uploadDirArchive(ctx *gin.Context, owner *model.BucketOwner) {
	bucketPath := owner.Bucket

	// Prepare to uncompress the uploaded data
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

	// Clean the destination bucket directory (or do an atomic way - extract to some temp directory and move when extreact finished successfully)
	bucketContent, err := ioutil.ReadDir(bucketPath)
	if err != nil {
		log.Error(err, "read bucket dir")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	for _, bucketEntry := range bucketContent {
		err = os.RemoveAll(bucketPath + "/" + bucketEntry.Name())
		if err != nil {
			log.Info("Cleaning bucket dir, cannot delete")
		}
	}

	// Extract the tar archive
	untarReader := tar.NewReader(ungzReader)
	for {
		hdr, err := untarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Error(err, "read tar")
				ctx.Status(http.StatusInternalServerError)
				return
			}
		}
		fmt.Printf("HDR: %v", &hdr)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(bucketPath+"/"+hdr.Name, hdr.FileInfo().Mode().Perm()); err != nil {
				log.Error(err, "create dir")
				ctx.Status(http.StatusInternalServerError)
				return
			}
		case tar.TypeReg: // Regular file
			var file *os.File
			if file, err = os.Create(bucketPath + "/" + hdr.Name); err != nil {
				log.Error(err, "create file")
				ctx.Status(http.StatusInternalServerError)
				return
			}
			//os.Chmod / chown?
			if _, err = io.Copy(file, untarReader); err != nil {
				log.Error(err, "copy to tar")
				ctx.Status(http.StatusInternalServerError)
				return
			}
			file.Close()
		default:
			// types that are not files or dirs are skipped / add warning?
		}
	}

	ctx.Status(http.StatusAccepted)
}

func (h *BucketHandler) getDirArchive(ctx *gin.Context, owner *model.BucketOwner) {
	dir := owner.Bucket + ctx.Param(Wildcard)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Bucket (sub)directory doesn't exist.",
		})
		return
	}

	var tarOutput bytes.Buffer
	tarWriter := tar.NewWriter(&tarOutput)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		hdr, err := tar.FileInfoHeader(info, path)
		if err != nil {
			panic(err)
		}

		// Scope path in archive to the given dir
		hdr.Name = strings.Replace(path, dir, "", 1)
		if hdr.Name == "" {
			return nil
		}

		switch hdr.Typeflag {
		case tar.TypeDir, tar.TypeSymlink:
			// Add file/directory header to the archive
			if err := tarWriter.WriteHeader(hdr); err != nil {
				panic(err)
			}
		case tar.TypeReg:
			// Add file&data to the archive
			if err := tarWriter.WriteHeader(hdr); err != nil {
				panic(err)
			}
			file, _ := os.Open(path)

			if _, err = io.Copy(tarWriter, file); err != nil {
				panic(err)
			}
		default:
			// Other file types like block/character device are skipped.
			// Complete list of types: https://pkg.go.dev/archive/tar#pkg-constants
		}

		return nil
	})

	if err := tarWriter.Close(); err != nil {
		fmt.Println(err)
	}

	if err != nil {
		fmt.Println(err)
	}

	fromTar := bufio.NewReader(&tarOutput)

	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(dir)+".tar.gz"))

	gzWriter := gzip.NewWriter(ctx.Writer)
	defer gzWriter.Close()

	gzWriter.Name = path.Base(dir) + ".tar.gz"
	gzWriter.Comment = "Tackle 2 bucket data archive"
	if _, err := io.Copy(gzWriter, fromTar); err != nil {
		fmt.Println(err)
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
