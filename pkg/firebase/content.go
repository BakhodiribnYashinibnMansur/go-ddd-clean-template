package firebase

type Content struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

const (
	En = "en"
	Ru = "ru"
	Uz = "uz"
)

var OrderNewMap = map[string]Content{
	Uz: OrderNewUz,
	Ru: OrderNewRu,
	En: OrderNewEn,
}
var OrderNewUz = Content{
	Title: "Yangi buyurtma mavjud",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var OrderNewRu = Content{
	Title: "Новый заказ",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var OrderNewEn = Content{
	Title: "New Order",
	Body:  "Click on the notification to view",
}
var OrderStatusPreparingMap = map[string]Content{
	Uz: OrderStatusPreparingUz,
	Ru: OrderStatusPreparingRu,
	En: OrderStatusPreparingEn,
}
var OrderStatusPreparingUz = Content{
	Title: "Buyurtma tayyorlash boshlandi",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var OrderStatusPreparingRu = Content{
	Title: "Заказ начинается готовиться",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var OrderStatusPreparingEn = Content{
	Title: "Order beginning to prepare",
	Body:  "Click on the notification to view",
}
var OrderStatusPackagingMap = map[string]Content{
	Uz: OrderPackagingUz,
	Ru: OrderPackagingRu,
	En: OrderPackagingEn,
}
var OrderPackagingUz = Content{
	Title: "Buyurtma qadoqlanayapti",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var OrderPackagingRu = Content{
	Title: "Заказ упаковывается",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var OrderPackagingEn = Content{
	Title: "Order is packaging",
	Body:  "Click on the notification to view",
}
var OrderStatusDeliveringMap = map[string]Content{
	Uz: OrderStatusDeliveringUz,
	Ru: OrderStatusDeliveringRu,
	En: OrderStatusDeliveringEn,
}
var OrderStatusDeliveringUz = Content{
	Title: "Buyurtma yetkazish boshlandi",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var OrderStatusDeliveringRu = Content{
	Title: "Заказ начинается доставляться",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var OrderStatusDeliveringEn = Content{
	Title: "Order beginning to deliver",
	Body:  "Click on the notification to view",
}

var OrderStatusDeliveredMap = map[string]Content{
	Uz: OrderStatusDeliveredUz,
	Ru: OrderStatusDeliveredRu,
	En: OrderStatusDeliveredEn,
}
var OrderStatusDeliveredUz = Content{
	Title: "Buyurtma manzilga yetib keldi ",
	Body:  "Marhamat qilib buyurtmani olishingiz mumkin",
}
var OrderStatusDeliveredRu = Content{
	Title: "Заказ доставлен на адрес",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var OrderStatusDeliveredEn = Content{
	Title: "Order delivered to the address",
	Body:  "Click on the notification to view",
}

var OrderStatusSuccessMap = map[string]Content{
	Uz: OrderSuccessUz,
	Ru: OrderSuccessRu,
	En: OrderSuccessEn,
}
var OrderSuccessUz = Content{
	Title: "Buyurtma muvaffaqiyatli yakunlandi",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var OrderSuccessRu = Content{
	Title: "Заказ завершен успешно",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var OrderSuccessEn = Content{
	Title: "Order completed successfully",
	Body:  "Click on the notification to view",
}
var OrderCanceledMap = map[string]Content{
	Uz: OrderCanceledUz,
	Ru: OrderCanceledRu,
	En: OrderCanceledEn,
}
var OrderCanceledUz = Content{
	Title: "Buyurtma bekor qilindi",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var OrderCanceledRu = Content{
	Title: "Заказ отменен",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var OrderCanceledEn = Content{
	Title: "Order canceled",
	Body:  "Click on the notification to view",
}

var OrderReturnMap = map[string]Content{
	Uz: OrderReturnUz,
	Ru: OrderReturnRu,
	En: OrderReturnEn,
}
var OrderReturnUz = Content{
	Title: "Buyurtma qaytarildi",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var OrderReturnRu = Content{
	Title: "Заказ возвращен",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var OrderReturnEn = Content{
	Title: "Order returned",
	Body:  "Click on the notification to view",
}
var DeliverLookingUpUz = Content{
	Title: "Courier qidirilayapti",
	Body:  "Ko'proq ma'lumot olish uchun eslatmani bosing",
}
var DeliverLookingUpRu = Content{
	Title: "Курьер ищется",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var DeliverLookingUpEn = Content{
	Title: "Courier is looking up",
	Body:  "Click on the notification to view",
}

var DeliverFoundMap = map[string]Content{
	Uz: DeliverFoundUz,
	Ru: DeliverFoundRu,
	En: DeliverFoundEn,
}
var DeliverFoundUz = Content{
	Title: "Yetkazib berish topildi",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var DeliverFoundRu = Content{
	Title: "Курьер найден",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var DeliverFoundEn = Content{
	Title: "Courier found",
	Body:  "Click on the notification to view",
}
var DeliverArrivedMap = map[string]Content{
	Uz: DeliverArrivedUz,
	Ru: DeliverArrivedRu,
	En: DeliverArrivedEn,
}
var DeliverArrivedUz = Content{
	Title: "Courier manzilga yetib keldi",
	Body:  "Ko'rish uchun eslatmani bosing",
}
var DeliverArrivedRu = Content{
	Title: "Курьер прибыл на адрес",
	Body:  "Нажмите на уведомление, чтобы просмотреть",
}
var DeliverArrivedEn = Content{
	Title: "Courier arrived at the address",
	Body:  "Click on the notification to view",
}
