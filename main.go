package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _MainLogger = shim.NewLogger("CarTrackLogger")

//SmartContract for car tracking
type SmartContract struct {
}

//CarDetails represents the car record to be stored in ledger
type CarDetails struct {
	ObjType       string `json:"objType"`
	ChasisNumber  string `json:"chasisNumber"`
	Manufacturer  string `json:"manufacturer"`
	Year          string `json:"makeYear"`
	Model         string `json:"model"`
	Color         string `json:"color"`
	LisenseNunber string `json:"licNumber"`
	Status        string `json:"status"`
	Dealer        string `json:"dealer"`
	OwnerName     string `json:"owner"`
	UpdateTs      string `json:"ts"`
	TrxnID        string `json:"trxnId"`
	UpdateBy      string `json:"updBy"`
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_MainLogger.Infof("Inside the init method ")

	return shim.Success(nil)
}
func (sc *SmartContract) probe(stub shim.ChaincodeStubInterface) pb.Response {
	ts := ""
	_MainLogger.Info("Inside probe method")
	tst, err := stub.GetTxTimestamp()
	if err == nil {
		ts = tst.String()
	}
	output := "{\"status\":\"Success\",\"ts\" : \"" + ts + "\" }"
	_MainLogger.Info("Retuning " + output)
	return shim.Success([]byte(output))
}

func (sc *SmartContract) createCarEntry(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_MainLogger.Errorf("Invalid number of arguments")
		return shim.Error("Invalid number of arguments")
	}
	var carDetails CarDetails
	if err := json.Unmarshal([]byte(args[0]), &carDetails); err != nil {
		_MainLogger.Errorf("Unable to parse the input car details JSON %v", err)
		return shim.Error("Unable to parse the input car details JSON")
	}
	idOk, manuf := sc.getInvokerIdentity(stub)
	if !idOk {
		return shim.Error("Unable to retrive the invoker ID")
	}
	if strings.TrimSpace(carDetails.ChasisNumber) == "" {
		_MainLogger.Error("No chasis number provided")
		return shim.Error("No chasis number provided")
	}
	carDetails.Manufacturer = manuf
	carDetails.ObjType = "car"
	carDetails.UpdateBy = manuf
	carDetails.Status = "NEW"
	carDetails.TrxnID = stub.GetTxID()
	carDetails.UpdateTs = sc.getTrxnTS(stub)
	jsonBytesToStore, _ := json.Marshal(carDetails)
	//TODO: Check the chasis number
	if err := stub.PutState(carDetails.ChasisNumber, jsonBytesToStore); err != nil {
		_MainLogger.Errorf("Unable to store the car details %v", err)
		return shim.Error("Unable to store the car details ")
	}

	return shim.Success([]byte(jsonBytesToStore))
}
func (sc *SmartContract) saveKV(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments")
	}
	inputJSON := args[0]
	kvList := make([]map[string]string, 0)
	err := json.Unmarshal([]byte(inputJSON), &kvList)
	if err != nil {
		return shim.Error("Can not convert input JSON to valid input")
	}
	if len(kvList) == 0 {
		return shim.Error("Empty data provided")
	}
	for _, kv := range kvList {
		key := kv["key"]
		value := kv["value"]
		txID := stub.GetTxID()
		dataToStore := map[string]string{
			"value":  value,
			"trxnId": txID,
			"id":     key,
		}
		jsonBytesToStore, _ := json.Marshal(dataToStore)
		stub.PutState(key, jsonBytesToStore)
	}

	return shim.Success([]byte(fmt.Sprintf("%d records saved", len(kvList))))
}
func (sc *SmartContract) query(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments")
	}
	key := args[0]
	data, err := stub.GetState(key)
	if err != nil {
		return shim.Success(nil)

	}

	return shim.Success(data)
}

//Invoke is the entry point for any transaction
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, _ := stub.GetFunctionAndParameters()
	switch action {
	case "probe":
		response = sc.probe(stub)
	case "createCarEntry":
		response = sc.createCarEntry(stub)
	case "saveKV":
		response = sc.saveKV(stub)
	case "query":
		response = sc.query(stub)
	default:
		response = shim.Error("Invalid action provoided")
	}
	return response
}

func (sc *SmartContract) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
	//Following id comes in the format X509::<Subject>::<Issuer>>
	enCert, err := id.GetX509Certificate(stub)
	if err != nil {
		return false, "Unknown."
	}
	return true, fmt.Sprintf("%s", enCert.Subject.CommonName)

}
func (sc *SmartContract) getTrxnTS(stub shim.ChaincodeStubInterface) string {
	txTime, err := stub.GetTxTimestamp()
	if err != nil {
		return "0000.00.00.00.00.000"
	}
	var ts time.Time
	newTS := ts.Add(time.Duration(txTime.Seconds) * time.Second)
	return newTS.Format("2006.01.02.15.04.05.000")

}
func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		_MainLogger.Criticalf("Error starting  chaincode: %v", err)
	}
}
