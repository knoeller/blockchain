/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"strconv"
	"encoding/json"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {
}
var assetIndexStr = "_assetindex"
type Asset struct {
	Name string `json:"name"`
	Value string `json:"value"`
	User string `json:"user"`
}

func main() {
	shim.Start(new(SimpleChaincode))
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))
	if err != nil {
		return nil, err
	}
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)
	err = stub.PutState(assetIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return t.Invoke(stub, function, args)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "delete" {
		res, err := t.Delete(stub, args)
		return res, err
	} else if function == "write" {
		return t.Write(stub, args)
	} else if function == "create" {
		return t.create(stub, args)
	} else if function == "transfer" {
		res, err := t.transfer(stub, args)
		return res, err
	}
	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "read" {
		return t.read(stub, args)
	}
	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}
	name = args[0]
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}
	return valAsbytes, nil
}

func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	name := args[0]
	err := stub.DelState(name)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}
	assetAsBytes, err := stub.GetState(assetIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get asset index")
	}
	var assetIndex []string
	json.Unmarshal(assetAsBytes, &assetIndex)
	for i,val := range assetIndex{
		if val == name{
			assetIndex = append(assetIndex[:i], assetIndex[i+1:]...)
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(assetIndex)
	err = stub.PutState(assetIndexStr, jsonAsBytes)
	return nil, nil
}

func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string
	var err error
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}
	name = args[0]
	value = args[1]
	err = stub.PutState(name, []byte(value))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *SimpleChaincode) create(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	//   0       1       2  
	// "key", "value", "user", 
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	name := args[0]
	value := strings.ToLower(args[1])
	user := strings.ToLower(args[2])
	asset1AsBytes, err := stub.GetState(name)
	if err != nil {
		return nil, errors.New("Failed to get asset name")
	}
	res := Asset{}
	json.Unmarshal(asset1AsBytes, &res)
	if res.Name == name{
		return nil, errors.New("This asset arleady exists")
	}
	str := `{"name": "` + name + `", "value": "` + value + `", "user": "` + user + `"}`
	err = stub.PutState(name, []byte(str))
	if err != nil {
		return nil, err
	}
	asset2AsBytes, err := stub.GetState(assetIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get asset index")
	}
	var assetIndex []string
	json.Unmarshal(asset2AsBytes, &assetIndex)
	assetIndex = append(assetIndex, name)
	jsonAsBytes, _ := json.Marshal(assetIndex)
	err = stub.PutState(assetIndexStr, jsonAsBytes)
	return nil, nil
}

func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	//   0       1
	// "name", "user"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	assetAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Asset{}
	json.Unmarshal(assetAsBytes, &res)
	res.User = args[1]
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
