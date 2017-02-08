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
var blockhausIndexStr = "_blockhausindex"
type Blockhaus struct{
	Name string `json:"name"`
	Color string `json:"color"`
	Size int `json:"size"`
	User string `json:"user"`
}
type Description struct{
	Color string `json:"color"`
	Size int `json:"size"`
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
	err = stub.PutState(blockhausIndexStr, jsonAsBytes)
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
	} else if function == "init_blockhaus" {
		return t.init_blockhaus(stub, args)
	} else if function == "set_user" {
		res, err := t.set_user(stub, args)
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
	blockhausAsBytes, err := stub.GetState(blockhausIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get blockhaus index")
	}
	var blockhausIndex []string
	json.Unmarshal(blockhausAsBytes, &blockhausIndex)
	for i,val := range blockhausIndex{
		if val == name{
			blockhausIndex = append(blockhausIndex[:i], blockhausIndex[i+1:]...)
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(blockhausIndex)
	err = stub.PutState(blockhausIndexStr, jsonAsBytes)
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

func (t *SimpleChaincode) init_blockhaus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	//   0       1       2     3
	// "key", "arg1", "arg2", "user"
	if len(args) != 4 {
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
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}
	name := args[0]
	color := strings.ToLower(args[1])
	user := strings.ToLower(args[3])
	size, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}
	blockhaus1AsBytes, err := stub.GetState(name)
	if err != nil {
		return nil, errors.New("Failed to get blockhaus name")
	}
	res := Blockhaus{}
	json.Unmarshal(blockhaus1AsBytes, &res)
	if res.Name == name{
		return nil, errors.New("This blockhaus arleady exists")
	}
	str := `{"name": "` + name + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "user": "` + user + `"}`
	err = stub.PutState(name, []byte(str))
	if err != nil {
		return nil, err
	}
	blockhaus2AsBytes, err := stub.GetState(blockhausIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get blockhaus index")
	}
	var blockhausIndex []string
	json.Unmarshal(blockhaus2AsBytes, &blockhausIndex)
	blockhausIndex = append(blockhausIndex, name)
	jsonAsBytes, _ := json.Marshal(blockhausIndex)
	err = stub.PutState(blockhausIndexStr, jsonAsBytes)
	return nil, nil
}

func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	blockhausAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Blockhaus{}
	json.Unmarshal(blockhausAsBytes, &res)
	res.User = args[1]
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
