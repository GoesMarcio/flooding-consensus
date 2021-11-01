package Flooding

import (
	"fmt"
	// "os"
	"math/rand"
	"time"

	// "log"
	"encoding/json"
	"strings"

	. "../BEB"
)

const (
	Prop    = "proposal"
	Deliver = "deliver"
	Decided = "decided"
)

type Proposal struct {
	From   string `json:"from"`
	Number int    `json:"number"`
	Round  int    `json:"round"`
}

type Flooding_Module struct {
	Correct_events []string
	Round          int
	Decision       bool
	ReceivedFrom   [][]string
	Proposals      [][]Proposal
	BEB            BestEffortBroadcast_Module
}

type JSON map[string]interface{}

func (module Flooding_Module) Init(addresses []string) {
	module.Correct_events = addresses
	n_proccess := len(module.Correct_events)
	module.Round = 1
	module.Decision = false
	module.ReceivedFrom = make([][]string, n_proccess)
	module.Proposals = make([][]Proposal, n_proccess)
	module.ReceivedFrom[0] = module.Correct_events

	module.BEB = BestEffortBroadcast_Module{
		Req: make(chan BestEffortBroadcast_Req_Message),
		Ind: make(chan BestEffortBroadcast_Ind_Message)}

	module.BEB.Init(addresses[0])
	module.Send(Prop)
	module.Receive()
}

func (module Flooding_Module) Send(typeMessage string) {
	switch typeM := typeMessage; typeM {
	case Prop:
		prop := Proposal{From: module.Correct_events[0], Number: generateRandom(), Round: module.Round}
		module.Proposal(prop)
	default:
		fmt.Printf("Outra opcao")
	}
}

func (module Flooding_Module) Receive() {
	print("ovo recebe")
	go func() {
		for {
			in := <-module.BEB.Ind
			message := strings.Split(in.Message, "§")
			in.From = message[1]
			in.Message = message[0]

			var data JSON
			json.Unmarshal([]byte(in.Message), &data)

			switch typeM := data["type"]; typeM {
			case Prop:
				// prop := Proposal{From: data[""]}
				// prop := Proposal{From: module.Correct_events[0], Number: generateRandom(), Round: module.Round}
				// module.Proposal(prop)
				fmt.Printf("Prop")

			default:
				fmt.Printf("Outra opcao")
			}

			// var anyJson map[string]interface{}
			// json.Unmarshal([]byte, &anyJson)
			// fmt.Printf("Message from %v: %v\n", in.From, in.Message)

			// receivedFrom := module.ReceivedFrom[proposal.Round]
			// receivedFrom = append(receivedFrom, proposal.From)
			// module.ReceivedFrom[proposal.Round] = receivedFrom

			// proposals := module.Proposals[proposal.Round]
			// proposals = append(proposals, proposal)
			// module.Proposa	ls[proposal.Round] = proposals

			// module.CheckAndDecide()
		}
	}()
}

func (module Flooding_Module) Proposal(v Proposal) {
	print("ovo envia")
	go func() {
		module.Proposals[module.Round] = Union(module.Proposals[module.Round], []Proposal{v})

		data := make(JSON)
		data["data"] = module.Proposals[module.Round]
		data["type"] = Prop

		messageJson, _ := json.Marshal(data)
		encoded := string(messageJson)

		req := BestEffortBroadcast_Req_Message{
			Addresses: module.Correct_events[1:],
			Message:   encoded + "§" + module.Correct_events[0]}
		module.BEB.Req <- req
	}()

}

func (module Flooding_Module) CheckAndDecide() {
	if isSubSet(module.Correct_events, module.ReceivedFrom[module.Round]) && !module.Decision {
		fmt.Print()
	} else {
		fmt.Print()
	}
}

func (module Flooding_Module) Crash() {
	//aqui enviar para os outros processos
}

func Union(a, b []Proposal) []Proposal {
	m := make(map[Proposal]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

func isSubSet(conj1 []string, conj2 []string) bool {
	return true
}

func generateRandom() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	return r1.Intn(100)
}

func encodeJson(proposals []Proposal, typeM string) string {
	messageJson, _ := json.Marshal(proposals)

	return string(messageJson)
}

func decodeJson(proposals string) []Proposal {
	res := []Proposal{}
	json.Unmarshal([]byte(proposals), &res)

	return res
}
