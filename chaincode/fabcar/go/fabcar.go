package main


import (
	"bytes"
	"math"
	//"os/exec"

	//"syscall"

	//"debug/elf"
	"time"
	//"os/exec"

	//"container/list"
	"fmt"
	"encoding/json"
	//"math"

	//"container/list"
	"github.com/hyperledger/fabric/core/chaincode/shim"

	sc "github.com/hyperledger/fabric/protos/peer"


	//"time"
	"strconv"
	"strings"
	//"unicode"

)


var TRUSTMIN float64=2.0
var REGULATORPERIOD int64=60
var DR  float64 = 0.0


// Define the Smart Contract structure
type SmartContract struct {
}

type Commodity struct{
	//ID string `json:"ID"`
	Name string `json:"Name"`
	ProducerID  string `json:"Producer"`//This the trader's name
	OwnerID string `json:"Owner ID"`
	//Type string `json:"type"`
	Inventory float64 `json:"Inventory"`
	Price float64 `json:"Price"`
	DeliveryTime float64 `json:"Delivery Time"`

	MaxTemp float64 `json:"MaxTemperature"`
	MinTemp float64 `json:"MinTemperature"`
	MaxDamage float64 `json:"MaxDamageTemperature"` //Damage temperature higher bound
	MinDamage float64 `json:"MinDamageTemperature"` //Damage temperature lower bound

	RepSens []float64 `json:"Sensor Ratings"`
	//RepSensT []int64 `json:"Sensor Rating Times"`

	OverallRep float64 `json:"Final Reputation"`

}

//type SellerScore_com struct{
//	ComID string `json:"commodity ID"`
//	Trust_score float64 `json:"trust score"`
//	Seller_reps []float64 `json:" reputation scores"`
//	Overall_rep float64 `json:"over reputation score"`
//}


//for now, only consider one trader has one products and its trust and reputation score.
type Trader struct{
	Name string `json:"Trader Name"`
	Status bool `json:"Trader Status"`//true or false
	CommoID string `json:"Commodity ID"`//commodity name
	TrustScore float64 `json:"Trust Score"`//initial trust value is trust_min

	ScorerID []string `json:" Scorer IDs"` //for recording scorerID

	RatingByBuyer[][5]float64 `json:" Buyer Ratings"`//for recording customer ratings,they are 5-dimensional vector ，five indicators are product price, product quality, service price, service quality and delivery time
	RatingByBuyerT []int64 `json:"Buyer Rating Times"`//for recording the rating moments.

	RatingByRegulator []float64 `json:"Regulator Ratings"`//for recording the ratings from regulators
	RatingByRegulatorT []int64 `json:"Regulator Rating Times"`//for recording the moment corresponding to regulator's ratings

	SellersRep []float64
	SellersID []string

	SuccessTradeN int `json:"Successful Trades Number"`



}



