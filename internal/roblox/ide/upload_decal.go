package ide	

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/roblox"
)

var UploadDecalErrors = struct {
	ErrNotLoggedIn       error
	ErrTokenInvalid      error
	ErrInappropriateName error
}{
	ErrNotLoggedIn:       errors.New("not logged in"),
	ErrTokenInvalid:      errors.New("XSRF token validation failed"),
	ErrInappropriateName: errors.New("inappropriate name or description"),
}

func newUploadDecalURL(groupID int64, name, description string) string {
	endpoint := fmt.Sprintf(
		"https://data.roblox.com/ide/publish/UploadNewAsset?assetTypeName=Decal&name=%s&description=%s",
		url.QueryEscape(name),
		url.QueryEscape(description),
	)
	if groupID > 0 {
		endpoint += fmt.Sprintf("&groupId=%d", groupID)
	}
	return endpoint
}

func newUploadDecalRequest(groupID int64, name, description string, data *bytes.Buffer) (*http.Request, error) {
	endpoint := newUploadDecalURL(groupID, name, description)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(data.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RobloxStudio/WinInet")
	req.Header.Set("Content-Type", "image/png")
	return req, nil
}

func NewUploadDecalHandler(
	c *roblox.Client,
	name, description string,
	data *bytes.Buffer,
	groupID int64,
) (func() (int64, error), error) {
	req, err := newUploadDecalRequest(groupID, name, description, data)
	if err != nil {
		return func() (int64, error) { return 0, nil }, err
	}

	return func() (int64, error) {
		req.AddCookie(&http.Cookie{
			Name:  ".ROBLOSECURITY",
			Value: c.Cookie,
		})
		req.Header.Set("x-csrf-token", c.GetToken())

		resp, err := c.DoRequest(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			id, err := strconv.ParseInt(string(body), 10, 64)
			if err != nil {
				return 0, err
			}
			return id, nil
		case http.StatusForbidden:
			strBody := string(body)
			if strBody == "NotLoggedIn" {
				return 0, UploadDecalErrors.ErrNotLoggedIn
			} else if strBody == "XSRF Token Validation Failed" {
				c.SetToken(resp.Header.Get("x-csrf-token"))
				return 0, UploadDecalErrors.ErrTokenInvalid
			}
			return 0, errors.New(resp.Status)
		case http.StatusUnprocessableEntity:
			if string(body) == "Inappropriate name or description." {
				req, _ = newUploadDecalRequest(groupID, "[Censored]", description, data)
				return 0, UploadDecalErrors.ErrInappropriateName
			}
			return 0, errors.New(resp.Status)
		default:
			return 0, errors.New(resp.Status)
		}
	}, nil
}
