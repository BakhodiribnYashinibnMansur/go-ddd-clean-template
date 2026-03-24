package asynq

// Task type constants.
const (
	// Image processing tasks
	TypeImageResize    = "image:resize"
	TypeImageOptimize  = "image:optimize"
	TypeImageThumbnail = "image:thumbnail"

	// File processing tasks
	TypeFileUpload   = "file:upload"
	TypeFileDelete   = "file:delete"
	TypeFileCompress = "file:compress"

	// Notification tasks
	TypePushNotification = "notification:push"
	TypeSMSNotification  = "notification:sms"

	// Report generation tasks
	TypeReportGenerate = "report:generate"
	TypeReportExport   = "report:export"

	// Cleanup tasks
	TypeCleanupOldSessions = "cleanup:old_sessions"
	TypeCleanupTempFiles   = "cleanup:temp_files"

	// System tasks
	TypeSystemSeed = "system:seed"

	// Audit tasks
	TypeAuditLog = "audit:log"
)

// Queue name constants.
const (
	QueueCritical = "critical" // For time-sensitive tasks
	QueueDefault  = "default"  // For normal priority tasks
	QueueLow      = "low"      // For background tasks
)
