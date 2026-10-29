package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"expense-app/backend/internal/model"
	"expense-app/backend/internal/storage"
)

type imageObject struct {
	metadata storage.ObjectMetadata
	body     []byte
}

type fakeImageStorage struct {
	objects   map[string]imageObject
	events    *[]string
	deleted   []string
	headCalls int
}

func (f *fakeImageStorage) PresignPut(_ context.Context, key, contentType string, size int64, _ time.Duration) (storage.PresignedPut, error) {
	return storage.PresignedPut{URL: "https://signed.invalid/put", Headers: map[string]string{"Content-Length": "123"}}, nil
}
func (f *fakeImageStorage) Head(_ context.Context, key string) (storage.ObjectMetadata, error) {
	f.headCalls++
	object, ok := f.objects[key]
	if !ok {
		return storage.ObjectMetadata{}, storage.ErrObjectNotFound
	}
	return object.metadata, nil
}
func (f *fakeImageStorage) ReadPrefix(_ context.Context, key string, limit int64) ([]byte, error) {
	object, ok := f.objects[key]
	if !ok {
		return nil, errors.New("not found")
	}
	if int64(len(object.body)) > limit {
		return object.body[:limit], nil
	}
	return object.body, nil
}
func (f *fakeImageStorage) Copy(_ context.Context, source, destination string) error {
	if f.events != nil {
		*f.events = append(*f.events, "copy")
	}
	object, ok := f.objects[source]
	if !ok {
		return errors.New("not found")
	}
	f.objects[destination] = object
	return nil
}
func (f *fakeImageStorage) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://signed.invalid/get", nil
}
func (f *fakeImageStorage) Delete(_ context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	if f.events != nil {
		*f.events = append(*f.events, "object_delete:"+key)
	}
	delete(f.objects, key)
	return nil
}

type fakeExpenseRepository struct {
	expenses    map[string]model.Expense
	events      *[]string
	createCalls int
}

func (f *fakeExpenseRepository) CreateExpense(_ context.Context, expense *model.Expense) error {
	f.createCalls++
	f.expenses[expense.ID] = *expense
	return nil
}
func (f *fakeExpenseRepository) ListExpenses(context.Context, string, *time.Time, *time.Time) ([]model.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseRepository) GetExpense(_ context.Context, _, id string) (model.Expense, error) {
	expense, ok := f.expenses[id]
	if !ok {
		return model.Expense{}, model.ErrExpenseNotFound
	}
	return expense, nil
}
func (f *fakeExpenseRepository) UpdateExpense(_ context.Context, _, id string, updates map[string]any) (model.Expense, error) {
	if f.events != nil {
		*f.events = append(*f.events, "db_update")
	}
	expense, ok := f.expenses[id]
	if !ok {
		return model.Expense{}, model.ErrExpenseNotFound
	}
	for key, value := range updates {
		if key == "image_key" {
			v := value.(string)
			expense.ImageKey = &v
		}
	}
	f.expenses[id] = expense
	return expense, nil
}
func (f *fakeExpenseRepository) DeleteExpense(_ context.Context, _, id string) error {
	if _, ok := f.expenses[id]; !ok {
		return model.ErrExpenseNotFound
	}
	if f.events != nil {
		*f.events = append(*f.events, "db_delete")
	}
	delete(f.expenses, id)
	return nil
}

func TestCreateRejectsTemporaryImageOwnedByAnotherUser(t *testing.T) {
	images := &fakeImageStorage{objects: map[string]imageObject{}}
	repo := &fakeExpenseRepository{expenses: map[string]model.Expense{}}
	svc := NewExpenseServiceWithImages(repo, images, ExpenseImageConfig{PutURLTTL: 10 * time.Minute, GetURLTTL: 15 * time.Minute})
	_, err := svc.Create(context.Background(), "user-a", CreateExpenseInput{Amount: 1, ImageKey: strPtr("tmp/users/user-b/00000000-0000-4000-8000-000000000001.png")})
	if !errors.Is(err, ErrInvalidImageKey) {
		t.Fatalf("error = %v, want invalid image key", err)
	}
	if images.headCalls != 0 || repo.createCalls != 0 {
		t.Fatalf("foreign key reached storage or DB: head=%d create=%d", images.headCalls, repo.createCalls)
	}
}

func TestPresignCreatesUserScopedTemporaryKey(t *testing.T) {
	images := &fakeImageStorage{objects: map[string]imageObject{}}
	svc := NewExpenseServiceWithImages(&fakeExpenseRepository{}, images, ExpenseImageConfig{PutURLTTL: 10 * time.Minute, GetURLTTL: 15 * time.Minute})
	result, err := svc.PresignImageUpload(context.Background(), "user-a", "image/webp", 2048)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result.ImageKey, "tmp/users/user-a/") || !strings.HasSuffix(result.ImageKey, ".webp") {
		t.Fatalf("unexpected temporary key: %q", result.ImageKey)
	}
	if result.ExpiresIn != 600 || result.UploadURL == "" || result.Headers["Content-Length"] != "123" {
		t.Fatalf("unexpected presign response: %+v", result)
	}
}

