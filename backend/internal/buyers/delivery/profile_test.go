package delivery

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"
)

type avatarRepoStub struct {
	key, contentType string
	body             []byte
}

func (s *avatarRepoStub) Put(_ context.Context, key string, body io.Reader, _ int64, contentType string) error {
	s.key, s.contentType = key, contentType
	var err error
	s.body, err = io.ReadAll(body)
	return err
}

func (*avatarRepoStub) Bucket() string { return "media" }

func TestUploadAvatarValidation(t *testing.T) {
	var pngData bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	if err := png.Encode(&pngData, img); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		file  string
		data  []byte
		valid bool
	}{
		{"valid PNG", "photo.PNG", pngData.Bytes(), true},
		{"fake PNG", "photo.png", []byte("not an image"), false},
		{"wrong extension", "photo.jpg", pngData.Bytes(), false},
		{"unsupported extension", "photo.txt", pngData.Bytes(), false},
		{"missing extension", "photo", pngData.Bytes(), false},
		{"truncated PNG", "photo.png", pngData.Bytes()[:30], false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &avatarRepoStub{}
			h := &BuyersHandler{avatarStorage: repo}
			url, err := h.uploadAvatar(context.Background(), 6, tc.file, bytes.NewReader(tc.data), int64(len(tc.data)))
			if !tc.valid {
				if !errors.Is(err, errInvalidAvatar) || repo.key != "" {
					t.Fatalf("invalid avatar: err = %v, uploaded key = %q", err, repo.key)
				}
				return
			}
			if err != nil || url == nil || *url != "/media/buyers/6/avatar-6.png" || repo.key != "buyers/6/avatar-6.png" || repo.contentType != "image/png" || !bytes.Equal(repo.body, tc.data) {
				t.Fatalf("unexpected upload: url = %v, err = %v, repo = %+v", url, err, repo)
			}
		})
	}
}