/*
 * The Init method is called when the Smart Contract "scorecalculation" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
// ChaincodeStubInterface is used by deployable chaincode apps to access and
// modify their ledgers
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "score calculation"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "createCommodity" { // a new commodity is created
		return s.createCommodity(APIstub, args)
	} else if function =="initLedger"{
		return s.initLedger(APIstub)
	} else if function == "createTrader" { //a new seller enter the fabric network
		return s.createTrader(APIstub, args)
	} else if function == "readTemperature" { // sensor reads temperature and calculate the reputation,
		// which should be recorded into database, and may also generate warning notification
		return s.calcRepuCom(APIstub, args)
	} else if function == "startTrade" { // a trade transaction, invokes 1, caculate reputation(t),2， read reputation(t)
		//and read rep_reg, rep_buyer, to calculate rep_seller
		return s.startTrade(APIstub, args)
	} else if function == "receiptCommodity" { //calculate the overall reputation of the commodity
		return s.receiptCommodity(APIstub, args)
	} else if function == "queryAllCommodities"{
		return s.queryAllCommodities(APIstub)
	}else if function == "queryCommodity"{
		return s.queryCommodity(APIstub,args)
	}else if function == "queryAllTraders"{
		return s.queryAllTraders(APIstub)
	}else if function=="queryTrader"{
		return s.queryTrader(APIstub,args)
	} else if function == "regulatorRating"{
		return s.regulatorRate(APIstub,args)
	//} else if function == "queryOtherCustomersRating"{
	//	return s.querySellerRatings(APIstub,args)
	}else if function == "computeSellerReputation"{
		return s.calcSellerRep(APIstub, args)
	}else if function == "findOptimalSeller"{
		return s.computeMatchScore(APIstub,args)
	}else if function =="computeTrustScore"{
		return s.calTrustScore(APIstub, args)
	} else{
		return shim.Error("Invalid Smart Contract function name.")
	}

}




func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response{
	//traders :=[]Trader{
	//	{Name:"MengNiu",Status: true,CommoID: "COMMODITY0",TrustScore: TRUSTMIN},//TRADER0
	//	{Name:"Alice",Status: true,CommoID: "COMMODITY1",TrustScore: TRUSTMIN},//TRADER1
	//	{Name:"Bob",Status: true,CommoID: "COMMODITY2",TrustScore: TRUSTMIN},//TRADER2
	//	{Name:"Bella",Status: true,CommoID: "COMMODITY3",TrustScore: TRUSTMIN},//TRADER3
	//	{Name:"Jack",Status: true,CommoID: "COMMODITY4",TrustScore: TRUSTMIN},//TRADER4
	//	{Name:"Leo",Status: true,TrustScore: TRUSTMIN},//TRADER5
	//	{Name:"Buyer1",Status: true,TrustScore: TRUSTMIN},//TRADER6
	//	{Name:"Buyer2",Status: true,TrustScore: TRUSTMIN},//TRADER7
	//	{Name:"Buyer3",Status: true,TrustScore: TRUSTMIN},//TRADER8
	//	{Name:"Buyer4",Status: true,TrustScore: TRUSTMIN},//TRADER9
	//	{Name:"Buyer5",Status: true,TrustScore: TRUSTMIN},//TRADER10
	//}

	traders :=[]Trader{
		//{RatingByBuyer: [][5]float64{{7.330952380952381,7,5,3,2},{6.830952380952381,8,4,4,3},{7.330952380952381,9,3,5,2},{6.830952380952381,8,5,6,3},{6.330952380952381,9,4,3,4},{5.830952380952381,10,3,4,3},{7.330952380952381,9,3,2,6},{6.830952380952381,10,5,5,5},{6.330952380952381,9,7,6,6},{7.330952380952381,8,9,8,7},{6.330952380952381,10,10,9,8},{6.830952380952381,8,9,10,4},{7.330952380952381,9,5,6,7},{6.830952380952381,9,7,7,8},{6.330952380952381,10,7,6,9},{6.330952380952381,9,5,6,7},{5.330952380952381,6,3,4,5}},ScorerID:[]string{"TRADER5","TRADER5","TRADER5","TRADER5","TRADER5","TRADER5","TRADER6","TRADER6","TRADER6","TRADER6","TRADER6","TRADER6","TRADER7","TRADER7","TRADER7","TRADER7","TRADER7"},RatingByBuyerT: []int64{1618143613,1618143657,1618143668,1618143681,1618143693,1618143705,1618143999,1618144016,1618144028,1618144043,1618144056,1618144075,1618144318,1618144330,1618144341,1618144351,1618144365},CommoID: "COMMODITY0",DecayFactor:0,RatingByRegulatorT: []int64{1618143200,1618143217,1618143220,1618143224,1618143227,1618143230,1618143233,1618143240},RatingByRegulator: []float64{6,5,4,3,7,8,9,9},SuccessTradeN: 17,Name: "MengNiu",Status: true,TrustScore: TRUSTMIN},
		//{RatingByBuyer:[][5]float64{{6.442857142857143,8,4,3,6},{5.942857142857143,9,5,4,5},{6.442857142857143,7,4,5,4},{6.942857142857143,6,3,3,2},{4.442857142857143,8,4,2,1},{7.942857142857143,9,5,4,5},{7.442857142857143,9,9,9,8},{7.442857142857143,8,8,8,7},{6.942857142857143,6,5,6,7},{6.442857142857143,6,5,5,7},{7.442857142857143,8,5,6,8},{7.942857142857143,9,8,9,9},{7.442857142857143,10,7,6,8},{7.942857142857143,9,8,6,9}},ScorerID:[]string{"TRADER5","TRADER5","TRADER5","TRADER5","TRADER5","TRADER6","TRADER6","TRADER6","TRADER6","TRADER6","TRADER8","TRADER8","TRADER8","TRADER8"},RatingByBuyerT: []int64{1618143770,1618143783,1618143795,1618143810,1618143823,1618144206,1618144217,1618144229,1618144241,1618144252,1618144605,1618144618,1618144630,1618144643},CommoID: "COMMODITY1",RatingByRegulatorT: []int64{1618143286,1618143289,1618143293,1618143295,1618143298,1618143301,1618143305},RatingByRegulator: []float64{2,3,5,6,8,10,9},SuccessTradeN: 14,Name: "Alice",Status: true,TrustScore: TRUSTMIN},
		//{RatingByBuyer:[][5]float64{{5.916666666666666,2,4,3,5},{6.416666666666666,6,5,5,7},{6.916666666666666,5,5,4,6},{6.916666666666666,5,3,4,6},{6.416666666666666,8,6,5,5},{7.416666666666666,9,6,8,6},{5.916666666666666,6,2,3,4}},ScorerID: []string{"TRADER5","TRADER5","TRADER5","TRADER7","TRADER7","TRADER7","TRADER7"},RatingByBuyerT: []int64{1618143889,1618143899,1618143911,1618144418,1618144428,1618144439,1618144450},CommoID: "COMMODITY2",DecayFactor: 0,RatingByRegulatorT: []int64{1618143353,1618143357,1618143360,1618143365,1618143368,1618143371,1618143374},RatingByRegulator: []float64{7,9,10,7,4,5,4},SuccessTradeN: 7,Name: "Bob",Status:true,TrustScore: TRUSTMIN},
		//{RatingByBuyer:[][5]float64{{3.6799999999999997,2,5,4,6},{3.1799999999999997,4,4,6,5},{4.68,6,7,7,6},{6.18,9,9,8,5},{4.18,2,5,6,5},{4.68,3,4,5,4},{5.18,6,7,6,5},{4.18,5,3,6,3}},ScorerID: []string{"TRADER7","TRADER7","TRADER7","TRADER7","TRADER8","TRADER8","TRADER8","TRADER8"},RatingByBuyerT: []int64{1618144502,1618144512,1618144523,1618144534,1618144690,1618144701,1618144714,1618144725},CommoID: "COMMODITY3",RatingByRegulatorT: []int64{1618143416,1618143419,1618143422,1618143426,1618143428,1618143432,1618143435},RatingByRegulator: []float64{8,9,10,9,10,8,10},SuccessTradeN: 8,Name: "Bella",Status: true,TrustScore: TRUSTMIN},
		//{CommoID: "COMMODITY4",RatingByRegulatorT: []int64{1618143465,1618143467,1618143471,1618143475,1618143478,1618143482,1618143484,1618143489},RatingByRegulator: []float64{2,4,6,10,3,9,2,8},SuccessTradeN: 0,Name: "Jack",Status: true,TrustScore: TRUSTMIN},
		//{CommoID:"COMMODITY2",SuccessTradeN: 14,Name: "Buyer0",Status: true,TrustScore: TRUSTMIN},
		//{CommoID:"COMMODITY1",SuccessTradeN:11,Name:"Buyer1",Status:true,TrustScore:TRUSTMIN},
		//{CommoID:"COMMODITY3",SuccessTradeN:13,Name:"Buyer2",Status:true,TrustScore:TRUSTMIN},
		//{CommoID:"COMMODITY3",SuccessTradeN:8,Name:"Buyer3",Status:true,TrustScore:TRUSTMIN},
		//{Name:"Buyer4",Status:true,TrustScore:TRUSTMIN},
		//{Name:"Buyer5",Status:true,TrustScore:TRUSTMIN},
		{Name:"R0", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY0"},//TRADER0
		{Name:"R1", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY1"},//TRADER1
		{Name:"R2", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY2"},//TRADER2
		{Name:"R3", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY3"},//TRADER3
		{Name:"R4", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY4"},//TRADER4
		{Name:"R5", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY5"},//TRADER5
		{Name:"R6", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY6"},//TRADER6
		{Name:"R7", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY7"},//TRADER7
		{Name:"R8", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY8"},//TRADER8
		{Name:"R9", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY9"},//TRADER9
		{Name:"R10", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY10"},//TRADER10
		{Name:"R11", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY11"},//TRADER11
		{Name:"R12", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY12"},//TRADER12
		{Name:"R13", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY13"},//TRADER13
		{Name:"R14", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY14"},//TRADER14
		{Name:"R15", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY15"},//TRADER15
		{Name:"R16", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY16"},//TRADER16
		{Name:"R17", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY17"},//TRADER17
		{Name:"R18", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY18"},//TRADER18
		{Name:"R19", Status: true,TrustScore: TRUSTMIN, CommoID:"COMMODITY19"},//TRADER19
		{Name:"B0", Status: true,TrustScore: TRUSTMIN},//TRADER20
		{Name:"B1", Status: true,TrustScore: TRUSTMIN},//TRADER21
		{Name:"B2", Status: true,TrustScore: TRUSTMIN},//TRADER22
		{Name:"B3", Status: true,TrustScore: TRUSTMIN},//TRADER23
		{Name:"B4", Status: true,TrustScore: TRUSTMIN},//TRADER24
		{Name:"B5", Status: true,TrustScore: TRUSTMIN},//TRADER25
		{Name:"B6", Status: true,TrustScore: TRUSTMIN},//TRADER26
		{Name:"B7", Status: true,TrustScore: TRUSTMIN},//TRADER27
		{Name:"B8", Status: true,TrustScore: TRUSTMIN},//TRADER28
		{Name:"B9", Status: true,TrustScore: TRUSTMIN},//TRADER29
	}


	//commodities :=[]Commodity{
	//	//COMMODITY0
	//	{Name:"Milk",ProducerID: "TRADER0",OwnerID: "TRADER0",Inventory:400.0,Price:10.0,DeliveryTime:2.0, MaxTemp:25.0,MinTemp: 0.0,MaxDamage: 37.0, MinDamage: -10.0},
	//	//COMMODITY1
	//	{Name:"Milk",ProducerID: "TRADER1",OwnerID: "TRADER1",Inventory:650.0,Price:9.0,DeliveryTime:4.0, MaxTemp:25.0,MinTemp: 0.0,MaxDamage: 30.0, MinDamage: -5.0},
	//	//COMMODITY2
	//	{Name:"Milk",ProducerID: "TRADER2",OwnerID: "TRADER2",Inventory:1000.0,Price:15.0,DeliveryTime:1.0, MaxTemp:25.0,MinTemp: 5.0,MaxDamage: 37.0, MinDamage: 0.0},
	//	//COMMODITY3
	//	{Name:"Milk",ProducerID: "TRADER3",OwnerID: "TRADER3",Inventory:223.0,Price:3.0,DeliveryTime:3.0, MaxTemp:15.0,MinTemp: 5.0,MaxDamage: 25.0, MinDamage: 0.0},
	//	//COMMODITY4
	//	{Name:"Milk",ProducerID: "TRADER4",OwnerID: "TRADER4",Inventory:50.0,Price:1.0,DeliveryTime:1.0, MaxTemp:15.0,MinTemp: 0.0,MaxDamage: 25.0, MinDamage: -15.0},
	//}

	commodities:=[]Commodity{
		//COMMODITY0
		{DeliveryTime:1,OverallRep:9.36923076923077,Inventory:15,MaxDamage:37,MaxTemp:25,MinDamage:-10,MinTemp:0,Name:"Milk",OwnerID:"TRADER0",Price:30,ProducerID:"TRADER0",RepSens:[]float64{10,10,10,10,10,10,10,10,10,10,10,10,1.8} },
		//COMMODITY1
		{DeliveryTime:1,OverallRep:10,Inventory:27,MaxDamage:30,MaxTemp:25,MinDamage:-5,MinTemp:0,Name:"Milk",OwnerID:"TRADER1",Price:27,ProducerID:"TRADER1", RepSens:[]float64{10,10,10,10,10,10,10,10,10,10,10,10,10,10}},
		//COMMODITY2
		{DeliveryTime:2,OverallRep:6.453333333333333,Inventory:25,MaxDamage:37,MaxTemp:25,MinDamage:0,MinTemp:5,Name:"Milk",OwnerID:"TRADER2",Price:24,ProducerID:"TRADER2", RepSens: []float64{10,10,10,10,10,1.6,10,10,1.2,10,10,1.2,1.6,0.8,0.4}},
		//COMMODITY3
		{DeliveryTime:2,OverallRep:6.828571428571429,Inventory:25,MaxDamage:25,MaxTemp:15,MinDamage:0,MinTemp:5,Name:"Milk",OwnerID:"TRADER3",Price:22,ProducerID:"TRADER3", RepSens: []float64{10,10,1.6,1.2,10,10,10,1.6,10,10,0.4,10,10,0.8}},
		//COMMODITY4
		{DeliveryTime:2,OverallRep:7.922222222222222,Inventory:25,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER4",Price:20,ProducerID:"TRADER4", RepSens: []float64 {10,10,10,10,10,10,10,1.8666666666666667,10,1.7333333333333334,1.4666666666666666,10}},
		//COMMODITY5
		{DeliveryTime:1,OverallRep:9.096296296296295,Inventory:15,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER5",Price:28,ProducerID:"TRADER5", RepSens: []float64 {10,10,10,10,10,10,10,10,1.8666666666666667}},
		//COMMODITY6
		{DeliveryTime:1,OverallRep:10,Inventory:15,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER6",Price:27,ProducerID:"TRADER6", RepSens: []float64 {10,10,10,10,10,10,10,10,10,10}},
		//COMMODITY7
		{DeliveryTime:1,OverallRep:8.738461538461538,Inventory:15,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER7",Price:25,ProducerID:"TRADER7", RepSens: []float64 {10,10,10,10,10,1.8666666666666667,1.7333333333333334,10,10,10,10,10,10}},
		//COMMODITY8
		{DeliveryTime:2,OverallRep:7.456410256410256,Inventory:20,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER8",Price:22,ProducerID:"TRADER8", RepSens: []float64 {10,10,10,1.8666666666666667,10,1.7333333333333334,10,10,1.4666666666666666,10,1.8666666666666667,10,10}},
		//COMMODITY9
		{DeliveryTime:2,OverallRep:6.7481481481481485,Inventory:20,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER9",Price:21,ProducerID:"TRADER9", RepSens: []float64{10,10,10,1.7333333333333334,10,1.8666666666666667,10,10,1.7333333333333334,10,10,10,10,10,1.7333333333333334,1.7333333333333334,1.4666666666666666,1.2}},
		//COMMoDITY10
		{DeliveryTime:3,OverallRep:7.188888888888889,Inventory:20,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER10",Price:23,ProducerID:"TRADER10",RepSens: []float64 {10,10,10,10,10,10,1.7333333333333334,1.4666666666666666,1.2,10,10,1.8666666666666667}},
		//COMMODITY11
		{DeliveryTime:3,OverallRep:5.033333333333334,Inventory:30,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER11",Price:16,ProducerID:"TRADER11", RepSens: []float64 {10,1.7333333333333334,10,1.6,10,1.3333333333333333,10,10,1.2,0.9333333333333333,1.8666666666666667,1.7333333333333334}},
		//COMMODITY12
		{DeliveryTime:3,OverallRep:6.1416666666666675,Inventory:30,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER12",Price:15,ProducerID:"TRADER12",RepSens: []float64 {10,1.7333333333333334,10,1.6,1.3333333333333333,0.8,0.6666666666666666,10,10,10,10,10,0.6666666666666666,10,10,1.4666666666666666}},
		//COMMODITY13
		{DeliveryTime:3,OverallRep:5.633333333333333,Inventory:30,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER13",Price:17,ProducerID:"TRADER13", RepSens: []float64{10,10,1.6,1.4666666666666666,1.3333333333333333,1.2,1.0666666666666667,0.9333333333333333,10,10,10,10}},
		//COMMODITY14
		{DeliveryTime:2,OverallRep:0,Inventory:25,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER14",Price:18,ProducerID:"TRADER14"},
		//COMMODITY15
		{DeliveryTime:1,OverallRep:0,Inventory:25,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER15",Price:18,ProducerID:"TRADER15"},
		//COMMODITY16
		{DeliveryTime:2,OverallRep:0,Inventory:25,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER16",Price:15,ProducerID:"TRADER16"},
		//COMMODITY17
		{DeliveryTime:3,OverallRep:0,Inventory:25,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER17",Price:16,ProducerID:"TRADER17"},
		//COMMODITY18
		{DeliveryTime:1,OverallRep:0,Inventory:30,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER18",Price:11,ProducerID:"TRADER18"},
		//COMMODITY19
		{DeliveryTime:2,OverallRep:0,Inventory:35,MaxDamage:25,MaxTemp:15,MinDamage:-15,MinTemp:0,Name:"Milk",OwnerID:"TRADER19",Price:10,ProducerID:"TRADER19"},


	}


	for i:=0;i<len(traders);i++{
		fmt.Println("Traders: i is ",i)
		traderAsBytes,_ := json.Marshal(traders[i])
		APIstub.PutState("TRADER"+strconv.Itoa(i),traderAsBytes)
		fmt.Println("Add ",traders[i])
	}
	for i:=0; i<len(commodities);i++{
		fmt.Println("Commodities: i is ",i)
		commodityAsBytes,_:=json.Marshal(commodities[i])
		APIstub.PutState("COMMODITY"+strconv.Itoa(i),commodityAsBytes)
		fmt.Println("Add ",commodities[i])
	}
	return shim.Success(nil)
}



/**
Focus on commodity
when create a commodity, the information:
inputs:args[0]= COMMODITY+ID, args[1]=commodityName,args[2]=ProducerID/OwnerID,args[3]=MaxTemp,args[4]= MinTemp, args[5]=MaxDamage,args[6]=MinDamage
args[7] = Inventory, args[8] = Price, args[9]=DeliveryTime
return success submission
 */
