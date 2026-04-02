package errors

// RetryPolicy defines retry behavior for an error code.
type RetryPolicy struct {
	Retryable  bool
	RetryAfter int // seconds
	MaxRetries int
}

// domainCodeEntry holds all metadata for a single error code.
type domainCodeEntry struct {
	code        string
	numeric     string
	httpStatus  int
	severity    ErrorSeverity
	category    ErrorCategory
	retry       RetryPolicy
	serviceCode string // which service error code this maps to
	en, uz, ru  string
}

var (
	domainNumericCodes  = make(map[string]string)
	domainSeverities    = make(map[string]ErrorSeverity)
	domainCategories    = make(map[string]ErrorCategory)
	domainRetryPolicies = make(map[string]RetryPolicy)
	domainToServiceCode = make(map[string]string)
)

func init() {
	registerDomainCodes()
	registerExternalCodes()
	registerRepoRetryPolicies()
}

func registerDomainCodes() {
	codes := []domainCodeEntry{
		// ================================================================
		// User module (6001-6010)
		// ================================================================
		{"USER_NOT_FOUND", "6001", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"User not found.",
			"Foydalanuvchi topilmadi.",
			"Пользователь не найден."},
		{"USER_PHONE_EXISTS", "6002", 409, SeverityLow, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceAlreadyExists,
			"This phone number is already registered.",
			"Bu telefon raqami allaqachon ro'yxatdan o'tgan.",
			"Этот номер телефона уже зарегистрирован."},
		{"USER_INVALID_PASSWORD", "6003", 401, SeverityLow, CategorySecurity,
			RetryPolicy{false, 0, 0}, ErrServiceUnauthorized,
			"Invalid password.",
			"Noto'g'ri parol.",
			"Неверный пароль."},
		{"USER_INACTIVE", "6004", 403, SeverityMedium, CategorySecurity,
			RetryPolicy{false, 0, 0}, ErrServiceForbidden,
			"User account is inactive.",
			"Foydalanuvchi hisobi faol emas.",
			"Учетная запись пользователя неактивна."},
		{"USER_NOT_APPROVED", "6005", 403, SeverityMedium, CategorySecurity,
			RetryPolicy{false, 0, 0}, ErrServiceForbidden,
			"User account is not approved.",
			"Foydalanuvchi hisobi tasdiqlanmagan.",
			"Учетная запись пользователя не одобрена."},
		{"USER_MAX_SESSIONS", "6006", 429, SeverityLow, CategoryBusiness,
			RetryPolicy{true, 300, 1}, ErrServiceBusinessRule,
			"Maximum sessions reached. Please close an existing session.",
			"Maksimal sessiyalar soniga yetildi. Mavjud sessiyani yoping.",
			"Достигнуто максимальное количество сессий. Закройте существующую сессию."},
		{"USER_SESSION_NOT_FOUND", "6007", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Session not found.",
			"Sessiya topilmadi.",
			"Сессия не найдена."},
		{"USER_WEAK_PASSWORD", "6008", 400, SeverityLow, CategoryValidation,
			RetryPolicy{false, 0, 0}, ErrServiceValidation,
			"Password must be at least 8 characters.",
			"Parol kamida 8 ta belgidan iborat bo'lishi kerak.",
			"Пароль должен содержать не менее 8 символов."},
		{"USER_INVALID_PHONE", "6009", 400, SeverityLow, CategoryValidation,
			RetryPolicy{false, 0, 0}, ErrServiceValidation,
			"Invalid phone number format.",
			"Telefon raqami formati noto'g'ri.",
			"Неверный формат номера телефона."},
		{"USER_INVALID_EMAIL", "6010", 400, SeverityLow, CategoryValidation,
			RetryPolicy{false, 0, 0}, ErrServiceValidation,
			"Invalid email address.",
			"Email manzili noto'g'ri.",
			"Неверный адрес электронной почты."},

		// ================================================================
		// Announcement module (6011-6012)
		// ================================================================
		{"ANNOUNCEMENT_NOT_FOUND", "6011", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Announcement not found.",
			"E'lon topilmadi.",
			"Объявление не найдено."},
		{"ANNOUNCEMENT_ALREADY_PUBLISHED", "6012", 409, SeverityLow, CategoryBusiness,
			RetryPolicy{false, 0, 0}, ErrServiceConflict,
			"Announcement is already published.",
			"E'lon allaqachon nashr qilingan.",
			"Объявление уже опубликовано."},

		// ================================================================
		// FeatureFlag module (6013-6016)
		// ================================================================
		{"FEATURE_FLAG_NOT_FOUND", "6013", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Feature flag not found.",
			"Feature flag topilmadi.",
			"Флаг функции не найден."},
		{"RULE_GROUP_NOT_FOUND", "6014", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Rule group not found.",
			"Qoidalar guruhi topilmadi.",
			"Группа правил не найдена."},
		{"INVALID_OPERATOR", "6015", 400, SeverityLow, CategoryValidation,
			RetryPolicy{false, 0, 0}, ErrServiceValidation,
			"Invalid operator.",
			"Noto'g'ri operator.",
			"Недопустимый оператор."},
		{"DUPLICATE_FLAG_KEY", "6016", 409, SeverityLow, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceAlreadyExists,
			"Feature flag key already exists.",
			"Feature flag kaliti allaqachon mavjud.",
			"Ключ флага функции уже существует."},

		// ================================================================
		// Integration module (6017-6019)
		// ================================================================
		{"INTEGRATION_NOT_FOUND", "6017", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Integration not found.",
			"Integratsiya topilmadi.",
			"Интеграция не найдена."},
		{"API_KEY_NOT_FOUND", "6018", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"API key not found.",
			"API kalit topilmadi.",
			"API ключ не найден."},
		{"API_KEY_INACTIVE", "6019", 403, SeverityMedium, CategorySecurity,
			RetryPolicy{false, 0, 0}, ErrServiceForbidden,
			"API key is inactive.",
			"API kalit faol emas.",
			"API ключ неактивен."},

		// ================================================================
		// Authz module (6020-6024)
		// ================================================================
		{"AUTHZ_ROLE_NOT_FOUND", "6020", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceRoleNotFound,
			"Role not found.",
			"Rol topilmadi.",
			"Роль не найдена."},
		{"AUTHZ_PERMISSION_NOT_FOUND", "6021", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServicePermissionNotFound,
			"Permission not found.",
			"Ruxsat topilmadi.",
			"Разрешение не найдено."},
		{"AUTHZ_POLICY_NOT_FOUND", "6022", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Policy not found.",
			"Siyosat topilmadi.",
			"Политика не найдена."},
		{"AUTHZ_SCOPE_NOT_FOUND", "6023", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceScopeNotFound,
			"Scope not found.",
			"Doira topilmadi.",
			"Область не найдена."},
		{"AUTHZ_DUPLICATE_PERMISSION", "6024", 409, SeverityLow, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceAlreadyExists,
			"Permission already exists.",
			"Ruxsat allaqachon mavjud.",
			"Разрешение уже существует."},

		// ================================================================
		// File module (6025)
		// ================================================================
		{"FILE_NOT_FOUND", "6025", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"File not found.",
			"Fayl topilmadi.",
			"Файл не найден."},

		// ================================================================
		// Audit module (6026)
		// ================================================================
		{"AUDIT_LOG_NOT_FOUND", "6026", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Audit log not found.",
			"Audit log topilmadi.",
			"Журнал аудита не найден."},

		// ================================================================
		// Notification module (6027)
		// ================================================================
		{"NOTIFICATION_NOT_FOUND", "6027", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Notification not found.",
			"Bildirishnoma topilmadi.",
			"Уведомление не найдено."},

		// ================================================================
		// IPRule module (6028)
		// ================================================================
		{"IP_RULE_NOT_FOUND", "6028", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"IP rule not found.",
			"IP qoidasi topilmadi.",
			"Правило IP не найдено."},

		// ================================================================
		// RateLimit module (6029)
		// ================================================================
		{"RATE_LIMIT_NOT_FOUND", "6029", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Rate limit not found.",
			"Tezlik cheklovi topilmadi.",
			"Ограничение скорости не найдено."},

		// ================================================================
		// SiteSetting module (6030)
		// ================================================================
		{"SITE_SETTING_NOT_FOUND", "6030", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Site setting not found.",
			"Sayt sozlamasi topilmadi.",
			"Настройка сайта не найдена."},

		// ================================================================
		// Translation module (6031)
		// ================================================================
		{"TRANSLATION_NOT_FOUND", "6031", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Translation not found.",
			"Tarjima topilmadi.",
			"Перевод не найден."},

		// ================================================================
		// UserSetting module (6032)
		// ================================================================
		{"USER_SETTING_NOT_FOUND", "6032", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"User setting not found.",
			"Foydalanuvchi sozlamasi topilmadi.",
			"Настройка пользователя не найдена."},

		// ================================================================
		// DataExport module (6033)
		// ================================================================
		{"DATA_EXPORT_NOT_FOUND", "6033", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Data export not found.",
			"Ma'lumot eksporti topilmadi.",
			"Экспорт данных не найден."},

		// ================================================================
		// Metric module (6034)
		// ================================================================
		{"METRIC_NOT_FOUND", "6034", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Metric not found.",
			"Metrika topilmadi.",
			"Метрика не найдена."},

		// ================================================================
		// SystemError module (6035)
		// ================================================================
		{"SYSTEM_ERROR_NOT_FOUND", "6035", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"System error not found.",
			"Tizim xatosi topilmadi.",
			"Системная ошибка не найдена."},

		// ================================================================
		// ErrorCode module (6036)
		// ================================================================
		{"ERROR_CODE_NOT_FOUND", "6036", 404, SeverityMedium, CategoryData,
			RetryPolicy{false, 0, 0}, ErrServiceNotFound,
			"Error code not found.",
			"Xato kodi topilmadi.",
			"Код ошибки не найден."},
	}

	for _, c := range codes {
		registerCode(c)
	}
}

