package webhooks

import "testing"

func TestValidateCapturedPayment(t *testing.T) {
	tests := []struct {
		name    string
		payment PaymentEntity
		wantErr bool
	}{
		{
			name: "valid captured payment",
			payment: PaymentEntity{
				ID:       "pay_test123",
				OrderID:  "order_test123",
				Amount:   190000,
				Currency: "INR",
				Status:   "captured",
				Captured: true,
			},
			wantErr: false,
		},
		{
			name: "missing payment ID",
			payment: PaymentEntity{
				OrderID:  "order_test123",
				Amount:   190000,
				Currency: "INR",
				Status:   "captured",
				Captured: true,
			},
			wantErr: true,
		},
		{
			name: "payment not captured",
			payment: PaymentEntity{
				ID:       "pay_test123",
				OrderID:  "order_test123",
				Amount:   190000,
				Currency: "INR",
				Status:   "failed",
				Captured: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCapturedPayment(tt.payment)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateCapturedPayment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}