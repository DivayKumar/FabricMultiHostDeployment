package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type ProductTransferSmartContract struct {
	contractapi.Contract
}

type Product struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Area      string `json:"area"`
	OwnerName string `json:"ownerName"`
	Value     string `json:"cost"`
}
type Signup struct {
	UserName string `json:"uname"`
	Email    string `json:"email"`
	AArea    string `json:"aarea"`
	Budget   string `json:"budget"`
	Password string `json:"password"`
}

type HistoryQueryResult struct {
	Record    *Product  `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
}

//This function will signup the user
func (pc *ProductTransferSmartContract) SignUp(ctx contractapi.TransactionContextInterface, uname string, email string, aarea string, budget string, password string) error {

	productJSON, err := ctx.GetStub().GetState(email)
	if err != nil {
		return fmt.Errorf("Failed to read the data from world state", err)
	}

	if productJSON != nil {
		return fmt.Errorf("the product %s already exists", email)
	}
	p := Signup{
		UserName: uname,
		Email:    email,
		AArea:    aarea,
		Budget:   budget,
		Password: password,
	}

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(email, data)
}

// This function helps to Add new product
func (pc *ProductTransferSmartContract) AddProduct(ctx contractapi.TransactionContextInterface, id string, name string, area string, ownerName string, cost string) error {

	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("Failed to read the data from world state", err)
	}

	if productJSON != nil {
		return fmt.Errorf("the product %s already exists", id)
	}

	prop := Product{
		ID:        id,
		Name:      name,
		Area:      area,
		OwnerName: ownerName,
		Value:     cost,
	}

	productBytes, err := json.Marshal(prop)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, productBytes)
}

// This function returns all the existing product
func (pc *ProductTransferSmartContract) QueryAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	productIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer productIterator.Close()

	var products []*Product
	for productIterator.HasNext() {
		productResponse, err := productIterator.Next()
		if productResponse != nil {

			if err != nil {
				return nil, err
			}

			var product *Product
			err = json.Unmarshal(productResponse.Value, &product)
			if err != nil {
				return nil, err
			}
			// fmt.Printf("- %d\n", product)

			products = append(products, product)
		}
	}

	return products, nil
}

// This function helps to query the product by Id
func (pc *ProductTransferSmartContract) QueryProductById(ctx contractapi.TransactionContextInterface, id string) (*Product, error) {
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the data from world state", err)
	}

	if productJSON == nil {
		return nil, fmt.Errorf("the product %s does not exist", id)
	}

	var product *Product
	err = json.Unmarshal(productJSON, &product)

	if err != nil {
		return nil, err
	}
	return product, nil
}

//login
func (pc *ProductTransferSmartContract) Login(ctx contractapi.TransactionContextInterface, email string) (*Signup, error) {
	dataJSON, err := ctx.GetStub().GetState(email)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the data from world state", err)
	}

	if dataJSON == nil {
		return nil, fmt.Errorf("the user %s does not exist", email)
	}

	var product *Signup
	err = json.Unmarshal(dataJSON, &product)

	if err != nil {
		return nil, err
	}
	return product, nil
}

//update budget
func (pc *ProductTransferSmartContract) UpdateBudget(ctx contractapi.TransactionContextInterface, gmail string, newBudget string) error {
	curr, err := pc.Login(ctx, gmail)
	if err != nil {
		return err
	}

	curr.Budget = newBudget

	productJSON, err := json.Marshal(curr)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(gmail, productJSON)

}

// This functions helps to transfer the location of the product
//change owner
//money exchange
func (pc *ProductTransferSmartContract) TransferProduct(ctx contractapi.TransactionContextInterface, from string, newOwner string, newArea string) error {
	curr, err := pc.QueryProductById(ctx, from)
	if err != nil {
		return err
	}

	curr.OwnerName = newOwner
	curr.Area = newArea

	productJSON, err := json.Marshal(curr)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(from, productJSON)

}

// GetProductHistory returns the chain of custody for an asset since issuance.
func (pc *ProductTransferSmartContract) GetProductHistory(ctx contractapi.TransactionContextInterface, id string) ([]HistoryQueryResult, error) {
	log.Printf("GetProductHistory: ID %v", id)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Product
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Product{
				ID: id,
			}
		}
		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}
		record := HistoryQueryResult{
			TxId:      response.TxId,
			Record:    &asset,
			Timestamp: timestamp,
		}
		records = append(records, record)
	}

	return records, nil
}

func main() {
	proTransferSmartContract := new(ProductTransferSmartContract)

	cc, err := contractapi.NewChaincode(proTransferSmartContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}

}