func (s *SmartContract) createCommodity(APIstub shim.ChaincodeStubInterface, args []string) sc.Response{
	//when create a commodity, the information: commodityID,name,max_temperature, min_temperature, damage_temperature
	// have to be uploaded
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments, Expecting 10")
	}

	//first check the OwnerID whether  in database already, if not, add OwnerID first.
	var producer Trader
	producerAsByte,err:=APIstub.GetState(args[2])
	if err !=nil{
		return shim.Error("No this trader, create trader first")
	}
	_=json.Unmarshal(producerAsByte,&producer)
	producer.CommoID = args[0]
	//update trader's new information
	producerAsByte,_=json.Marshal(producer)
	APIstub.PutState(args[3],producerAsByte)

	//second,update the commodity information
	maxTemp,err1 := strconv.ParseFloat(args[3],64)
	minTemp,err2:=strconv.ParseFloat(args[4],64)
	maxDamage,err3:=strconv.ParseFloat(args[5],64)
	minDamage,err4:=strconv.ParseFloat(args[6],64)
	inventory,err5:=strconv.ParseFloat(args[7],64)
	price,err6:=strconv.ParseFloat(args[8],64)
	delivery,err7:=strconv.ParseFloat(args[9],64)
	if err1!=nil || err2!=nil || err3!=nil|| err4!=nil || err5 !=nil || err6!=nil || err7!= nil{
		return shim.Error("Incorrect type of argument")
	}

	var commodity = Commodity{Name:args[1],ProducerID:args[2],OwnerID:args[2],MaxTemp:maxTemp,MinTemp: minTemp,
		MaxDamage:maxDamage,MinDamage:minDamage, Inventory: inventory,Price:price,DeliveryTime: delivery}

	// need to record the new commodity information into database
	commodityAsByte,_:=json.Marshal(commodity)
	APIstub.PutState(args[0],commodityAsByte)

	return shim.Success(nil)
}

