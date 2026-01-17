package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, _ := r.FormFile("thumbnail")
	mediaType, _, _ := mime.ParseMediaType(header.Header.Get("Content-Type"))

	if mediaType != "image/jpeg" || mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", errors.New("Invalid file type"))
		return
	}

	imgData, _ := io.ReadAll(file)

	video, _ := cfg.db.GetVideo(videoID)

	upload := thumbnail{
		data:      imgData,
		mediaType: mediaType,
	}

	parts := strings.Split(mediaType, "/")
	fp := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s.%s", videoID, parts[1]))
	f, _ := os.Create(fp)
	io.Copy(f, file)

	thumbnailUrl := fmt.Sprintf("http://localhost:%s/%s", cfg.port, fp)
	video.ThumbnailURL = &thumbnailUrl

	videoThumbnails[videoID] = upload
	cfg.db.UpdateVideo(video)

	respondWithJSON(w, http.StatusOK, video)
}
