package auth

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)

// DagHandler handles device authorization grant HTTP requests.
type DagHandler struct {
	storage *Storage
}

// Verify godoc
// @summary Device authorization verification page.
// @description Display page for user to enter device code.
// @tags auth
// @produce html
// @success 200
// @router /auth/device [get]
//
// Verify displays device authorization verification page.
func (h *DagHandler) Verify(ctx *gin.Context) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Device Authorization</title>
</head>
<body>
    <h1>Device Authorization</h1>
    <form method="POST" action="/auth/device">
        <label for="userCode">User Code:</label>
        <input type="text" id="userCode" name="userCode" required>
        <button type="submit">Authorize</button>
    </form>
</body>
</html>
`
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// VerifySubmit godoc
// @summary Submit device authorization.
// @description Process user authorization for device flow.
// @tags auth
// @accept application/x-www-form-urlencoded
// @produce html
// @success 200
// @router /auth/device [post]
// @param userCode formData string true "User code from device"
//
// VerifySubmit processes device authorization submission.
func (h *DagHandler) VerifySubmit(ctx *gin.Context) {
	userCode := ctx.PostForm("userCode")
	if userCode == "" {
		_ = ctx.Error(&BadRequestError{
			Reason: "userCode not provided.",
		})
		return
	}

	devAuth, found := h.storage.GetDevAuthByUserCode(userCode)
	if !found {
		_ = ctx.Error(&NotFound{
			Resource: "device authorization",
			Id:       userCode,
		})
		return
	}
	if devAuth.Done() || devAuth.Denied() {
		_ = ctx.Error(&BadRequestError{
			Reason: "userCode already used.",
		})
		return
	}
	if time.Now().After(devAuth.Expiration()) {
		_ = ctx.Error(&BadRequestError{
			Reason: "userCode expired.",
		})
		return
	}

	subject := h.currentUser(ctx)
	err := h.storage.UpdateDevAuth(userCode, subject, true, false, time.Now())
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Authorization Complete</title>
</head>
<body>
    <h1>Authorization Complete</h1>
    <p>You have successfully authorized the device. You may close this window.</p>
</body>
</html>
`
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// currentUser returns the authenticated user from the gin context.
func (h *DagHandler) currentUser(ctx *gin.Context) (user string) {
	// RichContext from api package stores User field
	// We can't import api package, so access via reflection
	key := "RichContext"
	object, found := ctx.Get(key)
	if found {
		// Use reflect to get User field without importing api package
		v := reflect.ValueOf(object)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			userField := v.FieldByName("User")
			if userField.IsValid() && userField.Kind() == reflect.String {
				user = userField.String()
			}
		}
	}
	return
}
