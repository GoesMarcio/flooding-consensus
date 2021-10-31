package Flooding

import(
	"fmt"
	// "os"
	"math/rand"
	"time"
	// "log"
	"encoding/json"
	"strings"
	. "../BEB"
)

type Proposal struct {
	From    string `json:"from"`
	Number  int `json:"number"`
}

type Flooding_Module struct {
	Correct_events [] string
	Round int
	Decision bool
	ReceivedFrom [][]string
	Proposals []Proposal
	BEB BestEffortBroadcast_Module
}

func (module Flooding_Module) Init(addresses []string) {
	module.Correct_events = addresses
	n_proccess := len(module.Correct_events)
	module.Round = 1
	module.Decision = false
	module.ReceivedFrom = make([][]string, n_proccess)
	// module.Proposals = make([]Proposal, n_proccess)
	module.ReceivedFrom[0] = module.Correct_events

	module.BEB = BestEffortBroadcast_Module{
		Req: make(chan BestEffortBroadcast_Req_Message),
		Ind: make(chan BestEffortBroadcast_Ind_Message)}

	module.BEB.Init(addresses[0])

	module.Send_Proposal(Proposal{From: addresses[0], Number: generateRandom()})
	module.Receive_Proposal()
}

func (module Flooding_Module) Send_Proposal(v Proposal){
	print("ovo envia")
	go func(){
		module.Proposals = append(module.Proposals, v)
		
		req := BestEffortBroadcast_Req_Message{
			Addresses: module.Correct_events[1:],
			Message:   encodeJson(module.Proposals) + "§" + module.Correct_events[0]}
		module.BEB.Req <- req
	}()
	
}

func (module Flooding_Module) Receive_Proposal(){
	print("ovo recebe")
	go func(){
		for {
			in := <-module.BEB.Ind
			message := strings.Split(in.Message, "§")
			in.From = message[1]
			// registro = append(registro, in.Message)
			in.Message = message[0]

			fmt.Printf("Message from %v: %v\n", in.From, in.Message)
			fmt.Println(decodeJson(in.Message))
		}
	}()
}



func generateRandom() int{
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	return r1.Intn(100)
}


func encodeJson(proposals []Proposal) string{
	messageJson, _ := json.Marshal(proposals)

	return string(messageJson)
}

func decodeJson(proposals string) []Proposal{
	res := []Proposal{}
	json.Unmarshal([]byte(proposals), &res)

	return res
}