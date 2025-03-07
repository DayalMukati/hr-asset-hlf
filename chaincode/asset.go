package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract for Asset Tokenization
type SmartContract struct {
	contractapi.Contract
}

// Token represents a tokenized asset
type Token struct {
	AssetID   string `json:"assetID"`
	Owner     string `json:"owner"`
	Amount    int    `json:"amount"`
}

// MintTokens creates new tokens for an asset
func (s *SmartContract) MintTokens(ctx contractapi.TransactionContextInterface, assetID string, amountStr string, owner string) error {
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount")
	}

	token := Token{
		AssetID: assetID,
		Owner:   owner,
		Amount:  amount,
	}

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, tokenJSON)
}

// TransferTokens transfers tokens from one owner to another
func (s *SmartContract) TransferTokens(ctx contractapi.TransactionContextInterface, assetID string, transferAmountStr string, newOwner string) error {
	transferAmount, err := strconv.Atoi(transferAmountStr)
	if err != nil {
		return fmt.Errorf("invalid transfer amount")
	}

	tokenJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if tokenJSON == nil {
		return fmt.Errorf("asset does not exist")
	}

	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return err
	}

	if token.Amount < transferAmount {
		return fmt.Errorf("insufficient balance")
	}

	token.Amount -= transferAmount
	newToken := Token{
		AssetID: assetID,
		Owner:   newOwner,
		Amount:  transferAmount,
	}

	updatedTokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}
	newTokenJSON, err := json.Marshal(newToken)
	if err != nil {
		return err
	}

	ctx.GetStub().PutState(assetID, updatedTokenJSON)
	ctx.GetStub().PutState(assetID+"_"+newOwner, newTokenJSON)

	return nil
}

// BurnTokens removes tokens from circulation
func (s *SmartContract) BurnTokens(ctx contractapi.TransactionContextInterface, assetID string, burnAmountStr string) error {
	burnAmount, err := strconv.Atoi(burnAmountStr)
	if err != nil {
		return fmt.Errorf("invalid burn amount")
	}

	tokenJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if tokenJSON == nil {
		return fmt.Errorf("asset does not exist")
	}

	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return err
	}

	if token.Amount < burnAmount {
		return fmt.Errorf("insufficient balance to burn")
	}

	token.Amount -= burnAmount
	if token.Amount == 0 {
		return ctx.GetStub().DelState(assetID)
	}

	updatedTokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, updatedTokenJSON)
}

// GetAssetBalance retrieves the balance of an asset
func (s *SmartContract) GetAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Token, error) {
	tokenJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}
	if tokenJSON == nil {
		return nil, fmt.Errorf("asset does not exist")
	}

	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating tokenization chaincode: %s", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting tokenization chaincode: %s", err)
	}
}