/*
query specific commodity：
input: args[0] = commodityID
*/
func (s *SmartContract) queryCommodity(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	commodityAsByte, err := APIstub.GetState(args[0])
	if err != nil{
		return shim.Error("Failed to read from world state")
	}

	//var commodity Commodity
	//commodity := new(Commodity)
	//_ = json.Unmarshal(commodityAsByte, commodity)
	//if err != nil{
	//	return shim.Error("Error: json convert to struct")
	//}
	//commodity := new(Commodity)


	return shim.Success(commodityAsByte)
}

/*
query all commodities recorded in the world state database
no input
 */
 */
func (s *SmartContract) queryAllCommodities(APIstub shim.ChaincodeStubInterface) sc.Response{
	startKey := "COMMODITY0"
	endKey := "COMMODITY999"

	resultsIterator,err:=APIstub.GetStateByRange(startKey,endKey)
	if err!=nil{
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten :=false
	for resultsIterator.HasNext(){
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllCommodities:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())


}

/*
query specific trader:
input: args[0]=traderID
 */
func (s *SmartContract) queryTrader(APIstub shim.ChaincodeStubInterface, args []string) sc.Response{
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}
	traderAsByte, err := APIstub.GetState(args[0])
	if err != nil{
		return shim.Error("Failed to read from world state")
	}
	return shim.Success(traderAsByte)
}

/*
query all traders:
no input
 */
