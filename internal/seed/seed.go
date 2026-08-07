// Package seed holds the demo dataset and loads it into Redis.
package seed

import (
	"context"

	"github.com/riferrei/securing-agent-secrets/internal/redisstore"
)

type Record struct {
	Customer redisstore.Customer
	PII      redisstore.PII
}

// Customers is the demo dataset, keyed by unpadded id. Business fields and PII
// live in separate structs; Apply writes them to separate keys.
var Customers = map[string]Record{
	"1": {
		redisstore.Customer{Firstname: "Ricardo", Lastname: "Ferreira", CustomerType: "Gold", CustomerSince: "2019-03-15", AccountStatus: "Active", Balance: "12450.00", Region: "North America"},
		redisstore.PII{Email: "ricardo@example.com", PhoneNumber: "+1-555-0101", SSN: "078-05-1120", Address: "123 Main St, Austin, TX"},
	},
	"2": {
		redisstore.Customer{Firstname: "Grace", Lastname: "Hopper", CustomerType: "Premium", CustomerSince: "2018-07-22", AccountStatus: "Active", Balance: "8300.50", Region: "EMEA"},
		redisstore.PII{Email: "grace@example.com", PhoneNumber: "+1-555-0102", SSN: "078-05-1121", Address: "45 Navy Yard, Arlington, VA"},
	},
	"3": {
		redisstore.Customer{Firstname: "Alan", Lastname: "Turing", CustomerType: "Silver", CustomerSince: "2020-01-10", AccountStatus: "Suspended", Balance: "210.00", Region: "EMEA"},
		redisstore.PII{Email: "alan@example.com", PhoneNumber: "+1-555-0103", SSN: "078-05-1122", Address: "78 Bletchley Rd, London"},
	},
	"4": {
		redisstore.Customer{Firstname: "Ada", Lastname: "Lovelace", CustomerType: "Gold", CustomerSince: "2017-11-02", AccountStatus: "Active", Balance: "25600.75", Region: "EMEA"},
		redisstore.PII{Email: "ada@example.com", PhoneNumber: "+1-555-0104", SSN: "078-05-1123", Address: "12 Analytical Ave, London"},
	},
	"5": {
		redisstore.Customer{Firstname: "Katherine", Lastname: "Johnson", CustomerType: "Premium", CustomerSince: "2019-06-30", AccountStatus: "Active", Balance: "15200.00", Region: "North America"},
		redisstore.PII{Email: "katherine@example.com", PhoneNumber: "+1-555-0105", SSN: "078-05-1124", Address: "34 Langley Blvd, Hampton, VA"},
	},
	"6": {
		redisstore.Customer{Firstname: "Dennis", Lastname: "Ritchie", CustomerType: "Silver", CustomerSince: "2021-02-14", AccountStatus: "Closed", Balance: "0.00", Region: "North America"},
		redisstore.PII{Email: "dennis@example.com", PhoneNumber: "+1-555-0106", SSN: "078-05-1125", Address: "9 Unix Way, Murray Hill, NJ"},
	},
	"7": {
		redisstore.Customer{Firstname: "Margaret", Lastname: "Hamilton", CustomerType: "Gold", CustomerSince: "2016-09-09", AccountStatus: "Active", Balance: "30150.25", Region: "North America"},
		redisstore.PII{Email: "margaret@example.com", PhoneNumber: "+1-555-0107", SSN: "078-05-1126", Address: "1 Apollo Dr, Cambridge, MA"},
	},
	"8": {
		redisstore.Customer{Firstname: "Ken", Lastname: "Thompson", CustomerType: "Premium", CustomerSince: "2018-04-18", AccountStatus: "Active", Balance: "9800.00", Region: "APAC"},
		redisstore.PII{Email: "ken@example.com", PhoneNumber: "+1-555-0108", SSN: "078-05-1127", Address: "5 Plan9 St, Sydney"},
	},
}

// Apply writes every demo customer into Redis, business and PII to separate
// keys. It is idempotent. Writing requires the admin user; the agent's ACL
// cannot.
func Apply(ctx context.Context, store *redisstore.Store) error {
	for id, r := range Customers {
		if err := store.SetCustomer(ctx, id, r.Customer); err != nil {
			return err
		}
		if err := store.SetPII(ctx, id, r.PII); err != nil {
			return err
		}
	}
	return nil
}
