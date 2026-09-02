package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/fileobject"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type fileRepository struct {
	client *dbent.Client
}

func NewFileRepository(client *dbent.Client) service.FileRepository {
	return &fileRepository{client: client}
}

func (r *fileRepository) Create(ctx context.Context, file *service.FileObject) (*service.FileObject, error) {
	if file == nil {
		return nil, service.ErrFileInvalidInput
	}
	client := clientFromContext(ctx, r.client)
	builder := client.FileObject.Create().
		SetOwnerUserID(file.OwnerUserID).
		SetNillableAPIKeyID(file.APIKeyID).
		SetPurpose(file.Purpose).
		SetStorageProvider(file.StorageProvider).
		SetBucket(file.Bucket).
		SetObjectKey(file.ObjectKey).
		SetNillableOriginalFilename(file.OriginalFilename).
		SetMimeType(file.MimeType).
		SetSizeBytes(file.SizeBytes).
		SetNillableSha256(file.SHA256).
		SetStatus(fileobject.Status(file.Status)).
		SetMetadata(file.Metadata).
		SetNillableUploadedAt(file.UploadedAt).
		SetNillableExpiresAt(file.ExpiresAt)

	created, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return fileObjectEntityToService(created), nil
}

func (r *fileRepository) GetByID(ctx context.Context, id int64) (*service.FileObject, error) {
	client := clientFromContext(ctx, r.client)
	file, err := client.FileObject.Query().
		Where(fileobject.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrFileNotFound
		}
		return nil, err
	}
	return fileObjectEntityToService(file), nil
}

func (r *fileRepository) UpdateStatus(ctx context.Context, id int64, status string, uploadedAt *time.Time) (*service.FileObject, error) {
	client := clientFromContext(ctx, r.client)
	builder := client.FileObject.UpdateOneID(id).
		SetStatus(fileobject.Status(status)).
		SetNillableUploadedAt(uploadedAt)

	updated, err := builder.Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrFileNotFound
		}
		return nil, err
	}
	return fileObjectEntityToService(updated), nil
}

func fileObjectEntityToService(in *dbent.FileObject) *service.FileObject {
	if in == nil {
		return nil
	}
	return &service.FileObject{
		ID:               in.ID,
		CreatedAt:        in.CreatedAt,
		UpdatedAt:        in.UpdatedAt,
		DeletedAt:        in.DeletedAt,
		OwnerUserID:      in.OwnerUserID,
		APIKeyID:         in.APIKeyID,
		Purpose:          in.Purpose,
		StorageProvider:  in.StorageProvider,
		Bucket:           in.Bucket,
		ObjectKey:        in.ObjectKey,
		OriginalFilename: in.OriginalFilename,
		MimeType:         in.MimeType,
		SizeBytes:        in.SizeBytes,
		SHA256:           in.Sha256,
		Status:           in.Status.String(),
		Metadata:         in.Metadata,
		UploadedAt:       in.UploadedAt,
		ExpiresAt:        in.ExpiresAt,
	}
}
