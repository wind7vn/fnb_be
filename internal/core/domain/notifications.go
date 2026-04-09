package domain

const (
	NotiTypeSystem        = "SYSTEM"
	NotiTypeNewOrder      = "NEW_ORDER"
	NotiTypeOrderUpdate   = "ORDER_UPDATE"
	NotiTypeCallStaff     = "CALL_STAFF"
	NotiTypePaymentRecall = "PAYMENT_RECALL"
)

const (
	TopicStaff   = "staff_notifications"
	TopicKitchen = "kitchen_notifications"
)

type PushNotificationPayload struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data"`
	TargetID string            `json:"target_id,omitempty"`
}
