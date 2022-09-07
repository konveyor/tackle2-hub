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
	pathlib "path"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
)

//
// Routes
const (
	BucketsRoot = "/buckets"
	BucketRoot  = BucketsRoot + "/:" + ID
)

//
// BucketHandler provides bucket management.
type BucketHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h BucketHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	// TODO: add auth middleware when buckets scope added to the token
	// routeGroup.Use(auth.AuthorizationRequired(h.AuthProvider, "buckets"))
	routeGroup.GET(BucketRoot, h.Get)
	routeGroup.PUT(BucketRoot+"/:"+ID, h.Put)
}

// Get godoc
// @summary Upload bucket content archive by ID.
// @description Upload a bucket content .tar.gz archive by ID as PUT (replace all previous bucket content).
// @tags put
// @produce octet-stream
// @success 200 {binary}
// @router /buckets/{id} [put]
// @param id path string true "Bucket ID"
func (h BucketHandler) Put(ctx *gin.Context) {
	invalidSymbols := regexp.MustCompile("[^a-zA-Z0-9-_]")
	bucketID := invalidSymbols.ReplaceAllString(ctx.Param("ID"), "")
	bucketPath := "/tmp/bucket" + bucketID

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

// Get godoc
// @summary Get a bucket content archive by ID.
// @description Get a bucket content .tar.gz archive by ID.
// @tags get
// @produce octet-stream
// @success 200 {binary}
// @router /buckets/{id} [get]
// @param id path string true "Bucket ID"
func (h BucketHandler) Get(ctx *gin.Context) {

	invalidSymbols := regexp.MustCompile("[^a-zA-Z0-9-_]")
	bucketID := invalidSymbols.ReplaceAllString(ctx.Param("ID"), "")

	if _, err := os.Stat("/tmp/bucket/" + bucketID); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, "Bucket doesn't exist.")
		return
	}

	var tarOutput bytes.Buffer
	tarWriter := tar.NewWriter(&tarOutput)

	err := filepath.Walk("/tmp/bucket/"+bucketID, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)

		hdr, err := tar.FileInfoHeader(info, path)
		if err != nil {
			panic(err)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := tarWriter.WriteHeader(hdr); err != nil {
				panic(err)
			}
		case tar.TypeReg:
			if err := tarWriter.WriteHeader(hdr); err != nil {
				panic(err)
			}
			file, _ := os.Open(path)

			if _, err = io.Copy(tarWriter, file); err != nil {
				panic(err)
			}
		default:
			// No symlinks, devices etc are added to the archive / add warning?
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

	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", bucketID+".tar.gz"))

	gzWriter := gzip.NewWriter(ctx.Writer)
	defer gzWriter.Close()

	gzWriter.Name = bucketID + ".tar.gz"
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
