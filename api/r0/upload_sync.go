package r0

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/h2non/filetype"
	"github.com/sirupsen/logrus"
	"github.com/t2bot/matrix-media-repo/api/_apimeta"
	"github.com/t2bot/matrix-media-repo/api/_responses"
	"github.com/t2bot/matrix-media-repo/common"
	"github.com/t2bot/matrix-media-repo/common/rcontext"
	"github.com/t2bot/matrix-media-repo/datastores"
	"github.com/t2bot/matrix-media-repo/pipelines/pipeline_upload"
	"github.com/t2bot/matrix-media-repo/util"
)

type MediaUploadedResponse struct {
	ContentUri string `json:"content_uri,omitempty"`
}

func UploadMediaSync(r *http.Request, rctx rcontext.RequestContext, user _apimeta.UserInfo) interface{} {
	filename := filepath.Base(r.URL.Query().Get("filename"))
	rctx = rctx.LogWithFields(logrus.Fields{
		"filename": filename,
	})
	// GK CUSTOMIZATION: Sanitize the filename
	if len(filename) > rctx.Config.Uploads.MaxFilenameLength {
		rctx.Log.Info("Filename too long")
		return &_responses.ErrorResponse{
			Code:         common.ErrCodeBadRequest,
			Message:      "Filename too long.",
			InternalCode: common.ErrCodeBadRequest,
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream" // binary
	} else {
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
		if !util.IsSupportedFileType(kind.Extension, rctx.Config.Uploads.SupportedFileTypes) {
			rctx.Log.Info("Unsupported file type: ", kind.Extension)
			return &_responses.ErrorResponse{
				Code:         common.ErrCodeBadRequest,
				Message:      "Unsupported file type.",
				InternalCode: common.ErrCodeBadRequest,
			}
		}
		//
	}

	// Early sizing constraints (reject requests which claim to be too large/small)
	if sizeRes := uploadRequestSizeCheck(rctx, r); sizeRes != nil {
		return sizeRes
	}

	// Actually upload
	media, err := pipeline_upload.Execute(rctx, r.Host, "", r.Body, contentType, filename, user.UserId, datastores.LocalMediaKind)
	if err != nil {
		if errors.Is(err, common.ErrQuotaExceeded) {
			return _responses.QuotaExceeded()
		}
		rctx.Log.Error("Unexpected error uploading media: ", err)
		sentry.CaptureException(err)
		return _responses.InternalServerError("Unexpected Error")
	}

	return &MediaUploadedResponse{
		ContentUri: util.MxcUri(media.Origin, media.MediaId),
	}
}

func uploadRequestSizeCheck(rctx rcontext.RequestContext, r *http.Request) *_responses.ErrorResponse {
	maxSize := rctx.Config.Uploads.MaxSizeBytes
	minSize := rctx.Config.Uploads.MinSizeBytes
	if maxSize > 0 || minSize > 0 {
		if r.ContentLength > 0 {
			if maxSize > 0 && maxSize < r.ContentLength {
				return _responses.RequestTooLarge()
			}
			if minSize > 0 && minSize > r.ContentLength {
				return _responses.RequestTooSmall()
			}
		} else {
			header := r.Header.Get("Content-Length")
			if header != "" {
				parsed, _ := strconv.ParseInt(header, 10, 64)
				if maxSize > 0 && maxSize < parsed {
					return _responses.RequestTooLarge()
				}
				if minSize > 0 && minSize > parsed {
					return _responses.RequestTooSmall()
				}
			}
		}
	}
	return nil
}
