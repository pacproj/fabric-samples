package cc2

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Account
type SmartContract struct {
	contractapi.Contract
}

// Account describes basic details of what makes up a simple account
type Account struct {
	ID      string  `json:"ID"`
	Owner   string  `json:"owner"`
	Balance float64 `json:"balance"`
}

// InitLedger adds a base set of accounts to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	accounts := []Account{
		{ID: "account1", Owner: "Alice", Balance: 30.87},
		{ID: "account2", Owner: "Bob", Balance: 4.3},
		{ID: "account3", Owner: "Carol", Balance: 10},
		{ID: "account4", Owner: "Max", Balance: 6.192},
		{ID: "account5", Owner: "Adriana", Balance: 7},
		{ID: "account6", Owner: "Michel", Balance: 1},
	}

	for _, account := range accounts {
		accountJSON, err := json.Marshal(account)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(account.ID, accountJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAccount issues a new account to the world state with given details.
func (s *SmartContract) CreateAccount(ctx contractapi.TransactionContextInterface, id string, owner string, balance float64) error {
	exists, err := s.AccountExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the account %s already exists", id)
	}

	account := Account{
		ID:      id,
		Owner:   owner,
		Balance: balance,
	}
	accountJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, accountJSON)
}

// ReadAccount returns the account stored in the world state with given id.
func (s *SmartContract) ReadAccount(ctx contractapi.TransactionContextInterface, id string) (*Account, error) {
	accountJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accountJSON == nil {
		return nil, fmt.Errorf("the account %s does not exist", id)
	}

	var account Account
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// UpdateAccount updates an existing account in the world state with provided parameters.
func (s *SmartContract) UpdateAccount(ctx contractapi.TransactionContextInterface, id string, owner string, balance float64) error {
	exists, err := s.AccountExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the account %s does not exist", id)
	}

	// overwriting original account with new account
	account := Account{
		ID:      id,
		Owner:   owner,
		Balance: balance,
	}
	accountJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, accountJSON)
}

// DeleteAccount deletes an given account from the world state.
func (s *SmartContract) DeleteAccount(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AccountExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the account %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AccountExists returns true when account with given ID exists in world state
func (s *SmartContract) AccountExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	accountJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return accountJSON != nil, nil
}

// Transfer reduce by Value the Balance field of first given account
// and increase by Value the Balance field of second given account
func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, idFrom string, idTo string, Value float64) error {
	accountFrom, err := s.ReadAccount(ctx, idFrom)
	if err != nil {
		return err
	}

	if accountFrom.Balance-Value < 0 {
		return fmt.Errorf("The sender's balance is too small: %+v", accountFrom)
	}

	accountFrom.Balance -= Value
	accountFromJSON, err := json.Marshal(accountFrom)
	if err != nil {
		return err
	}

	accountTo, err := s.ReadAccount(ctx, idFrom)
	if err != nil {
		return err
	}

	accountTo.Balance += Value
	accountToJSON, err := json.Marshal(accountTo)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(idFrom, accountFromJSON)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(idTo, accountToJSON)
}

// GetAllAccounts returns all accounts found in world state
func (s *SmartContract) GetAllAccounts(ctx contractapi.TransactionContextInterface) ([]*Account, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all accounts in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var accounts []*Account
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var account Account
		err = json.Unmarshal(queryResponse.Value, &account)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, &account)
	}

	return accounts, nil
}
