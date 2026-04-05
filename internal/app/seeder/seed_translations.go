package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//nolint:funlen // static data table: predefined translation seed records
func (s *Seeder) seedTranslations(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding translations...", zap.Int("count", count))

	now := time.Now()

	type translationSet struct {
		entityType string
		key        string
		uz         string
		ru         string
		en         string
	}

	predefined := []translationSet{
		{"ui", "auth.login", "Kirish", "Вход", "Login"},
		{"ui", "auth.logout", "Chiqish", "Выход", "Logout"},
		{"ui", "auth.register", "Ro'yxatdan o'tish", "Регистрация", "Register"},
		{"ui", "error.not_found", "Topilmadi", "Не найдено", "Not Found"},
		{"ui", "error.forbidden", "Ruxsat berilmagan", "Доступ запрещён", "Forbidden"},
		{"ui", "error.server_error", "Server xatosi", "Ошибка сервера", "Server Error"},
		{"ui", "nav.dashboard", "Boshqaruv paneli", "Панель управления", "Dashboard"},
		{"ui", "nav.settings", "Sozlamalar", "Настройки", "Settings"},
		{"ui", "nav.users", "Foydalanuvchilar", "Пользователи", "Users"},
		{"ui", "action.save", "Saqlash", "Сохранить", "Save"},
		{"ui", "action.cancel", "Bekor qilish", "Отмена", "Cancel"},
		{"ui", "action.delete", "O'chirish", "Удалить", "Delete"},
		{"ui", "action.search", "Qidirish", "Поиск", "Search"},
		{"ui", "message.success", "Muvaffaqiyatli", "Успешно", "Success"},
		{"ui", "message.confirm_delete", "O'chirishni tasdiqlaysizmi?", "Подтвердите удаление?", "Confirm deletion?"},
	}

	langs := []string{"uz", "ru", "en"}

	for _, t := range predefined {
		entityID := uuid.New()
		translations := map[string]string{"uz": t.uz, "ru": t.ru, "en": t.en}
		for _, lang := range langs {
			data := fmt.Sprintf(`{"%s":"%s"}`, t.key, translations[lang])
			_, err := s.pool.Exec(ctx,
				`INSERT INTO translations (id, entity_type, entity_id, lang_code, data, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				uuid.New(), t.entityType, entityID, lang, data, now, now,
			)
			if err != nil {
				s.logger.Warnc(ctx, "Failed to create predefined translation", zap.Error(err), zap.String("key", t.key))
			}
		}
	}

	entityTypes := []string{"ui", "email", "notification", "error"}

	remaining := count - len(predefined)*len(langs)
	for i := 0; i < remaining; i++ {
		if i < 0 {
			break
		}
		entityID := uuid.New()
		entityType := entityTypes[gofakeit.Number(0, len(entityTypes)-1)]
		lang := langs[gofakeit.Number(0, len(langs)-1)]
		key := fmt.Sprintf("%s.%s_%d", entityType, gofakeit.Word(), i)
		data := fmt.Sprintf(`{"%s":"%s"}`, key, gofakeit.Sentence(3))

		_, err := s.pool.Exec(ctx,
			`INSERT INTO translations (id, entity_type, entity_id, lang_code, data, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), entityType, entityID, lang, data, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random translation", zap.Error(err))
		}
	}

	return nil
}
