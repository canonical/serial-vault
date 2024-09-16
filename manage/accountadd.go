package manage

import (
	"fmt"
	"github.com/CanonicalLtd/serial-vault/datastore"
)

// AccountAddCommand handles adding a new account for the serial-vault-admin command
type AccountAddCommand struct {
	ResellerAPI bool `short:"r" long:"reseller" description:"Enable the reseller API"`
}

// Execute the adding of an account
func (cmd AccountAddCommand) Execute(args []string) error {
	err := checkAccountIDArg(args, "Add")
	if err != nil {
		return err
	}

	// Open the database and create the account
	openDatabase()
	account := datastore.Account{
		AuthorityID: args[0],
		ResellerAPI: cmd.ResellerAPI,
	}
	if err := datastore.Environ.DB.CreateAccount(account); err != nil {
		return fmt.Errorf("error creating the account: %v", err)
	}

	fmt.Printf("Account '%s' created successfully\n", account.AuthorityID)
	return nil
}

func checkAccountIDArg(args []string, action string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("%s account expects an 'account ID' argument", action)
	case 1:
		return nil
	default:
		return fmt.Errorf("%s account expects a single 'account ID' argument", action)
	}
}