func (s *SmartContract) queryAllTraders(APIstub shim.ChaincodeStubInterface) sc.Response{
	startKey := "TRADER0"
	endKey := "TRADER999"

	resultsIterator,err:=APIstub.GetStateByRange(startKey,endKey)
	if err!=nil{
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten :=false
	for resultsIterator.HasNext(){
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllCommodities:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())


}



/*
calculate the reputation score of commodity,according to the sensor read temperature to calculate the reputation score at time t.
sensor node reads temporary temperature and calculate the reputation score of commodity currently and
may generate warning notification
input: args[0]=commodityID, args[1]=current temperature
 */
func (s *SmartContract) calcRepuCom(APIstub shim.ChaincodeStubInterface, args []string) sc.Response{
	if len(args) !=2{
		return shim.Error("Incorrect number of arguments, Expecting 2")
	}

	//get commodity temperature requirement from database
	commodityAsByte,_ :=APIstub.GetState(args[0])

	commodity := Commodity{}
	_ = json.Unmarshal(commodityAsByte,&commodity)

	currTemp,_:=strconv.ParseFloat(args[1],64)//currTemp is the current time t when sensor reads the temperature of commodity
	curRep := 0.0 //current reputation， initial value is 0.0
	var warn bool //the flag indicating whether invoke warning notification

	if (currTemp<=commodity.MaxTemp) && (currTemp>=commodity.MinTemp){
		curRep = 10.0//set 10 is the max reputation score
		warn = false
	}else if (currTemp>commodity.MaxTemp) && (currTemp<commodity.MaxDamage){
		curRep = (commodity.MaxDamage - currTemp)/(commodity.MaxDamage-commodity.MaxTemp)*2.0
		warn = true
	}else if (currTemp<commodity.MinTemp) && (currTemp>commodity.MinDamage){
		curRep = (currTemp-commodity.MinDamage)/(commodity.MinTemp-commodity.MinDamage)*2.0
		warn = true
	}else{
		curRep = 0.0
		warn = true
	}
	//record the curr_rep into Rep_sens
	commodity.RepSens = append(commodity.RepSens,curRep )

	commodityAsJson, _ := json.Marshal(commodity)
	APIstub.PutState(args[0], commodityAsJson)

	if warn == true{
		return shim.Success([]byte("Temperature warning"))//warning notification
	}

	return shim.Success(nil)
}

/*
function: do regulator rating transaction, should be updated in a regular period
input: args[0]= traderID, args[1] = regulator's rating
 */
func (s *SmartContract)regulatorRate(APIstub shim.ChaincodeStubInterface, args []string) sc.Response{
	if len(args)!= 2{
		shim.Error("Incorrect number of arguments, Expecting 2")
	}
	traderAsByte,err:=APIstub.GetState(args[0])
	if err!= nil{
		shim.Error("Cannot retrieve trader from database.")
	}
	trader := Trader{}
	_=json.Unmarshal(traderAsByte,&trader)

	regularTime,_ :=APIstub.GetTxTimestamp()
	regularTimeInt:=time.Unix(regularTime.Seconds, int64(regularTime.Nanos)).Unix()

	rate,_ :=strconv.ParseFloat(args[1],64)

	//trader.RegulatorRating = rate
	//trader.RegulatorRatingTime = regularTimeInt
	trader.RatingByRegulator = append(trader.RatingByRegulator,rate)
	trader.RatingByRegulatorT=append(trader.RatingByRegulatorT,regularTimeInt)

	traderAsByte,_=json.Marshal(trader)
	APIstub.PutState(args[0],traderAsByte)

	return shim.Success(nil)
}

/**
function for trade event, do customer's rating basing on customer's feedback on trade
function input:
args[0]=commodityID
args[1]=sellerID
args[2]=buyerID
args[3] = the feedbacks on price of the product
args[4] = the feedbacks on price of the service
args[5] = the feedbacks on quality of service
args[6] = the feedbacks on delivery time
args[7] = the feedbacks on quality of the product

output:
calculate the up-to-date quality of the trading commodity. Combing it with the five indicators to form the trade score vector.
also record the buyerID.
*/

func (s *SmartContract)startTrade(APIstub shim.ChaincodeStubInterface,  args []string) sc.Response{
	if len(args) != 8{
		return shim.Error("Incorrect number of arguments, Expecting 8")
	}

	// read history reputation of the commodity and then calculate the reputation score of the commodity at this trade moment t
	commodityAsByte,_ :=APIstub.GetState(args[0])
	commodity := Commodity{}
	_=json.Unmarshal(commodityAsByte,&commodity)

	if len(commodity.RepSens) ==0{
		return shim.Error("The commodity has not been checked by sensor yet! Unaccessible!")
	}
	sum :=0.0
	for i:=0; i<len(commodity.RepSens);i++{
		sum = sum + commodity.RepSens[i]
	}
	tmp1,_:=strconv.ParseFloat(args[7],64)
	ProductQuality := (sum/float64(len(commodity.RepSens)))/2.0 + tmp1/2.0 //this value is quality score,using average function to calculate
	//get the seller's information
	sellerAsByte,_:= APIstub.GetState(args[1])
	seller := Trader{}
	_=json.Unmarshal(sellerAsByte,&seller)

	//record the ratings given by customer
	FProductPrice,_ :=strconv.ParseFloat(args[3],64)
	FServiceQuality,_:=strconv.ParseFloat(args[4],64)
	FServicePrice,_:=strconv.ParseFloat(args[5],64)
	FDeliveryTime,_:= strconv.ParseFloat(args[6], 64)

	var rating = [5]float64{ProductQuality,FProductPrice,FServiceQuality,FServicePrice,FDeliveryTime}
	seller.RatingByBuyer = append(seller.RatingByBuyer,rating)
	sellerAsByte,_ = json.Marshal(seller)
	APIstub.PutState(args[1],sellerAsByte)

	//record the rating time
	ratingTime,_ :=APIstub.GetTxTimestamp()
	ratingTimeInt:=time.Unix(ratingTime.Seconds, int64(ratingTime.Nanos)).Unix()
	seller.RatingByBuyerT= append(seller.RatingByBuyerT, ratingTimeInt)

	//record who gives these ratings
	seller.ScorerID = append(seller.ScorerID,args[2])

	seller.SuccessTradeN += 1//plus the number of the successful trades by seller

	sellerAsByte,_ = json.Marshal(seller)
	APIstub.PutState(args[1],sellerAsByte)
	fmt.Println("Update seller ",seller)

	commodityAsByte,_ =json.Marshal(commodity)
	APIstub.PutState(args[0],commodityAsByte)
	fmt.Println("Update commodity ",commodity)


	buyerAsByte,_:=APIstub.GetState(args[2])
	var buyer Trader
	_ =json.Unmarshal(buyerAsByte,&buyer)
	buyer.CommoID = args[0]
	buyer.SuccessTradeN+=1//plus the number of the successful trades for buyer
	buyerAsByte,_ =json.Marshal(buyer)
	APIstub.PutState(args[2],buyerAsByte)
	fmt.Println("Update buyer ",buyer)
	return shim.Success(nil)
}

/**
Build a function to combine five indicators into a single number by assign the five indicators different weights
input:
ratings = array type five indicators value,
weights = array type five weights

return:
a single number float64 type
 */
func combineIndicators(ratings [5]float64, weights [5]float64) float64{
	r :=0.0
	for i, val := range ratings{
		r = r+ val*weights[i]
	}
	return r
}

/**
Build a function to use time decay function to combine an array.
input:
ratings = a slice contains ratings (five indicators has been combined before)
times = a slice contains rating times
decayFactor = a decay factor which controls the decay speed
cT = current time
output:
a single number float64 type
 */

func timeDecay(ratings []float64, times []int64, decayFactor float64, cT int64) float64{
	if len(ratings) != len(times){
		//break the function
	}

	if len(ratings)==0{
		return 0.0
	}

	var weights []float64

	//cT := time.Now().Unix()//return type is int64, indicates the current time
	//var w float64
	var eT float64 // the elapsed time

	//according to the time slice, computing the corresponding weights
	for i:=0;i<len(times);i++{
		eT = float64(cT-times[i])
		w := decayFactor/(decayFactor+(eT))
		weights = append(weights,w)
	}

	sumWeights :=0.0
	for i:=0;i<len(weights);i++{
		sumWeights = sumWeights+ weights[i]
	}

	sumR := 0.0

	for i, r := range ratings{
		sumR += r * weights[i]
	}

	return sumR/sumWeights


}


/**
The aim of this function is to calculate the seller's reputation for a customer.
to-do list:
1, separate the subjective ratings and objective ratings.
2, Using k-means clustering algorithm to filter unfair ratings
3, transform the 5-dimensional vector ratings into single number format.
4, Using time decay function to compute the overall subjective and objective rating, separately.
5, Combine the subjective and objective rating into an overall reputation of a seller.
6, record the reputation

input:
args[0] = customerID //是指想要query a trader's reputation 的那个trader
args[1] = sellerID
args[2] = the first important indicators
args[3] = the second important indicators
args[4] = the third important indicators
args[5] = the fourth important indicators
args[6] = the fifth important indicators
args[7] = "history" or "current"，to determine the speed of decay function


output:
calculate the overall reputation and record into the SellersRep of the trader
 */

func (s *SmartContract) calcSellerRep(APIstub shim.ChaincodeStubInterface,args []string) sc.Response {
	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments, expecting 8 arguments")
	}

	var buyer Trader
	buyerAsByte, err1 :=APIstub.GetState(args[0])
	if err1 != nil{
		return shim.Error("Buyer is not found!")
	}
	_=json.Unmarshal(buyerAsByte, &buyer)

	var seller Trader
	sellerAsByte, err2 := APIstub.GetState(args[1])
	if err2 != nil {
		return shim.Error("Seller is not found!")
	}
	_ = json.Unmarshal(sellerAsByte, &seller)


	var subRate [][5]float64
	var obRate [][5]float64
	var subRT []int64
	var obRT []int64
	for i := 0; i < len(seller.RatingByBuyer); i++ {
		if seller.ScorerID[i] == args[0] {
			subRate = append(subRate, seller.RatingByBuyer[i])
			subRT = append(subRT, seller.RatingByBuyerT[i])
		} else {
			obRate = append(obRate, seller.RatingByBuyer[i])
			obRT = append(obRT, seller.RatingByBuyerT[i])
		}
	}
	//******************************************************************apply k-means clustering************************************************
	//var c_obRate [][5]float64//storing filtered objective ratings
	//var c_obRT []int64
	//
	// if len(obRate) >2 {
	//	//obRate needs to be filtered through k-means clustering
	//	//apply k-means clustering
	//	s1, err3 := json.Marshal(obRate)
	//	//obRate is the collection of unfiltered objective ratings
	//	if err3 != nil {
	//		 return shim.Error("obRate cannot be transfered to a string!")
	//	}
	//	s2 := fmt.Sprintf("import filter; print (filter.filtering('%s'))", s1)
	//	//out, err := exec.Command("python3", "-c", s2).Output()
	//	cmd1:=exec.Command("/usr/local/bin/python3.9", "-c", s2)
	//	out,err:=cmd1.Output()
	//	if err != nil {
	//		//cmd1.Stderr.String()
	//		 return shim.Error(err.Error())
	//		 //when I invoked this function, this error always show "no such file or directory"
	//	}
	//	var cleaned_obRateI []int
	//	err4 := json.Unmarshal([]byte(out), &cleaned_obRateI)
	//	if err4 != nil {
	//		 return shim.Error("python output cannot be transfered!")
	//	}
	//	//according to cleaned_obRateI, to build the cleaned obRate, and cleaned obRatedT
	//	//var c_obRate [][5]float64
	//	//var c_obRT []int64
	//	for _, j := range cleaned_obRateI {
	//		 c_obRate = append(c_obRate, obRate[j])
	//		 c_obRT = append(c_obRT, obRT[j])
	//	}
	// }else{
	// 	c_obRate = obRate
	// 	c_obRT = obRT
	// }
//*************************************************************************************************************************************************
	//buyer.SubRate = subRate
	//buyer.SubRT = subRT
	//buyer.ObRate = c_obRate
	//buyer.ObRT = c_obRT
	//buyer.ObRate = obRate
	//buyer.ObRT = obRT
	//buyerAsByte,_=json.Marshal(buyer)
	//APIstub.PutState(args[0],buyerAsByte)
	//pass the obRate into k-means clustering python program.
	//and that program will return the cleaned obRate and the correspondng obRT

	//according to the important order of indicators provided by the buyer，to determine the weight of each indicator
	var weights [5]float64
	w:= 1.0
	for _,f:=range args[2:7]{
		if w>0.0625{
			w = w/2
		}
		switch f{
		case "Product Quality":
			weights[0] = w
		case "Product Price":
			weights[1] = w
		case "Service Quality":
			weights[2] = w
		case "Service Price":
			weights[3] = w
		case "Delivery time":
			weights[4] = w
		}
	}
	//buyer.Weights = weights


	//determine the decay factor
	var decayFactor float64
	if args[7] == "history"{
		decayFactor = 7
	}else{
		decayFactor = 1
	}
	//buyer.DecayFactor = decayFactor

	var subRatings []float64
	var obRatings []float64
	if len(subRate)>0 && len(obRate)>0{//when subjective and objective both obtained
		for i:=0; i<len(subRT); i++{
			subRatings = append(subRatings,combineIndicators(subRate[i], weights))
		}
		for i:=0; i<len(obRate); i++{
			obRatings = append(obRatings,combineIndicators(obRate[i],weights))
		}

	}else if len(subRate)>0 && len(obRate)==0{ // when only has subjective ratings
		for i:=0;i<len(subRT);i++{
			subRatings = append(subRatings,combineIndicators(subRate[i], weights))
		}
	}else if len(subRate)==0 && len(obRate)>0 {// when only has objective ratings

		for i:=0; i<len(obRate);i++{
			obRatings = append(obRatings,combineIndicators(obRate[i],weights))
		}
	}
	//buyer.ObRatings = obRatings
	//buyer.SubRatings = subRatings

	ct1, _ := APIstub.GetTxTimestamp()
	cT := time.Unix(ct1.Seconds, int64(ct1.Nanos)).Unix()
   //combine IR and WR
	var fr float64 //FR is the overall reputation of a seller which combined the objective and subjective ratings
	var ir, wr float64
	if len(subRT)>0 && len(obRT)>0{
		ir = timeDecay(subRatings, subRT, decayFactor, cT)
		wr = timeDecay(obRatings, obRT, decayFactor,cT)
		fr = 0.6*ir + 0.4*wr
	}else if len(subRT)>0 && len(obRT)==0{
		ir = timeDecay(subRatings, subRT, decayFactor, cT)
		wr = 0.0
		fr = ir
	}else if len(subRT)==0 && len(obRT)>0{
		wr = timeDecay(obRatings, obRT, decayFactor,cT)
		ir = 0.0
		fr = wr
	}else{
		ir = 0.0
		wr =0.0
		fr = DR
	}
	//buyer.Ir = ir
	//buyer.Wr = wr
	//buyer.Fr = fr

	var has bool
	has = false
	if len(buyer.SellersID)==0{
		goto label1
	}
	for i,id := range buyer.SellersID{
		if id == args[1]{
			buyer.SellersRep[i] = fr
			has = true
		}
	}
	label1:if has == false{
		buyer.SellersRep = append(buyer.SellersRep, fr)
		buyer.SellersID = append(buyer.SellersID, args[1])
	}


	buyerAsByte,_=json.Marshal(buyer)
	APIstub.PutState(args[0],buyerAsByte)
	return shim.Success(nil)
}


