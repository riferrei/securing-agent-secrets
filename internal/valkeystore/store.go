// Package valkeystore provides typed access to customer records in Valkey.
package valkeystore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/valkey-io/valkey-go"
)

var ErrNotFound = errors.New("customer not found")

type Customer struct {
	Firstname     string `json:"firstname"`
	Lastname      string `json:"lastname"`
	CustomerType  string `json:"customer_type"`
	CustomerSince string `json:"customer_since"`
	AccountStatus string `json:"account_status"`
	Balance       string `json:"balance"`
	Region        string `json:"region"`
}

type PII struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	SSN         string `json:"ssn"`
	Address     string `json:"address"`
}

type Store struct {
	vk valkey.Client
}

func New(ctx context.Context, host, port, user, password string) (*Store, error) {
	vk, err := valkey.NewClient(valkey.ClientOption{
		InitAddress:  []string{net.JoinHostPort(host, port)},
		Username:     user,
		Password:     password,
		DisableCache: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to valkey: %w", err)
	}
	if err := vk.Do(ctx, vk.B().Ping().Build()).Error(); err != nil {
		vk.Close()
		return nil, fmt.Errorf("connecting to valkey: %w", err)
	}
	return &Store{vk: vk}, nil
}

func (s *Store) Close() error { s.vk.Close(); return nil }

// Key zero-pads a customer id to four digits: "1" -> "customer:0001".
func Key(id string) string { return "customer:" + NormalizeID(id) }

// PIIKey is the sibling key holding the sensitive fields the agent's ACL denies.
func PIIKey(id string) string { return "pii:" + NormalizeID(id) }

func (s *Store) SetCustomer(ctx context.Context, id string, c Customer) error {
	return s.vk.Do(ctx, s.vk.B().Hset().Key(Key(id)).FieldValue().
		FieldValue("firstname", c.Firstname).
		FieldValue("lastname", c.Lastname).
		FieldValue("customer_type", c.CustomerType).
		FieldValue("customer_since", c.CustomerSince).
		FieldValue("account_status", c.AccountStatus).
		FieldValue("balance", c.Balance).
		FieldValue("region", c.Region).
		Build()).Error()
}

func (s *Store) SetPII(ctx context.Context, id string, p PII) error {
	return s.vk.Do(ctx, s.vk.B().Hset().Key(PIIKey(id)).FieldValue().
		FieldValue("email", p.Email).
		FieldValue("phone_number", p.PhoneNumber).
		FieldValue("ssn", p.SSN).
		FieldValue("address", p.Address).
		Build()).Error()
}

func (s *Store) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	h, err := s.vk.Do(ctx, s.vk.B().Hgetall().Key(Key(id)).Build()).AsStrMap()
	if err != nil {
		return nil, err
	}
	if len(h) == 0 {
		return nil, ErrNotFound
	}
	return &Customer{
		Firstname:     h["firstname"],
		Lastname:      h["lastname"],
		CustomerType:  h["customer_type"],
		CustomerSince: h["customer_since"],
		AccountStatus: h["account_status"],
		Balance:       h["balance"],
		Region:        h["region"],
	}, nil
}

func (s *Store) GetCustomerRaw(ctx context.Context, id string) (string, error) {
	c, err := s.GetCustomer(ctx, id)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Store) Exists(ctx context.Context, id string) (bool, error) {
	n, err := s.vk.Do(ctx, s.vk.B().Exists().Key(Key(id)).Build()).AsInt64()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