func TestDeleteExpenseDeletesImageAfterDatabaseRow(t *testing.T) {
	events := []string{}
	key := "users/user-a/00000000-0000-4000-8000-000000000001.png"
	repo := &fakeExpenseRepository{events: &events, expenses: map[string]model.Expense{
		"expense-id": {ID: "expense-id", UserID: "user-a", Amount: 1, ImageKey: &key},
	}}
	images := &fakeImageStorage{events: &events, objects: map[string]imageObject{key: {metadata: storage.ObjectMetadata{Size: 12, ContentType: "image/png"}}}}
	svc := NewExpenseServiceWithImages(repo, images, ExpenseImageConfig{PutURLTTL: 10 * time.Minute, GetURLTTL: 15 * time.Minute})
	if err := svc.Delete(context.Background(), "user-a", "expense-id"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(events, ",") != "db_delete,object_delete:"+key {
		t.Fatalf("events = %v; want DB delete then object delete", events)
	}
	if len(images.deleted) != 1 || images.deleted[0] != key {
		t.Fatalf("deleted keys = %v", images.deleted)
	}
}

func TestReplaceImageUpdatesDatabaseBeforeDeletingOldImage(t *testing.T) {
	userID := "user-a"
	oldKey := "users/user-a/00000000-0000-4000-8000-000000000001.png"
	tempKey := "tmp/users/user-a/00000000-0000-4000-8000-000000000002.png"
	events := []string{}
	repo := &fakeExpenseRepository{events: &events, expenses: map[string]model.Expense{
		"expense-id": {ID: "expense-id", UserID: userID, Amount: 100, ImageKey: &oldKey},
	}}
	png := []byte("\x89PNG\r\n\x1a\n")
	images := &fakeImageStorage{events: &events, objects: map[string]imageObject{
		oldKey:  {metadata: storage.ObjectMetadata{Size: int64(len(png)), ContentType: "image/png"}, body: png},
		tempKey: {metadata: storage.ObjectMetadata{Size: int64(len(png)), ContentType: "image/png"}, body: png},
	}}
	svc := NewExpenseServiceWithImages(repo, images, ExpenseImageConfig{PutURLTTL: 10 * time.Minute, GetURLTTL: 15 * time.Minute})
	newTempKey := tempKey
	if _, err := svc.Update(context.Background(), userID, "expense-id", UpdateExpenseInput{ImageKey: &newTempKey}); err != nil {
		t.Fatal(err)
	}
	dbUpdateIndex, oldDeleteIndex := -1, -1
	for i, event := range events {
		if event == "db_update" {
			dbUpdateIndex = i
		}
		if event == "object_delete:"+oldKey {
			oldDeleteIndex = i
		}
	}
	if dbUpdateIndex < 0 || oldDeleteIndex <= dbUpdateIndex {
		t.Fatalf("old image must be deleted after DB update, events=%v", events)
	}
}

func TestFinalizeTemporaryImageCopiesValidObjectAndDeletesTemporaryKey(t *testing.T) {
	userID := "user-a"
	temporaryKey := "tmp/users/user-a/00000000-0000-4000-8000-000000000001.png"
	png := []byte("\x89PNG\r\n\x1a\n")
	images := &fakeImageStorage{objects: map[string]imageObject{
		temporaryKey: {metadata: storage.ObjectMetadata{Size: int64(len(png)), ContentType: "image/png"}, body: png},
	}}
	svc := NewExpenseServiceWithImages(&fakeExpenseRepository{}, images, ExpenseImageConfig{PutURLTTL: 10 * time.Minute, GetURLTTL: 15 * time.Minute})
	key, err := svc.FinalizeImage(context.Background(), userID, temporaryKey)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, "users/user-a/") || strings.Contains(key, "/tmp/") {
		t.Fatalf("unexpected permanent key %q", key)
	}
	if len(images.deleted) != 1 || images.deleted[0] != temporaryKey {
		t.Fatalf("temp object not deleted: %v", images.deleted)
	}
}

func TestImageMagicTypeDoesNotTrustContentTypeAlone(t *testing.T) {
	if got := detectImageType([]byte(strings.Repeat("x", 12))); got != "" {
		t.Fatalf("detected fake image type %q", got)
	}
	if got := detectImageType([]byte("\x89PNG\r\n\x1a\n")); got != "image/png" {
		t.Fatalf("detected type = %q, want image/png", got)
	}
}

func strPtr(value string) *string { return &value }