//This seller reputation function does not need to be invoked by outside entities, it is used by recalling from the function computeMatchScore
func sellerRep(seller Trader,buyerID string,indicators [5]string,decayFactor float64, cT int64) float64{

	//区分subjective和objective ratings
	var subRate [][5]float64
	var obRate [][5]float64
	var subRT []int64
	var obRT []int64
	if len(seller.RatingByBuyer)==0{
		return 0.0
	}

	for i := 0; i < len(seller.RatingByBuyer); i++ {
		if seller.ScorerID[i] == buyerID {
			subRate = append(subRate, seller.RatingByBuyer[i])
			subRT = append(subRT, seller.RatingByBuyerT[i])
		} else {
			obRate = append(obRate, seller.RatingByBuyer[i])
			obRT = append(obRT, seller.RatingByBuyerT[i])
		}
	}

	//can also copy the code which applies k-means function in here for filtering


	//determine weights
	var weights [5]float64
	w:= 1.0
	for _,f:=range indicators{
		if w>0.0625{
			w = w/2
		}
		switch f{
		case "Product Quality":
			weights[0] = w
		case "Product Price":
			weights[1] = w
		case "Service Quality":
			weights[2] = w
		case "Service Price":
			weights[3] = w
		case "Delivery time":
			weights[4] = w
		}
	}

	var subRatings []float64
	var obRatings []float64
	if len(subRate)>0 && len(obRate)>0{//when subjective and objective both obtained
		for i:=0; i<len(subRT); i++{
			subRatings = append(subRatings,combineIndicators(subRate[i], weights))
		}

		for i:=0; i<len(obRate); i++{
			obRatings = append(obRatings,combineIndicators(obRate[i],weights))
		}

	}else if len(subRate)>0 && len(obRate)==0{ // when only has subjective ratings

		for i:=0;i<len(subRT);i++{
			subRatings = append(subRatings,combineIndicators(subRate[i], weights))
		}

	}else if len(subRate)==0 && len(obRate)>0 {// when only has objective ratings
		for i:=0; i<len(obRate);i++{
			obRatings = append(obRatings,combineIndicators(obRate[i],weights))
		}
	}

	//combine IR and WR
	var fr float64 //FR is the overall reputation of a seller which combined the objective and subjective ratings
	var ir, wr float64
	if len(subRT)>0 && len(obRate)>0{
		ir = timeDecay(subRatings, subRT, decayFactor, cT)
		wr = timeDecay(obRatings, obRT, decayFactor,cT)
		fr = 0.6*ir + 0.4*wr
	}else if len(subRT)>0 && len(obRate)==0{
		ir = timeDecay(subRatings, subRT, decayFactor, cT)
		wr = 0.0
		fr = ir
	}else if len(subRT)==0 && len(obRate)>0{
		wr = timeDecay(obRatings, obRT, decayFactor,cT)
		ir = 0.0
		fr = wr
	}else{
		ir = 0.0
		wr =0.0
		fr = DR
	}
	return fr
}

/*********************************************************************************************
this function is to compute the matching score for a buyer
and return the optimal sellerID and name to the buyer

features order is Price, Inventory, delivery time, commodity reputation, regulator's rating, seller's reputation
Price, Inventory are hard requirement

indicators list :Product Quality,Product Price,Service Quality,Service Price,Delivery time
**********************************************************************************************
the input:
args[0] = BuyerID
args[1] = Product name
args[2] = satisfaction vector  [6]float64 follow features order，corresponding to seller's information
args[3] = constant weight vector [6]float64 follow features order
args[4] = the first important indicators
args[5] = the second important indicators
args[6] = the third important indicators
args[7] = the fourth important indicators
args[8] = the fifth important indicators
args[9] = "history" or "current" to determine the spped of the decay function
args[10] = satisfaction vector corresponding to the status vector, []float64
 */

