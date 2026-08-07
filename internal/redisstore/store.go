// Package redisstore provides typed access to customer records in Redis.
package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
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
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone_number"`
	SSN           string `json:"ssn"`
	Address       string `json:"address"`
}

type Store struct {
	rdb *redis.Client
}

func New(ctx context.Context, redisURL string) (*Store, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}
	return &Store{rdb: rdb}, nil
}

func (s *Store) Close() error { return s.rdb.Close() }

// Key zero-pads a customer id to four digits: "1" -> "customer:0001".
func Key(id string) string { return "customer:" + NormalizeID(id) }

func (s *Store) SetCustomer(ctx context.Context, id string, c Customer) error {
	return s.rdb.HSet(ctx, Key(id), map[string]any{
		"firstname":      c.Firstname,
		"lastname":       c.Lastname,
		"customer_type":  c.CustomerType,
		"customer_since": c.CustomerSince,
		"account_status": c.AccountStatus,
		"balance":        c.Balance,
		"region":         c.Region,
		"email":          c.Email,
		"phone_number":   c.PhoneNumber,
		"ssn":            c.SSN,
		"address":        c.Address,
	}).Err()
}

func (s *Store) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	h, err := s.rdb.HGetAll(ctx, Key(id)).Result()
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
		Email:         h["email"],
		PhoneNumber:   h["phone_number"],
		SSN:           h["ssn"],
		Address:       h["address"],
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
	n, err := s.rdb.Exists(ctx, Key(id)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