func registerExternalCodes() {
	codes := []domainCodeEntry{
		// ================================================================
		// Firebase FCM (7001-7010)
		// ================================================================
		{ErrExtFirebaseSendFailed, CodeExtFirebaseSendFailed, 502, SeverityHigh, CategoryExternal,
			RetryPolicy{true, 15, 3}, ErrServiceDependency,
			"Notification sending failed.",
			"Bildirishnoma yuborishda xatolik.",
			"Ошибка отправки уведомления."},
		{ErrExtFirebaseInvalidToken, CodeExtFirebaseInvalidToken, 400, SeverityLow, CategoryExternal,
			RetryPolicy{false, 0, 0}, ErrServiceValidation,
			"Invalid device token.",
			"Qurilma tokeni noto'g'ri.",
			"Недействительный токен устройства."},
		{ErrExtFirebaseQuotaExceeded, CodeExtFirebaseQuotaExceeded, 429, SeverityHigh, CategoryExternal,
			RetryPolicy{true, 60, 3}, ErrServiceDependency,
			"Notification quota exceeded.",
			"Bildirishnoma limiti oshdi.",
			"Превышен лимит уведомлений."},
		{ErrExtFirebaseUnavailable, CodeExtFirebaseUnavailable, 503, SeverityHigh, CategoryExternal,
			RetryPolicy{true, 30, 5}, ErrServiceDependency,
			"Notification service temporarily unavailable.",
			"Bildirishnoma servisi vaqtincha mavjud emas.",
			"Сервис уведомлений временно недоступен."},

		// ================================================================
		// Telegram (7011-7020)
		// ================================================================
		{ErrExtTelegramAPIError, CodeExtTelegramAPIError, 502, SeverityMedium, CategoryExternal,
			RetryPolicy{true, 10, 2}, ErrServiceDependency,
			"Telegram notification failed.",
			"Telegram xabarnoma yuborishda xatolik.",
			"Ошибка отправки Telegram уведомления."},
		{ErrExtTelegramTimeout, CodeExtTelegramTimeout, 504, SeverityMedium, CategoryExternal,
			RetryPolicy{true, 15, 3}, ErrServiceDependency,
			"Telegram request timed out.",
			"Telegram so'rov vaqti tugadi.",
			"Время ожидания запроса Telegram истекло."},
		{ErrExtTelegramConnection, CodeExtTelegramConnection, 502, SeverityHigh, CategoryExternal,
			RetryPolicy{true, 10, 3}, ErrServiceDependency,
			"Cannot connect to Telegram.",
			"Telegramga ulanib bo'lmadi.",
			"Не удалось подключиться к Telegram."},
		{ErrExtTelegramRateLimit, CodeExtTelegramRateLimit, 429, SeverityMedium, CategoryExternal,
			RetryPolicy{true, 60, 2}, ErrServiceDependency,
			"Telegram rate limit exceeded.",
			"Telegram so'rovlar limiti oshdi.",
			"Превышен лимит запросов Telegram."},

		// ================================================================
		// Asynq (7021-7030)
		// ================================================================
		{ErrExtAsynqEnqueueFailed, CodeExtAsynqEnqueueFailed, 500, SeverityHigh, CategoryExternal,
			RetryPolicy{true, 5, 3}, ErrServiceDependency,
			"Background task failed to queue.",
			"Fon vazifasi navbatga qo'yilmadi.",
			"Фоновая задача не была поставлена в очередь."},
		{ErrExtAsynqConnection, CodeExtAsynqConnection, 500, SeverityCritical, CategoryExternal,
			RetryPolicy{true, 10, 5}, ErrServiceDependency,
			"Task queue connection failed.",
			"Vazifalar navbatiga ulanishda xatolik.",
			"Ошибка подключения к очереди задач."},
		{ErrExtAsynqTimeout, CodeExtAsynqTimeout, 504, SeverityHigh, CategoryExternal,
			RetryPolicy{true, 5, 3}, ErrServiceDependency,
			"Task queue operation timed out.",
			"Vazifalar navbati operatsiyasi vaqti tugadi.",
			"Время ожидания операции очереди задач истекло."},
		{ErrExtAsynqPayloadError, CodeExtAsynqPayloadError, 400, SeverityLow, CategoryValidation,
			RetryPolicy{false, 0, 0}, ErrServiceValidation,
			"Invalid task payload.",
			"Vazifa ma'lumotlari noto'g'ri.",
			"Недопустимые данные задачи."},

		// ================================================================
		// EventBus (7031-7040)
		// ================================================================
		{ErrExtEventBusPublishFailed, CodeExtEventBusPublishFailed, 500, SeverityHigh, CategoryExternal,
			RetryPolicy{true, 5, 3}, ErrServiceDependency,
			"Event publish failed.",
			"Hodisa nashr qilishda xatolik.",
			"Ошибка публикации события."},
		{ErrExtEventBusConnection, CodeExtEventBusConnection, 500, SeverityCritical, CategoryExternal,
			RetryPolicy{true, 10, 5}, ErrServiceDependency,
			"Event bus connection failed.",
			"Hodisalar magistraliga ulanishda xatolik.",
			"Ошибка подключения к шине событий."},
		{ErrExtEventBusTimeout, CodeExtEventBusTimeout, 504, SeverityHigh, CategoryExternal,
			RetryPolicy{true, 5, 3}, ErrServiceDependency,
			"Event bus operation timed out.",
			"Hodisalar magistrali operatsiyasi vaqti tugadi.",
			"Время ожидания шины событий истекло."},
	}

	for _, c := range codes {
		registerCode(c)
	}
}

