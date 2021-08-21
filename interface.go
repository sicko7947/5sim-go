package fivesimgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	http "github.com/zMrKrabz/fhttp"
)

type Client interface {
	GetProducts(country, operator, service string) (Products, error)
	GetUserInfo() (*UserInfo, error)
	GetBalance() (float32, error)
	GetEmail() (string, error)
	GetID() (int, error)
	GetRating() (int, error)
	BuyActivationNumber(country, operator, name, forwardingNumber string) (*ActivationOrder, error)
	BuyHostingNumber(country, operator, name string) (*HostingOrder, error)
	CheckOrder(orderID int) (*ActivationOrder, error)
	FinishOrder(orderID int) (*ActivationOrder, error)
	CancelOrder(orderID int) (*ActivationOrder, error)
	BanOrder(orderID int) (*ActivationOrder, error)
}

func NewClient(apiKey string, referral string) Client {
	return &client{
		APIKey:   apiKey,
		Referral: referral,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				ForceAttemptHTTP2: true,
			},
		},
	}
}

// GetProducts returns a list of products specified by country, operator and service
// If country and/or operator are empty strings, the service will choose at random
// If service is an empty string, it will return all the available services for that country and operator
func (c *client) GetProducts(country, operator, service string) (Products, error) {
	// If country is empty, it will pass "any" to the service
	if country == "" {
		country = ANY
	}
	// If operator is empty, it will pass "any" to the service
	if operator == "" {
		operator = ANY
	}

	var resp *http.Response
	var err error
	// If service is not given, don"t ask for it
	if service == "" {
		// Make request
		resp, err = c.makeGetRequest(
			fmt.Sprintf("%s/guest/products/%s/%s", FivesimAPIEndpoint, country, operator),
			&url.Values{},
		)
	} else {
		// Make request
		resp, err = c.makeGetRequest(
			fmt.Sprintf("%s/guest/products/%s/%s/%s", FivesimAPIEndpoint, country, operator, service),
			&url.Values{},
		)
	}
	if err != nil {
		return Products{}, err
	}

	// Check status code
	if resp.StatusCode != 200 {
		return Products{}, fmt.Errorf("%s", resp.Status)
	}

	// Read request body
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return Products{}, err
	}
	resp.Body.Close()

	// Unmarshal the body into a different struct if the service is not given
	var products Products
	// If a service is not given, unmarshal the response into multiple products
	if service == "" {
		err = json.Unmarshal(r, &products)
		if err != nil {
			return Products{}, err
		}
	} else {
		// Else unmarshal it in a single product and then create a false Products map
		// that only contains a single product
		var product Product
		err = json.Unmarshal(r, &product)
		if err != nil {
			return Products{}, err
		}
		products = map[string]Product{service: product}
	}

	return products, nil
}

// GetUserInfo returns ID, Email, Balance and rating of the user in a single request
func (c *client) GetUserInfo() (*UserInfo, error) {
	// Make request
	resp, err := c.makeGetRequest(
		fmt.Sprintf("%s/user/profile", FivesimAPIEndpoint),
		&url.Values{},
	)

	if err != nil {
		return &UserInfo{}, err
	}

	// Check status code
	if resp.StatusCode != 200 {
		return &UserInfo{}, fmt.Errorf("%s", resp.Status)
	}

	// Read request body
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return &UserInfo{}, err
	}
	resp.Body.Close()

	// Unmarshal the body into a struct
	var info UserInfo
	err = json.Unmarshal(r, &info)
	if err != nil {
		return &UserInfo{}, err
	}

	return &info, nil
}

// GetBalance returns user's balance
func (c *client) GetBalance() (float32, error) {
	info, err := c.GetUserInfo()
	if err != nil {
		return 0.0, err
	}

	return info.Balance, nil
}

// GetEmail returns user's email
func (c *client) GetEmail() (string, error) {
	info, err := c.GetUserInfo()
	if err != nil {
		return "", err
	}

	return info.Email, nil
}

// GetID returns user's ID
func (c *client) GetID() (int, error) {
	info, err := c.GetUserInfo()
	if err != nil {
		return 0, err
	}

	return info.ID, nil
}

// GetRating returns user's rating
func (c *client) GetRating() (int, error) {
	info, err := c.GetUserInfo()
	if err != nil {
		return 0, err
	}

	return info.Rating, nil
}

