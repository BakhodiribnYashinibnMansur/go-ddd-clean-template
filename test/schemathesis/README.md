# Schemathesis API Testing

Schemathesis - OpenAPI/GraphQL schema'dan avtomatik test case'lar yaratadigan va API'dagi bug'larni topadigan vosita.

## O'rnatish

```bash
# Python orqali
pip install schemathesis

# Yoki uv orqali (tavsiya etiladi)
uv pip install schemathesis
```

## Foydalanish

### 1. CLI orqali (Tez test)

```bash
# API'ni ishga tushiring
make run

# Boshqa terminalda test'larni ishga tushiring
cd test/schemathesis
chmod +x run_tests.sh
./run_tests.sh
```

### 2. Python script orqali (To'liq test)

```bash
# Test faylini ishga tushiring
pytest test_api.py -v

# Yoki to'g'ridan-to'g'ri
python test_api.py
```

### 3. Bitta endpoint'ni test qilish

```bash
schemathesis run http://localhost:8080/swagger/doc.json \
    --base-url=http://localhost:8080 \
    --endpoint="/api/v1/users" \
    --method=POST \
    --checks=all
```

## Qanday test'lar bajariladi?

### ✅ Avtomatik tekshiruvlar:

1. **500 Errors** - Server crash bo'ladigan edge case'lar
2. **Schema Validation** - Response schema'ga mos keladimi
3. **Status Codes** - To'g'ri HTTP status code qaytaradimi
4. **Content-Type** - To'g'ri content type bormi
5. **Response Time** - Response tez qaytadimi

### 🔄 Stateful Testing:

Realistic workflow'larni test qiladi:
- User yaratish → Olish → Yangilash → O'chirish
- Sign in → Session yaratish → Session olish → Revoke qilish
- Policy yaratish → Role'ga biriktirish → Authorization test

## Konfiguratsiya

### Environment variables:

```bash
export API_URL=http://localhost:8080
export SCHEMA_PATH=http://localhost:8080/swagger/doc.json
export MAX_EXAMPLES=100  # Har bir endpoint uchun test case'lar soni
export WORKERS=4         # Parallel worker'lar soni
```

### Test sozlamalari:

`test_api.py` faylida:
- `max_examples` - Test case'lar soni
- `deadline` - Har bir test uchun timeout
- `phases` - Test fazalari

## Authentication

Protected endpoint'lar uchun token qo'shish:

```python
# test_api.py da
@schema.hooks.before_call
def add_auth_header(context, case):
    if "/auth/" not in case.path:
        case.headers = case.headers or {}
        case.headers["Authorization"] = "Bearer YOUR_TOKEN"
```

Yoki CLI'da:

```bash
schemathesis run http://localhost:8080/swagger/doc.json \
    --base-url=http://localhost:8080 \
    --header="Authorization: Bearer YOUR_TOKEN"
```

## CI/CD Integration

### GitHub Actions:

```yaml
- name: Run Schemathesis Tests
  run: |
    pip install schemathesis
    schemathesis run http://localhost:8080/swagger/doc.json \
      --base-url=http://localhost:8080 \
      --checks=all \
      --hypothesis-max-examples=50
```

### Makefile:

```makefile
.PHONY: test-api
test-api:
	@echo "Running Schemathesis tests..."
	@cd test/schemathesis && ./run_tests.sh
```

## Natijalarni tahlil qilish

### Muvaffaqiyatli test:

```
✅ GET /api/v1/users/{user_id} - PASSED
   - 50 test cases generated
   - All responses valid
   - No schema violations
```

### Xato topilgan:

```
❌ POST /api/v1/users - FAILED
   Status: 500
   Body: {"username": "a", "phone": "+1"}
   Error: Internal Server Error
   
   Issue: Username minimum length validation not working
```

## Foydali buyruqlar

```bash
# Faqat GET endpoint'larni test qilish
schemathesis run $SCHEMA --method=GET

# Ma'lum tag'li endpoint'larni test qilish
schemathesis run $SCHEMA --tag=users

# Deprecated endpoint'larni o'tkazib yuborish
schemathesis run $SCHEMA --exclude-deprecated

# Birinchi xatoda to'xtash
schemathesis run $SCHEMA --exitfirst

# Batafsil log
schemathesis run $SCHEMA --show-errors-tracebacks -v
```

## Qo'shimcha ma'lumot

- [Schemathesis Documentation](https://schemathesis.readthedocs.io/)
- [GitHub Repository](https://github.com/schemathesis/schemathesis)
- [Live Benchmarks](https://workbench.schemathesis.io)

## Muammolarni hal qilish

### API ishlamayapti:
```bash
# API holatini tekshiring
curl http://localhost:8080/health
```

### Schema topilmayapti:
```bash
# Schema mavjudligini tekshiring
curl http://localhost:8080/swagger/doc.json
```

### Token muammosi:
```bash
# Avval sign-in qiling va token oling
curl -X POST http://localhost:8080/api/v1/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{"phone": "+998901234567", "password": "password"}'
```