// registerRepoRetryPolicies adds retry policies for existing repo/service/handler codes.
func registerRepoRetryPolicies() {
	policies := map[string]RetryPolicy{
		// Repo layer — retryable
		ErrRepoTimeout:    {true, 5, 3},
		ErrRepoConnection: {true, 10, 5},

		// Service layer
		ErrServiceDependency: {true, 10, 3},

		// Handler layer
		ErrHandlerServiceUnavailable: {true, 30, 5},
		ErrHandlerTooManyRequests:    {true, 60, 3},
		ErrHandlerInternal:           {true, 10, 3},
	}

	for code, policy := range policies {
		domainRetryPolicies[code] = policy
	}
}

func registerCode(c domainCodeEntry) {
	ConfigureError(c.code, ErrorDetailConfig{
		Message:    UserMessage{En: c.en, Uz: c.uz, Ru: c.ru},
		HTTPStatus: c.httpStatus,
	})
	domainNumericCodes[c.code] = c.numeric
	domainSeverities[c.code] = c.severity
	domainCategories[c.code] = c.category
	domainRetryPolicies[c.code] = c.retry
	if c.serviceCode != "" {
		domainToServiceCode[c.code] = c.serviceCode
	}
}

// GetRetryPolicy returns retry policy for an error code.
func GetRetryPolicy(code string) RetryPolicy {
	if p, ok := domainRetryPolicies[code]; ok {
		return p
	}
	if IsRetryable(code) {
		return RetryPolicy{Retryable: true, RetryAfter: 10, MaxRetries: 3}
	}
	return RetryPolicy{}
}

// GetDomainNumericCode returns numeric code for a domain error code, or empty string.
func GetDomainNumericCode(code string) string {
	return domainNumericCodes[code]
}

// GetDomainServiceCode returns the service error code a domain code maps to.
func GetDomainServiceCode(code string) string {
	if sc, ok := domainToServiceCode[code]; ok {
		return sc
	}
	return ErrServiceUnknown
}
