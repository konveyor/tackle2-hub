package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"io"
	"net/http"
	"os"
	pathlib "path"
)

//
// BucketHandler provides bucket management.
type BucketHandler struct {
}

//
// content at path.
func (h *BucketHandler) content(ctx *gin.Context, owner *model.BucketOwner) {
	rPath := ctx.Param(Wildcard)
	ctx.File(pathlib.Join(
		owner.Bucket,
		rPath))
}

//
// upload file at path.
func (h *BucketHandler) upload(ctx *gin.Context, owner *model.BucketOwner) {
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
