package mock

import (
	"time"

	"gct/internal/domain"
	"github.com/brianvoe/gofakeit/v7"
)

// FileInfo generates a fake domain.FileInfo
func FileInfo() *domain.FileInfo {
	return &domain.FileInfo{
		FileName:    gofakeit.FirstName() + "_" + gofakeit.LastName() + ".jpg",
		FileURL:     "https://example.com/files/" + UUID().String() + ".jpg",
		FileSize:    int64(gofakeit.IntRange(1024, 10485760)), // 1KB to 10MB
		ContentType: randomImageContentType(),
		UploadedAt:  time.Now(),
		BucketName:  "uploads",
	}
}

// FileInfos generates multiple fake domain.FileInfo
func FileInfos(count int) []*domain.FileInfo {
	files := make([]*domain.FileInfo, count)
	for i := range count {
		files[i] = FileInfo()
	}
	return files
}

// FileInfoWithContentType generates a fake domain.FileInfo with specific content type
func FileInfoWithContentType(contentType string) *domain.FileInfo {
	fileInfo := FileInfo()
	fileInfo.ContentType = contentType
	return fileInfo
}

// FileInfoImage generates a fake image domain.FileInfo
func FileInfoImage() *domain.FileInfo {
	return FileInfoWithContentType(randomImageContentType())
}

// FileInfoDocument generates a fake document domain.FileInfo
func FileInfoDocument() *domain.FileInfo {
	return FileInfoWithContentType(randomDocumentContentType())
}

// FileInfoVideo generates a fake video domain.FileInfo
func FileInfoVideo() *domain.FileInfo {
	return FileInfoWithContentType(randomVideoContentType())
}

// FileUploadRequest generates a fake domain.FileUploadRequest
func FileUploadRequest() *domain.FileUploadRequest {
	fileName := gofakeit.FirstName() + "_" + gofakeit.LastName() + ".jpg"
	return &domain.FileUploadRequest{
		FileName:    fileName,
		FileSize:    int64(gofakeit.IntRange(1024, 10485760)), // 1KB to 10MB
		ContentType: randomImageContentType(),
		// File field is io.Reader and cannot be mocked here
	}
}

// FileUploadRequestWithContentType generates a fake domain.FileUploadRequest with specific content type
func FileUploadRequestWithContentType(contentType string) *domain.FileUploadRequest {
	request := FileUploadRequest()
	request.ContentType = contentType
	return request
}

// randomImageContentType returns a random image content type
func randomImageContentType() string {
	contentTypes := []string{
		domain.ContentTypeJPG,
		domain.ContentTypeJPEG,
		domain.ContentTypePNG,
		domain.ContentTypeSVG,
		domain.ContentTypeHEIC,
		domain.ContentTypeHEIF,
	}
	return contentTypes[gofakeit.IntRange(0, len(contentTypes)-1)]
}

// randomDocumentContentType returns a random document content type
func randomDocumentContentType() string {
	contentTypes := []string{
		domain.ContentTypePDF,
		domain.ContentTypeDOC,
		domain.ContentTypeDOCX,
		domain.ContentTypeXLSX,
	}
	return contentTypes[gofakeit.IntRange(0, len(contentTypes)-1)]
}

// randomVideoContentType returns a random video content type
func randomVideoContentType() string {
	contentTypes := []string{
		domain.ContentTypeMP4,
		domain.ContentTypeMPEG,
		domain.ContentTypeAVI,
		domain.ContentType3GP,
		domain.ContentTypeWEBM,
	}
	return contentTypes[gofakeit.IntRange(0, len(contentTypes)-1)]
}
