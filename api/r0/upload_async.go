package r0

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/getsentry/sentry-go"
	"github.com/h2non/filetype"
	"github.com/sirupsen/logrus"
	"github.com/t2bot/matrix-media-repo/api/_apimeta"
	"github.com/t2bot/matrix-media-repo/api/_responses"
	"github.com/t2bot/matrix-media-repo/api/_routers"
	"github.com/t2bot/matrix-media-repo/common"
	"github.com/t2bot/matrix-media-repo/common/rcontext"
	"github.com/t2bot/matrix-media-repo/pipelines/pipeline_upload"
	"github.com/t2bot/matrix-media-repo/util"
)

func UploadMediaAsync(r *http.Request, rctx rcontext.RequestContext, user _apimeta.UserInfo) interface{} {
	server := _routers.GetParam("server", r)
	mediaId := _routers.GetParam("mediaId", r)
	filename := filepath.Base(r.URL.Query().Get("filename"))
	// GK CUSTOMIZATION: Sanitize the filename
	if len(filename) > 24 {
		return &_responses.ErrorResponse{
			Code:         common.ErrCodeBadRequest,
			Message:      "Filename too long.",
			InternalCode: common.ErrCodeBadRequest,
		}
	}

	// GK CUSTOMIZATION: Check if the file type is supported
	buf, err := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(buf))
	if err != nil {
		return &_responses.ErrorResponse{
			Code:         common.ErrCodeBadRequest,
			Message:      "Error reading file.",
			InternalCode: common.ErrCodeBadRequest,
		}
	}
	kind, err := filetype.Match(buf)
	if err != nil {
		return &_responses.ErrorResponse{
			Code:         common.ErrCodeBadRequest,
			Message:      "Error matching file type.",
			InternalCode: common.ErrCodeBadRequest,
		}
	}
	if !util.IsSupportedFileType(kind.Extension) {
		return &_responses.ErrorResponse{
			Code:         common.ErrCodeBadRequest,
			Message:      "Unsupported file type.",
			InternalCode: common.ErrCodeBadRequest,
		}
	}

	rctx = rctx.LogWithFields(logrus.Fields{
		"mediaId":  mediaId,
		"server":   server,
		"filename": filename,
	})

	if r.Host != server {
		return &_responses.ErrorResponse{
			Code:         common.ErrCodeNotFound,
			Message:      "Upload request is for another domain.",
			InternalCode: common.ErrCodeForbidden,
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream" // binary
	}

	// Early sizing constraints (reject requests which claim to be too large/small)
	if sizeRes := uploadRequestSizeCheck(rctx, r); sizeRes != nil {
		return sizeRes
	}

	// Actually upload
	_, err = pipeline_upload.ExecutePut(rctx, server, mediaId, r.Body, contentType, filename, user.UserId)
	if err != nil {
		if errors.Is(err, common.ErrQuotaExceeded) {
			return _responses.QuotaExceeded()
		} else if errors.Is(err, common.ErrAlreadyUploaded) {
			return &_responses.ErrorResponse{
				Code:         common.ErrCodeCannotOverwrite,
				Message:      "This media has already been uploaded.",
				InternalCode: common.ErrCodeCannotOverwrite,
			}
		} else if errors.Is(err, common.ErrWrongUser) {
			return &_responses.ErrorResponse{
				Code:         common.ErrCodeForbidden,
				Message:      "You do not have permission to upload this media.",
				InternalCode: common.ErrCodeForbidden,
			}
		} else if errors.Is(err, common.ErrExpired) {
			return &_responses.ErrorResponse{
				Code:         common.ErrCodeNotFound,
				Message:      "Media expired or not found.",
				InternalCode: common.ErrCodeNotFound,
			}
		}
		rctx.Log.Error("Unexpected error uploading media: ", err)
		sentry.CaptureException(err)
		return _responses.InternalServerError("Unexpected Error")
	}

	return &MediaUploadedResponse{
		//ContentUri: util.MxcUri(media.Origin, media.MediaId), // This endpoint doesn't return a URI
	}
}
