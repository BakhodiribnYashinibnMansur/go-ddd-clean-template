package errors

// UserMessage represents a user-friendly error message with i18n support
type UserMessage struct {
	En string `json:"en"` // English
	Uz string `json:"uz"` // Uzbek
	Ru string `json:"ru"` // Russian
}

// userMessages contains user-friendly messages for all error codes
var userMessages = map[string]UserMessage{
	// Validation Errors (400)
	ErrBadRequest: {
		En: "The request could not be processed due to invalid data.",
		Uz: "So'rov noto'g'ri ma'lumotlar tufayli qayta ishlanmadi.",
		Ru: "Запрос не может быть обработан из-за неверных данных.",
	},
	ErrInvalidInput: {
		En: "The provided input is invalid. Please check your data and try again.",
		Uz: "Kiritilgan ma'lumot noto'g'ri. Iltimos, ma'lumotlaringizni tekshiring va qayta urinib ko'ring.",
		Ru: "Предоставленные данные недействительны. Пожалуйста, проверьте данные и попробуйте снова.",
	},
	ErrValidation: {
		En: "Validation failed. Please ensure all required fields are filled correctly.",
		Uz: "Tekshirish muvaffaqiyatsiz tugadi. Iltimos, barcha majburiy maydonlar to'g'ri to'ldirilganligiga ishonch hosil qiling.",
		Ru: "Проверка не удалась. Пожалуйста, убедитесь, что все обязательные поля заполнены правильно.",
	},

	// Authentication Errors (401)
	ErrUnauthorized: {
		En: "Authentication is required to access this resource.",
		Uz: "Ushbu resursga kirish uchun autentifikatsiya talab qilinadi.",
		Ru: "Для доступа к этому ресурсу требуется аутентификация.",
	},
	ErrInvalidToken: {
		En: "The authentication token is invalid. Please log in again.",
		Uz: "Autentifikatsiya tokeni noto'g'ri. Iltimos, qayta kiring.",
		Ru: "Токен аутентификации недействителен. Пожалуйста, войдите снова.",
	},
	ErrExpiredToken: {
		En: "Your session has expired. Please log in again.",
		Uz: "Sessiyangiz muddati tugadi. Iltimos, qayta kiring.",
		Ru: "Ваша сессия истекла. Пожалуйста, войдите снова.",
	},
	ErrRevokedToken: {
		En: "Your session has been revoked. Please log in again.",
		Uz: "Sessiyangiz bekor qilindi. Iltimos, qayta kiring.",
		Ru: "Ваша сессия была отозвана. Пожалуйста, войдите снова.",
	},

	// Authorization Errors (403)
	ErrForbidden: {
		En: "You don't have permission to access this resource.",
		Uz: "Sizda ushbu resursga kirish huquqi yo'q.",
		Ru: "У вас нет разрешения на доступ к этому ресурсу.",
	},
	ErrPermissionDenied: {
		En: "You don't have the required permissions to perform this action.",
		Uz: "Sizda ushbu amalni bajarish uchun zarur ruxsatlar yo'q.",
		Ru: "У вас нет необходимых разрешений для выполнения этого действия.",
	},
	ErrDisabledAccount: {
		En: "Your account has been disabled. Please contact support.",
		Uz: "Hisobingiz o'chirilgan. Iltimos, qo'llab-quvvatlash xizmatiga murojaat qiling.",
		Ru: "Ваша учетная запись отключена. Пожалуйста, свяжитесь со службой поддержки.",
	},

	// Not Found Errors (404)
	ErrNotFound: {
		En: "The requested resource was not found.",
		Uz: "So'ralgan resurs topilmadi.",
		Ru: "Запрашиваемый ресурс не найден.",
	},
	ErrUserNotFound: {
		En: "User not found. Please check the user ID and try again.",
		Uz: "Foydalanuvchi topilmadi. Iltimos, foydalanuvchi ID sini tekshiring va qayta urinib ko'ring.",
		Ru: "Пользователь не найден. Пожалуйста, проверьте ID пользователя и попробуйте снова.",
	},
	ErrSessionNotFound: {
		En: "Session not found. Please log in again.",
		Uz: "Sessiya topilmadi. Iltimos, qayta kiring.",
		Ru: "Сессия не найдена. Пожалуйста, войдите снова.",
	},

	// Conflict Errors (409)
	ErrConflict: {
		En: "A conflict occurred. The resource you're trying to create already exists.",
		Uz: "Konflikt yuz berdi. Siz yaratmoqchi bo'lgan resurs allaqachon mavjud.",
		Ru: "Произошел конфликт. Ресурс, который вы пытаетесь создать, уже существует.",
	},
	ErrAlreadyExists: {
		En: "This resource already exists. Please use a different identifier.",
		Uz: "Bu resurs allaqachon mavjud. Iltimos, boshqa identifikatordan foydalaning.",
		Ru: "Этот ресурс уже существует. Пожалуйста, используйте другой идентификатор.",
	},

	// Server Errors (500)
	ErrInternal: {
		En: "An internal server error occurred. Our team has been notified.",
		Uz: "Ichki server xatosi yuz berdi. Bizning jamoamiz xabardor qilindi.",
		Ru: "Произошла внутренняя ошибка сервера. Наша команда была уведомлена.",
	},
	ErrDatabase: {
		En: "A database error occurred. Please try again later.",
		Uz: "Ma'lumotlar bazasi xatosi yuz berdi. Iltimos, keyinroq qayta urinib ko'ring.",
		Ru: "Произошла ошибка базы данных. Пожалуйста, попробуйте позже.",
	},
	ErrUnknown: {
		En: "An unexpected error occurred. Please try again or contact support.",
		Uz: "Kutilmagan xato yuz berdi. Iltimos, qayta urinib ko'ring yoki qo'llab-quvvatlash xizmatiga murojaat qiling.",
		Ru: "Произошла неожиданная ошибка. Пожалуйста, попробуйте снова или свяжитесь со службой поддержки.",
	},

	// Timeout Errors (504)
	ErrTimeout: {
		En: "The request timed out. Please try again.",
		Uz: "So'rov vaqti tugadi. Iltimos, qayta urinib ko'ring.",
		Ru: "Время ожидания запроса истекло. Пожалуйста, попробуйте снова.",
	},

	// Storage Errors
	ErrBucketNotFound: {
		En: "The storage bucket was not found.",
		Uz: "Saqlash qutisi topilmadi.",
		Ru: "Корзина хранилища не найдена.",
	},
	ErrFileNotFound: {
		En: "The requested file was not found.",
		Uz: "So'ralgan fayl topilmadi.",
		Ru: "Запрашиваемый файл не найден.",
	},

	// Repository Layer
	ErrRepoNotFound: {
		En: "The requested data was not found in the database.",
		Uz: "So'ralgan ma'lumot ma'lumotlar bazasida topilmadi.",
		Ru: "Запрашиваемые данные не найдены в базе данных.",
	},
	ErrRepoAlreadyExists: {
		En: "This record already exists in the database.",
		Uz: "Bu yozuv ma'lumotlar bazasida allaqachon mavjud.",
		Ru: "Эта запись уже существует в базе данных.",
	},
	ErrRepoDatabase: {
		En: "A database error occurred. Please try again later.",
		Uz: "Ma'lumotlar bazasi xatosi yuz berdi. Iltimos, keyinroq qayta urinib ko'ring.",
		Ru: "Произошла ошибка базы данных. Пожалуйста, попробуйте позже.",
	},
	ErrRepoTimeout: {
		En: "The database operation timed out. Please try again.",
		Uz: "Ma'lumotlar bazasi operatsiyasi vaqti tugadi. Iltimos, qayta urinib ko'ring.",
		Ru: "Время ожидания операции базы данных истекло. Пожалуйста, попробуйте снова.",
	},
	ErrRepoConnection: {
		En: "Unable to connect to the database. Please try again later.",
		Uz: "Ma'lumotlar bazasiga ulanib bo'lmadi. Iltimos, keyinroq qayta urinib ko'ring.",
		Ru: "Не удалось подключиться к базе данных. Пожалуйста, попробуйте позже.",
	},
	ErrRepoTransaction: {
		En: "A database transaction error occurred. Please try again.",
		Uz: "Ma'lumotlar bazasi tranzaksiyasi xatosi yuz berdi. Iltimos, qayta urinib ko'ring.",
		Ru: "Произошла ошибка транзакции базы данных. Пожалуйста, попробуйте снова.",
	},
	ErrRepoConstraint: {
		En: "The operation violates a database constraint.",
		Uz: "Operatsiya ma'lumotlar bazasi cheklovini buzadi.",
		Ru: "Операция нарушает ограничение базы данных.",
	},

	// Service Layer
	ErrServiceInvalidInput: {
		En: "The provided input is invalid. Please check your data.",
		Uz: "Kiritilgan ma'lumot noto'g'ri. Iltimos, ma'lumotlaringizni tekshiring.",
		Ru: "Предоставленные данные недействительны. Пожалуйста, проверьте данные.",
	},
	ErrServiceValidation: {
		En: "Validation failed. Please ensure all fields are correct.",
		Uz: "Tekshirish muvaffaqiyatsiz tugadi. Iltimos, barcha maydonlar to'g'riligiga ishonch hosil qiling.",
		Ru: "Проверка не удалась. Пожалуйста, убедитесь, что все поля правильные.",
	},
	ErrServiceNotFound: {
		En: "The requested resource was not found.",
		Uz: "So'ralgan resurs topilmadi.",
		Ru: "Запрашиваемый ресурс не найден.",
	},
	ErrServiceAlreadyExists: {
		En: "This resource already exists.",
		Uz: "Bu resurs allaqachon mavjud.",
		Ru: "Этот ресурс уже существует.",
	},
	ErrServiceUnauthorized: {
		En: "Authentication is required.",
		Uz: "Autentifikatsiya talab qilinadi.",
		Ru: "Требуется аутентификация.",
	},
	ErrServiceForbidden: {
		En: "You don't have permission to perform this action.",
		Uz: "Sizda ushbu amalni bajarish huquqi yo'q.",
		Ru: "У вас нет разрешения на выполнение этого действия.",
	},
	ErrServiceConflict: {
		En: "A conflict occurred with the current state of the resource.",
		Uz: "Resursning joriy holati bilan konflikt yuz berdi.",
		Ru: "Произошел конфликт с текущим состоянием ресурса.",
	},
	ErrServiceBusinessRule: {
		En: "This operation violates a business rule.",
		Uz: "Bu operatsiya biznes qoidasini buzadi.",
		Ru: "Эта операция нарушает бизнес-правило.",
	},
	ErrServiceDependency: {
		En: "A dependent service is unavailable. Please try again later.",
		Uz: "Bog'liq xizmat mavjud emas. Iltimos, keyinroq qayta urinib ko'ring.",
		Ru: "Зависимый сервис недоступен. Пожалуйста, попробуйте позже.",
	},
	ErrServiceRoleNotFound: {
		En: "The specified role was not found.",
		Uz: "Ko'rsatilgan rol topilmadi.",
		Ru: "Указанная роль не найдена.",
	},
	ErrServicePermissionNotFound: {
		En: "The specified permission was not found.",
		Uz: "Ko'rsatilgan ruxsat topilmadi.",
		Ru: "Указанное разрешение не найдено.",
	},
	ErrServicePolicyViolation: {
		En: "This action is denied by security policy.",
		Uz: "Bu amal xavfsizlik siyosati tomonidan rad etildi.",
		Ru: "Это действие запрещено политикой безопасности.",
	},
	ErrServiceScopeNotFound: {
		En: "The specified scope was not found.",
		Uz: "Ko'rsatilgan doira topilmadi.",
		Ru: "Указанная область не найдена.",
	},
}

