package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linode/linodego/internal/parseabletime"
	"github.com/linode/linodego/pkg/errors"
)

// Payment represents a Payment object
type Payment struct {
	// The unique ID of the Payment
	ID int `json:"id"`

	// The amount, in US dollars, of the Payment.
	USD json.Number `json:"usd,Number"`

	// When the Payment was made.
	Date *time.Time `json:"-"`
}

// PaymentCreateOptions fields are those accepted by CreatePayment
type PaymentCreateOptions struct {
	// CVV (Card Verification Value) of the credit card to be used for the Payment
	CVV string `json:"cvv,omitempty"`

	// The amount, in US dollars, of the Payment
	USD json.Number `json:"usd,Number"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *Payment) UnmarshalJSON(b []byte) error {
	type Mask Payment

	p := struct {
		*Mask
		Date *parseabletime.ParseableTime `json:"date"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Date = (*time.Time)(p.Date)

	return nil
}

// GetCreateOptions converts a Payment to PaymentCreateOptions for use in CreatePayment
func (i Payment) GetCreateOptions() (o PaymentCreateOptions) {
	o.USD = i.USD
	return
}

// PaymentsPagedResponse represents a paginated Payment API response
type PaymentsPagedResponse struct {
	*PageOptions
	Data []Payment `json:"data"`
}

// endpoint gets the endpoint URL for Payment
func (PaymentsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Payments.Endpoint()
	if err != nil {
		panic(err)
	}

	return endpoint
}

// appendData appends Payments when processing paginated Payment responses
func (resp *PaymentsPagedResponse) appendData(r *PaymentsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListPayments lists Payments
func (c *Client) ListPayments(ctx context.Context, opts *ListOptions) ([]Payment, error) {
	response := PaymentsPagedResponse{}
	err := c.listHelper(ctx, &response, opts)

	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetPayment gets the payment with the provided ID
func (c *Client) GetPayment(ctx context.Context, id int) (*Payment, error) {
	e, err := c.Payments.Endpoint()
	if err != nil {
		return nil, err
	}

	e = fmt.Sprintf("%s/%d", e, id)
	r, err := errors.CoupleAPIErrors(c.R(ctx).SetResult(&Payment{}).Get(e))

	if err != nil {
		return nil, err
	}

	return r.Result().(*Payment), nil
}

// CreatePayment creates a Payment
func (c *Client) CreatePayment(ctx context.Context, createOpts PaymentCreateOptions) (*Payment, error) {
	var body string

	e, err := c.Payments.Endpoint()

	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Payment{})

	if bodyData, err := json.Marshal(createOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, errors.New(err)
	}

	r, err := errors.CoupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}

	return r.Result().(*Payment), nil
}