func (s *SmartContract) computeMatchScore(APIstub shim.ChaincodeStubInterface, args []string) sc.Response{
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments, Expecting 11 arguments")
	}

	//get the buyer's information
	buyerID :=args[0]
	var buyer Trader
	buyerAsByte,err2:=APIstub.GetState(buyerID)
	if err2!=nil{
		return shim.Error("Buyer is not found!")
	}
	_=json.Unmarshal(buyerAsByte,&buyer)

	productName :=args[1]

	//get the satisfaction vector===>>a
	var a []float64 //storing satisfaction vector
	var tmp_v []string
	tmp_v = strings.Split(args[2],",")
	for _, ch :=range tmp_v{
		ai, _ := strconv.ParseFloat(ch,64)
		a = append(a,ai)
	}

	//get the constant weight vector ===>> c_v
	var c_v []float64 //storing constant weight vector
	tmp_v = strings.Split(args[3],",")
	for _, ch:=range tmp_v{
		c,_:=strconv.ParseFloat(ch,64)
		c_v = append(c_v,c)
	}

	var indicators =[5]string{args[4],args[5],args[6],args[7],args[8]}

	//according to args[9]，determine the decay factor
	var decayFactor float64
	if args[9] == "history"{
		decayFactor = 7
	}else{
		decayFactor = 1
	}


	var a2 []float64
	tmp_v = strings.Split(args[10],",")
	for _,ch:=range tmp_v{
		ai,_:=strconv.ParseFloat(ch,64)
		a2=append(a2,ai)
	}

	var sellersID []string// for recording the sellerIDs, corresponding to x.
	var x [][]float64 //for recording the seller's Information vectors
	//filtering out the candidates
	startKey := "COMMODITY0"
	endKey := "COMMODITY999"
	resultsIterator,err:=APIstub.GetStateByRange(startKey,endKey)
	if err!=nil{
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var tmp_c Commodity //temporary storing commodity struct

	for resultsIterator.HasNext(){
		queryResponse, err:=resultsIterator.Next()
		if err !=nil{
			return shim.Error(err.Error())
		}
		_=json.Unmarshal(queryResponse.Value,&tmp_c)
		if (tmp_c.Name == productName)&& ((tmp_c.Price<a[0]) || (tmp_c.Price == a[0])) && (tmp_c.Inventory > a[1] || tmp_c.Inventory == a[1]){
			//when the hard requirements are satisfied
			//record the ownersID
			sellersID = append(sellersID,tmp_c.OwnerID)

			//record the corresponding features into x array
			//t := []float64{tmp_c.Price,tmp_c.Inventory,tmp_c.DeliveryTime}
			t := []float64{tmp_c.Price,tmp_c.Inventory,tmp_c.DeliveryTime,tmp_c.OverallRep}
			x = append(x,t)
		}

	}
	//*******************************first the hard requirement must be satisfied. If no such seller，end the function****************************8
	if len(sellersID)==0{
		var buffer bytes.Buffer
		buffer.WriteString("Not found required sellers")
		return shim.Success(buffer.Bytes())
	}

	//if only one seller satisfied customer requirements, then return this seller, and no need to do the rest computing algorithm.
	if len(sellersID)==1{
		str2:=fmt.Sprintf("The optimal seller is %s",sellersID[0])
		var buffer bytes.Buffer
		buffer.WriteString(str2)
		return shim.Success(buffer.Bytes())
	}


	//get the current time
	ct1, _ := APIstub.GetTxTimestamp()
	cT := time.Unix(ct1.Seconds, int64(ct1.Nanos)).Unix()

	//according to the sellersID,calculate the sellers regulator's rating, and the seller's reputation
	var tmp_se Trader //storing temporary seller information
	for i,id:=range sellersID{
		tmpAsByte,err:=APIstub.GetState(id)
		if err!=nil{
			return shim.Error("Seller is not found!")
		}
		_=json.Unmarshal(tmpAsByte,&tmp_se)
		//compute overall regulator's rating by using time decay function

		regulatorRate := timeDecay(tmp_se.RatingByRegulator,tmp_se.RatingByRegulatorT,decayFactor,cT)
		x[i] = append(x[i],regulatorRate) //add regualtor Rating into feature vector


		//adding seller's reputation
		//calculate the seller's reputation
		sellerReputation := sellerRep(tmp_se,buyerID,indicators,decayFactor,cT)
		//add into feature vector
		x[i] = append(x[i],sellerReputation)
	}

	//******************************feature information vector is done***************************************
	//******************************computing feature vector*************************************************
	var features [][]float64
	for _,x_v:=range x{ //x_v is the vector,it contains seller's information value
		var f_v []float64
		for i, val:=range x_v{
			var tmp float64
			if i ==0{
				tmp = ((a[i]-val)/a[i])*10.0
			}else if i==1{
				if val -a[i]>a[i]{
					tmp = 10.0
				}else{
					tmp = ((val-a[i])/a[i])*10.0
				}
			}else if i==2{
				if a[i]-val >0{
					tmp = ((a[i]-val)/a[i])*10.0
				}else{
					tmp =1.0
				}
			}else{
				tmp = val
			}
			f_v =append(f_v,tmp)
		}
		features = append(features,f_v)
	}
	//insert overall commodity reputation

	//*********************************features vector is done*************************************************

	//******************************computing state vector **************************************
	//features is [][]float64,
	var stat [][]float64 //storing all candidates state vectors
	for _,x_v :=range features{//x_v is []float64
		var s_v []float64
		for j,xj:=range x_v{ //xj is float64, 0=<j<=4
			if xj>a2[j] || xj==a2[j]{
				s_v=append(s_v,math.Exp(xj-a2[j]))
			}else{
				s_v=append(s_v,1.0)
			}
			//s_v=append(s_v,math.Exp(math.Abs(xj-a[j])))
		}
		stat = append(stat,s_v)
	}

	//***************************state vector is done********************************************
	//***************************computing variable weight vector********************************
	var vw [][]float64 //storing candidate variable weight vectors
	for _,s_v:=range stat{//s_v is []float64
		var w []float64
		var sum = 0.0
		for j,sj :=range s_v{
			w = append(w,c_v[j]*sj)
			sum = sum + c_v[j]*sj
		}
		for i:=0;i<len(w);i++{
			w[i] = w[i]/sum
		}
		vw=append(vw,w)
	}
	//***************************variable weight vector is done********************************
	//***************************computing matching score************************************

	var m []float64 //storing candidate matching scores
	for i,v1:=range vw{
		tmp:=0.0
		for j,v2:=range v1{
			tmp = tmp+v2*features[i][j]
		}
		m = append(m,tmp)
	}


	//**********************matching score is done********************************************
	//**********************choosing optimal seller*******************************************
	//choose the sellerID, whose match score is the maximum
	var maxiScore float64
	var maxID string
	maxiScore = m[0]
	maxID = sellersID[0]
	for i,score:=range m{
		if score>maxiScore{
			maxiScore = score
			maxID = sellersID[i]
		}
	}
	var str_m string="["
	for _,val :=range m{
		t:=strconv.FormatFloat(val,'f',5,64)
		str_m = str_m+", "+t
	}
	str_m=str_m+"]"

	//str_sellersID,err3:=json.Marshal(sellersID)
	//if err3!=nil{
	//	return shim.Error("sellersID array cannot transfer to string type !")
	//}

	str_sellersID:="["
	for _,val:=range sellersID{
		str_sellersID = str_sellersID+", "+val
	}
	str_sellersID=str_sellersID+"]"



	str_x := "["
	for _,v:=range x{
		str_x = str_x+"["
		for _,val :=range v{
			t:=strconv.FormatFloat(val,'f',5,64)
			str_x=str_x+", "+t
		}
		str_x = str_x+"],\n"
	}
	str_x =str_x+"]"

	str_f:="["
	for _,v:=range features{
		str_f = str_f+"["
		for _,val:=range v{
			t:=strconv.FormatFloat(val,'f',5,64)
			str_f = str_f+t+","
		}
		str_f = str_f+"],\n"
	}
	str_f = str_f+"]"

	str_stat :="["
	for _,v:=range stat{
		str_stat = str_stat+"["
		for _,val :=range v{
			t:=strconv.FormatFloat(val,'f',5,64)
			str_stat=str_stat+", "+t
		}
		str_stat = str_stat+"],\n"
	}
	str_stat =str_stat+"]"


	str_vw :="["
	for _,v:=range vw{
		str_vw = str_vw+"["
		for _,val :=range v{
			t:=strconv.FormatFloat(val,'f',5,64)
			str_vw=str_vw+", "+t
		}
		str_vw = str_vw+"],\n"
	}
	str_vw =str_vw+"]"



	var buffer bytes.Buffer
	//if you don't need to know the intermediate results, then can delete the codes for printing those results
	result:=fmt.Sprintf("sellers information of each seller = %s \n features vectore features = %s\n status array of each sellers = %s \n variable weight of sellers = %s \n matching score = %s \n sellersID = %s \n The optimal sellerID is %s",str_x,str_f,str_stat,str_vw,str_m,str_sellersID,maxID)
	buffer.WriteString(result)
	return shim.Success(buffer.Bytes())


}






//when a new trader enter the fabric network, the information: traderID
//trader should enter fabric network first, then can allow to create commodity
/*
create a new trader information in traders database
input: args[0]=traderID, args[1]= traderName
 */
func (s *SmartContract) createTrader(APIstub shim.ChaincodeStubInterface,args []string) sc.Response{
	if len(args) != 2{
		return shim.Error("Incorrect number of arguments, Expecting 2 argument")
	}

	var trader Trader
	trader.Name = args[1]
	//trader.OverallRep = 0.0
	trader.TrustScore = TRUSTMIN
	trader.Status = true

	traderAsByte,_:=json.Marshal(trader)
	APIstub.PutState(args[0],traderAsByte)

	return shim.Success(nil)
}

/*
 The receipt transaction invokes this function, in order to calculate the overall commodity reputation
invoked by retailers
input:args[0]=commodity ID
 */
func (s *SmartContract) receiptCommodity(APIstub shim.ChaincodeStubInterface, args[] string) sc.Response{
	if len(args) !=1{
		return shim.Error("Incorrect number of arguments, Expecting 1 argument")
	}
	var commodity Commodity
	commodityAsByte, err := APIstub.GetState(args[0])
	if err != nil{
		return shim.Error("There is no the commodity")
	}
	_=json.Unmarshal(commodityAsByte,&commodity)

	//var retailer Trader
	//retailerAsByte,err1:=APIstub.GetState(commodity.OwnerID)
	//if err1 !=nil{
	//	return shim.Error("There is not this retailer! ")
	//}
	//_=json.Unmarshal(retailerAsByte,&retailer)
	//
	//if retailer.Kind !="retailer"{
	//	return shim.Error("This commodity is not at the end of supply chain")
	//}

	var sum float64
	for i:=0; i<len(commodity.RepSens); i++{
		sum = sum + commodity.RepSens[i]
	}
	commodity.OverallRep = sum / float64(len(commodity.RepSens))
	commodityAsByte,_ = json.Marshal(commodity)
	APIstub.PutState(args[0],commodityAsByte)
	return shim.Success(nil)
}