// GetUserMessage returns user-friendly message for error code in specified language
func GetUserMessage(code, lang string) string {
	msg, ok := userMessages[code]
	if !ok {
		// Return generic message if code not found
		return getUserMessageFallback(lang)
	}

	switch lang {
	case "uz":
		if msg.Uz != "" {
			return msg.Uz
		}
	case "ru":
		if msg.Ru != "" {
			return msg.Ru
		}
	}

	// Default to English
	if msg.En != "" {
		return msg.En
	}

	return getUserMessageFallback(lang)
}

// customHTTPStatuses stores dynamically loaded HTTP statuses
var customHTTPStatuses = make(map[string]int)

// SetHTTPStatus updates the HTTP status for an error code
func SetHTTPStatus(code string, status int) {
	customHTTPStatuses[code] = status
}

// GetHTTPStatus returns the HTTP status for an error code, or 0 if not found
func GetHTTPStatus(code string) int {
	if status, ok := customHTTPStatuses[code]; ok {
		return status
	}
	return 0
}

// ErrorDetailConfig represents configuration for an error code
type ErrorDetailConfig struct {
	Message    UserMessage
	HTTPStatus int
}

// ConfigureError updates both message and status for an error code
func ConfigureError(code string, config ErrorDetailConfig) {
	if config.HTTPStatus != 0 {
		SetHTTPStatus(code, config.HTTPStatus)
	}
	// Only update message if it's not empty
	if config.Message.En != "" || config.Message.Uz != "" || config.Message.Ru != "" {
		UpdateUserMessage(code, config.Message)
	}
}

// getUserMessageFallback returns generic error message
func getUserMessageFallback(lang string) string {
	switch lang {
	case "uz":
		return "Xatolik yuz berdi. Iltimos, qayta urinib ko'ring."
	case "ru":
		return "Произошла ошибка. Пожалуйста, попробуйте снова."
	default:
		return "An error occurred. Please try again."
	}
}

// GetUserMessageWithDetails returns user message with additional details
func GetUserMessageWithDetails(code, lang, details string) string {
	msg := GetUserMessage(code, lang)
	if details != "" {
		switch lang {
		case "uz":
			return msg + " Tafsilotlar: " + details
		case "ru":
			return msg + " Детали: " + details
		default:
			return msg + " Details: " + details
		}
	}
	return msg
}

// UpdateUserMessage allows updating user message for a specific error code
// This is useful for customizing messages at runtime
func UpdateUserMessage(code string, msg UserMessage) {
	userMessages[code] = msg
}