// BuyActivationNumber performs a "buy activation number" operation by selecting country, operator and product name
// and returns the operation information
func (c *client) BuyActivationNumber(country, operator, name, forwardingNumber string) (*ActivationOrder, error) {
	// If country is empty, it will pass "any" to the service
	if country == "" {
		country = ANY
	}
	// If operator is empty, it will pass "any" to the service
	if operator == "" {
		operator = ANY
	}

	// Check if any additional query values could be encapsulated
	queryValues := url.Values{}
	if forwardingNumber != "" {
		queryValues.Add("forwarding", "1")
		queryValues.Add("number", forwardingNumber)
	}

	if c.Referral != "" {
		queryValues.Add("ref", c.Referral)
	}

	// Make request
	resp, err := c.makeGetRequest(
		fmt.Sprintf("%s/user/buy/activation/%s/%s/%s",
			FivesimAPIEndpoint, country, operator, name,
		),
		&queryValues,
	)
	if err != nil {
		return &ActivationOrder{}, err
	}

	// Check status code
	if resp.StatusCode != 200 {
		return &ActivationOrder{}, fmt.Errorf("%s", resp.Status)
	}

	// Read request body
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return &ActivationOrder{}, err
	}
	resp.Body.Close()

	// Unmarshal the body into a struct
	var info ActivationOrder
	err = json.Unmarshal(r, &info)
	if err != nil {
		return &ActivationOrder{}, err
	}

	return &info, nil
}

// BuyHostingNumber performs a "buy hosting number" operation by selecting country, operator and product name
// and returns the operation information
func (c *client) BuyHostingNumber(country, operator, name string) (*HostingOrder, error) {
	// If country is empty, it will pass "any" to the service
	if country == "" {
		country = ANY
	}
	// If operator is empty, it will pass "any" to the service
	if operator == "" {
		operator = ANY
	}

	// Make request
	resp, err := c.makeGetRequest(
		fmt.Sprintf("%s/user/buy/activation/%s/%s/%s",
			FivesimAPIEndpoint, country, operator, name,
		),
		&url.Values{},
	)
	if err != nil {
		return &HostingOrder{}, err
	}

	// Check status code
	if resp.StatusCode != 200 {
		return &HostingOrder{}, fmt.Errorf("%s", resp.Status)
	}

	// Read request body
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return &HostingOrder{}, err
	}
	resp.Body.Close()

	// Unmarshal the body into a struct
	var info HostingOrder
	err = json.Unmarshal(r, &info)
	if err != nil {
		return &HostingOrder{}, err
	}

	return &info, nil
}

// baseOrderRequest performs a customizable order request
func (c *client) baseOrderRequest(orderType string, orderID int) (*ActivationOrder, error) {
	// Make request
	resp, err := c.makeGetRequest(
		fmt.Sprintf("%s/user/%s/%d",
			FivesimAPIEndpoint, orderType, orderID,
		),
		&url.Values{},
	)
	if err != nil {
		return &ActivationOrder{}, err
	}

	// Check status code
	if resp.StatusCode != 200 {
		return &ActivationOrder{}, fmt.Errorf("%s", resp.Status)
	}

	// Read request body
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return &ActivationOrder{}, err
	}
	resp.Body.Close()

	// Unmarshal the body into a struct
	var info ActivationOrder
	err = json.Unmarshal(r, &info)
	if err != nil {
		return &ActivationOrder{}, err
	}

	return &info, nil
}

// CheckOrder checks the order status
func (c *client) CheckOrder(orderID int) (*ActivationOrder, error) {
	return c.baseOrderRequest("check", orderID)
}

// FinishOrder sets the order status as "FINISHED"
func (c *client) FinishOrder(orderID int) (*ActivationOrder, error) {
	return c.baseOrderRequest("finish", orderID)
}

// CancelOrder sets the order status as "CANCELED"
func (c *client) CancelOrder(orderID int) (*ActivationOrder, error) {
	return c.baseOrderRequest("cancel", orderID)
}

// BanOrder sets the order status as "BANNED"
func (c *client) BanOrder(orderID int) (*ActivationOrder, error) {
	return c.baseOrderRequest("ban", orderID)
}