/*
Calculate the TrustScore of a trader for a commodity
inputs: args[0]=traderID
output: the trust score of a trader

This function:
1， combine the buyer's ratings
2, combine the regulator's ratings
3, combine 1 and 2 and the number of successfule trade to result the final trust score.


*/
func(s *SmartContract) calTrustScore (APIstub shim.ChaincodeStubInterface, args[] string) sc.Response{
	// traderID is args[0]
	if len(args) !=1  {
		return shim.Error("Incorrect number of arguments, Expecting 1 argument")
	}

	var trader Trader
	traderAsByte, _ := APIstub.GetState(args[0])
	_=json.Unmarshal(traderAsByte, &trader)


	ct1, _ := APIstub.GetTxTimestamp()
	cT := time.Unix(ct1.Seconds, int64(ct1.Nanos)).Unix()

	//************************************combine buyer's ratings **********************************************
	var buyerRate float64
	if len(trader.RatingByBuyer)>0{
		var r1 []float64
		for _,v:=range trader.RatingByBuyer{
			r1 = append(r1, combineIndicators(v,[5]float64{1,1,1,1,1}))
		}
		buyerRate =timeDecay(r1, trader.RatingByBuyerT,3,cT)
	}else{
		buyerRate =0.0
	}
	//**************************************done******************************************************************
	//**************************************combine regulator's ratings *******************************************
	var regulatorRate float64
	if len(trader.RatingByRegulator)>0{
		regulatorRate = timeDecay(trader.RatingByRegulator,trader.RatingByRegulatorT,3,cT)
	}else{
		regulatorRate =0.0
	}

	//**************************** the number of successful trade************************************************
	var successRate float64
	if trader.SuccessTradeN>6 || trader.SuccessTradeN==6{
		successRate =2
	}else if (trader.SuccessTradeN>4 && trader.SuccessTradeN<6) || trader.SuccessTradeN == 4{
		successRate = 1.5
	}else if (trader.SuccessTradeN>1 && trader.SuccessTradeN<3) || trader.SuccessTradeN==1 || trader.SuccessTradeN==3{
		successRate = 0.5
	}else{
		successRate = -1
	}

	//***************************combine all factors to result trust score*******************************************
	trustScore := 0.4*buyerRate+0.4*regulatorRate+0.2*successRate

	trader.TrustScore = trustScore
	traderAsByte,_=json.Marshal(trader)
	APIstub.PutState(args[0],traderAsByte)

	return shim.Success(nil)
}

	//calculate the overall reputation score of a trader at time tn
	//var overallRep float64
	////var t,currT int64
	//var t int64
	//var td float64 //td=currT-t
	//currT  = time.Now().UnixNano()
	//currT,err1 := APIstub.GetTxTimestamp()
	//if err1!=nil{
	//	shim.Error("Cannot read event time, try again")
	//}
	////currTInt := int64(*currT)
	////currTInt,_:=strconv.ParseInt(time.Unix(currT.Seconds,int64(currT.Nanos)).String(),10,64)
	//currTInt :=time.Unix(currT.Seconds, int64(currT.Nanos)).Unix()
	//for i, val :=range trader.SellerReps{
	//	t = trader.SellerRepsT[i]
	//	td = float64(t - currTInt )
	//	overallRep += val* math.Exp(-math.Abs(td))
	//}
	//trader.OverallRep = overallRep

	//calculate the overall reputation score, the second method:
	//currT := trader.SellerRepsT[len(trader.SellerRepsT)-1]
	//for i, val :=range trader.SellerReps{
	//	t = trader.SellerRepsT[i]
	//	td = float64(t-currT)/10.0
	//	overallRep += val * math.Pow(2,-math.Abs(td))
	//}
	//trader.OverallRep = overallRep
	//
	////计算trust score of a trader at time tn
	//var trust float64
	//if len(args)==1{
	//	trust = overallRep
	//	trader.TrustScore = trust
	//	//if trust <TRUSTMIN{
	//
	//	//}
	//	traderAsByte,_ = json.Marshal(trader)
	//	APIstub.PutState(args[0],traderAsByte)
	//
	//	return shim.Success(nil)
	//}

//
//	featureLevel:=strings.Split(string(args[1]),":")
//
//	var features []float64
//	features = append(features,overallRep)
//	if featureLevel[0] == "SuccessTradeNumber"{
//		switch {
//		case trader.SuccessTradeN==0: features = append(features,-1.0)
//		case trader.SuccessTradeN>=1 && trader.SuccessTradeN<=3:
//			features = append(features,0.5)
//		case trader.SuccessTradeN>=4 && trader.SuccessTradeN<6:
//			features = append(features,1.5)
//		case trader.SuccessTradeN>=6:
//			features = append(features,2)
//		}
//
//	}else{
//		shim.Error("Wrong feature name.")
//	}
//
//
//	var weights []float64
//	//featureLevel:=strings.Split(string(args[1]),":")
//	if featureLevel[1]=="high"{
//		weights = append(weights, 0.4)
//		weights = append(weights,0.6)
//	}else if featureLevel[1]=="medium"{
//		weights = append(weights, 0.5)
//		weights = append(weights, 0.5)
//	}else if featureLevel[1]=="low"{
//		weights = append(weights, 0.7)
//		weights = append(weights, 0.3)
//	}else{
//		shim.Error("Wrong feature level.")
//	}
//
//
//	//totalFeaturs := float64(len(args)-2)
//
//	//w := []float64{0.5,0.5}
//	//w = append(w, 0.5)
//
//	//var features []float64
//	//features = append(features,ovrallRep)
//
//	//for i:=2; i<len(args);i++{
//	//	if args[i] == "number of sales"{
//	//		w = append(w,0.5)
//	//		switch {
//	//			case trader.SuccessTradeN==0: features = append(features,-1.0)
//	//			case trader.SuccessTradeN>=1 && trader.SuccessTradeN<=3:
//	//				features = append(features,0.5)
//	//			case trader.SuccessTradeN>=4 && trader.SuccessTradeN<6:
//	//				features = append(features,1.5)
//	//			case trader.SuccessTradeN>=6:
//	//				features = append(features,2)
//	//		}
//	//	}
//	//}
//	//if featureLevel[0] == "SuccessTradeNumber"{
//	//	switch {
//	//				case trader.SuccessTradeN==0: features = append(features,-1.0)
//	//				case trader.SuccessTradeN>=1 && trader.SuccessTradeN<=3:
//	//					features = append(features,0.5)
//	//				case trader.SuccessTradeN>=4 && trader.SuccessTradeN<6:
//	//					features = append(features,1.5)
//	//				case trader.SuccessTradeN>=6:
//	//					features = append(features,2)
//	//			}
//	//
//	//}else{
//	//
//	//}
//
//
//	for i,val :=range features{
//		trust += val*weights[i]
//	}
//
//
//	trader.TrustScore = trust
//	//if trust <TRUSTMIN{
//	//
//	//}
//	if trust < TRUSTMIN{
//		trader.Status = false
//	}
//
//	traderAsByte,_ = json.Marshal(trader)
//	APIstub.PutState(args[0],traderAsByte)
//
//	return shim.Success(nil)
//}

/**
This function is to query the ratings from all customers to a specific seller, and then transfer this ratings to a local customer
, in order to let this local customer to separate the objective and subjective ratings and use k-means to remove unfair ratings.
The input:sellerID

The output: subjective ratings array, and objective ratings array
 */
//func(s *SmartContract) querySellerRatings (APIstub shim.ChiancodeStubInterface, args[] string) sc.Response{
//	if len(args) != 1{
//		return shim.Error("Incorrect number of arguments, expecting 2 arguments, customerID and sellerID")
//	}
//
//	//var customer Trader
//	//var seller Trader
//	//customerAsByte,err1:=APIstub.GetState(args[0])
//	//if err1 != nil{
//	//	return shim.Error("Customer is not found")
//	//}
//	//_=json.Unmarshal(customerAsByte,&customer)
//	//
//	//sellerAsByte,err2 :=APIstub.GetState(args[1])
//	//if err2 != nil {
//	//	return shim.Error("Seller is not found")
//	//}
//	//_=json.Unmarshal(sellerAsByte, &seller)
//	//
//	////build two slice to contain the subjective ratings and objective ratings, and also the corresponding rating time
//	//var subRatings [][4]int64
//	//var subRatingsT []float64
//	//var obRatings [][4]int64
//	//var obRatingsT []float64
//	//
//	//for i,rating :=range seller.RatingByCustomer{
//	//	if seller.scorer[i] == args[0]{
//	//		subRatings = append(subRatings, rating)
//	//		subRatingsT= append(subRatingsT, seller.RatingByCustomerT[i])
//	//	}else{
//	//		obRatings =append(obRatings,rating)
//	//		obRatingsT = append(obRatingsT, seller.RatingByCustomerT[i])
//	//	}
//	//}
//	//subRatingsAsByte,_ := json.Marshal()
//}

func main(){
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

}
