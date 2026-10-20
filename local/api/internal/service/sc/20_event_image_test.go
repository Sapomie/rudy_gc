package sc

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

type scEventImageWriteStub struct {
	moviex.ScEventWriteModel
	imagePath *string
	err       error
	committed bool
}

func (s *scEventImageWriteStub) UpdateEventMeta(_ context.Context, event *moviex.GSc, _ int64, now int64, imagePath *string) ([]*moviex.GList, *moviex.GSc, error) {
	s.imagePath = imagePath
	if s.err != nil && !s.committed {
		return nil, nil, s.err
	}
	updated := *event
	updated.UpdatedOn = now
	if imagePath != nil {
		updated.ImagePath = *imagePath
	}
	return nil, &updated, s.err
}

func scEventImageFixture(t *testing.T) []byte {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func scEventImageTestService(t *testing.T, writer *scEventImageWriteStub, oldImage string) (*Service, string) {
	t.Helper()
	dir := t.TempDir()
	deps := scOptionTestDeps(nil, nil, nil)
	deps.Config.SCEvent.ActiveImageDir = dir
	deps.ScModel = &updateEventScModelStub{row: &moviex.GSc{Id: 42, Name: "event-a", ImagePath: oldImage}}
	deps.GListModel = &updateEventGListModelStub{}
	deps.ScEventWriteModel = writer
	return NewService(deps), dir
}

func TestUpdateEventMetaAddsImage(t *testing.T) {
	writer := &scEventImageWriteStub{}
	service, dir := scEventImageTestService(t, writer, "")
	imageData := scEventImageFixture(t)
	updated, err := service.UpdateEventMeta(context.Background(), &types.ScEventEditForm{
		Name: "event-a", Remarks: "补图", Image: bytes.NewReader(imageData),
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ImagePath == "" || filepath.Dir(updated.ImagePath) != dir || filepath.Ext(updated.ImagePath) != ".png" {
		t.Fatalf("image path = %q", updated.ImagePath)
	}
	if writer.imagePath == nil || *writer.imagePath != updated.ImagePath || updated.Remarks != "补图" {
		t.Fatalf("saved event = %#v, image argument = %#v", updated, writer.imagePath)
	}
	got, err := os.ReadFile(updated.ImagePath)
	if err != nil || !bytes.Equal(got, imageData) {
		t.Fatalf("saved image = %v, err = %v", got, err)
	}
}

func TestUpdateEventMetaImageFailureAndReplacement(t *testing.T) {
	for _, scenario := range []string{"replace", "metadata-only", "database-failure", "cache-failure", "invalid-image", "copy-failure"} {
		t.Run(scenario, func(t *testing.T) {
			writer := &scEventImageWriteStub{}
			service, dir := scEventImageTestService(t, writer, "")
			oldPath := filepath.Join(dir, "original.png")
			oldData := []byte("original image")
			if err := os.WriteFile(oldPath, oldData, 0o644); err != nil {
				t.Fatal(err)
			}
			service.deps.ScModel.(*updateEventScModelStub).row.ImagePath = oldPath
			var input io.Reader = bytes.NewReader(scEventImageFixture(t))
			switch scenario {
			case "metadata-only":
				input = nil
			case "database-failure":
				writer.err = errors.New("database failed")
			case "cache-failure":
				writer.err = errors.New("cache failed")
				writer.committed = true
			case "invalid-image":
				input = bytes.NewBufferString("plain text")
			case "copy-failure":
				data := append(scEventImageFixture(t), make([]byte, 600)...)
				input = io.MultiReader(bytes.NewReader(data), iotest.ErrReader(errors.New("read failed")))
			}
			updated, err := service.UpdateEventMeta(context.Background(), &types.ScEventEditForm{Name: "event-a", Image: input})
			wantError := scenario != "replace" && scenario != "metadata-only"
			if (err != nil) != wantError {
				t.Fatalf("error = %v", err)
			}
			got, readErr := os.ReadFile(oldPath)
			if readErr != nil || !bytes.Equal(got, oldData) {
				t.Fatalf("original changed: %v, err = %v", got, readErr)
			}
			entries, readErr := os.ReadDir(dir)
			if readErr != nil {
				t.Fatal(readErr)
			}
			wantFiles := 1
			if scenario == "replace" || scenario == "cache-failure" {
				wantFiles = 2
				if writer.imagePath == nil || *writer.imagePath == oldPath {
					t.Fatal("new image must use an independent path")
				}
				if _, readErr := os.Stat(*writer.imagePath); readErr != nil {
					t.Fatalf("committed image missing: %v", readErr)
				}
			}
			if len(entries) != wantFiles {
				t.Fatalf("files = %d, want %d", len(entries), wantFiles)
			}
			if scenario == "metadata-only" && (writer.imagePath != nil || updated.ImagePath != oldPath) {
				t.Fatal("metadata update changed the image")
			}
		})
	}
}
