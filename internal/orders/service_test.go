package orders

import "testing"

func TestValidateOrderInput(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateOrderInput
        wantErr error
    }{
        {
            name:    "empty order",
            input:   CreateOrderInput{},
            wantErr: ErrEmptyOrder,
        },
        {
            name: "invalid quantity",
            input: CreateOrderInput{
                Items: []OrderItemInput{
                    {ProductID: 1, Quantity: 0},
                },
            },
            wantErr: ErrInvalidOrderItem,
        },
        {
            name: "duplicate product",
            input: CreateOrderInput{
                Items: []OrderItemInput{
                    {ProductID: 1, Quantity: 2},
                    {ProductID: 1, Quantity: 3},
                },
            },
            wantErr: ErrDuplicateProduct,
        },
        {
            name: "valid order",
            input: CreateOrderInput{
                Items: []OrderItemInput{
                    {ProductID: 1, Quantity: 2},
                    {ProductID: 2, Quantity: 1},
                },
            },
            wantErr: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateOrderInput(tt.input)

            if err != tt.wantErr {
                t.Fatalf("got error %v, want %v", err, tt.wantErr)
            }
        })
    }
}