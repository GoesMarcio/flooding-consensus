//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
package Flooding

import (
	"fmt"
	// "os"
	"math/rand"
	"reflect"
	"sort"
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
}

type Flooding_Module struct {
	Correct_events []string
	Round          int
	Decision       int
	ReceivedFrom   [][]string
	Proposals      [][]Proposal
	BEB            BestEffortBroadcast_Module
}

type JSON map[string]interface{}

func (module Flooding_Module) Init(addresses []string) {
	module.Correct_events = addresses
	n_proccess := len(module.Correct_events)
	module.Round = 1
	module.Decision = -1
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
		prop := Proposal{From: module.Correct_events[0], Number: generateRandom()}
		module.Proposal(prop)
	case Decided:
		fmt.Println("Envia decided")
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

			round := int(data["round"].(float64))

			switch typeM := data["type"]; typeM {
			case Prop:

				module.ReceivedFrom[round] = append(module.ReceivedFrom[round], in.From)

				value := reflect.ValueOf(data["data"])
				// var proposals_received = make([]Proposal, value.Len())

				for i := 0; i < value.Len(); i++ {
					mapped := value.Index((i)).Interface().(map[string]interface{})
					from := mapped["from"].(string)
					number := int(mapped["number"].(float64))

					prop_received := Proposal{From: from, Number: number}

					module.Proposals[round-1] = append(module.Proposals[round-1], prop_received)
				}

				// fmt.Println(module.Proposals[round-1])
				// fmt.Println(module.ReceivedFrom[round])
				// fmt.Println(module.Proposals[round])

				module.CheckAndDecide()

			case Deliver:
				fmt.Println()

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
		module.Proposals[0] = Union(module.Proposals[0], []Proposal{v})

		data := make(JSON)
		data["data"] = module.Proposals[0]
		data["type"] = Prop
		data["round"] = module.Round

		messageJson, _ := json.Marshal(data)
		encoded := string(messageJson)

		req := BestEffortBroadcast_Req_Message{
			Addresses: module.Correct_events[1:],
			Message:   encoded + "§" + module.Correct_events[0]}
		module.BEB.Req <- req
	}()

}

func (module Flooding_Module) CheckAndDecide() {
	fmt.Println(module.Correct_events)
	fmt.Println(module.ReceivedFrom[module.Round])
	if IsSubSet(module.Correct_events, module.ReceivedFrom[module.Round]) && module.Decision == -1 {
		if IsEqualSet(module.ReceivedFrom[module.Round], module.ReceivedFrom[module.Round-1]) {
			module.Decision = Min(module.Proposals[module.Round])

			data := make(JSON)
			data["data"] = module.Decision
			data["type"] = Decided
			data["round"] = module.Round

			messageJson, _ := json.Marshal(data)
			encoded := string(messageJson)

			req := BestEffortBroadcast_Req_Message{
				Addresses: module.Correct_events[1:],
				Message:   encoded + "§" + module.Correct_events[0]}
			module.BEB.Req <- req

			fmt.Println("Decisao: " + string(module.Decision))

		} else {
			module.Round = module.Round + 1

			data := make(JSON)
			data["data"] = module.Proposals
			data["type"] = Prop
			data["round"] = module.Round

		}

	}
}

func (module Flooding_Module) Crash() {
	//aqui enviar para os outros processos
}

func Min(a []Proposal) int {
	if len(a) == 0 {
		return -1
	}
	min := a[0].Number

	for i := 1; i < len(a); i++ {
		if a[i].Number < min {
			min = a[i].Number
		}
	}

	return min
}

func Union(a, b []Proposal) []Proposal {
	for _, item := range b {
		a = append(a, item)
	}
	return a
}

func IsEqualSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	a_copy := make([]string, len(a))
	b_copy := make([]string, len(b))

	copy(a_copy, a)
	copy(b_copy, b)

	sort.Strings(a_copy)
	sort.Strings(b_copy)

	return reflect.DeepEqual(a_copy, b_copy)
}

func IsSubSet(a, b []string) bool {

	for _, itema := range a {
		hasItem := false

		for _, itemb := range b {
			if itema == itemb {
				hasItem = true
			}
		}

		if !hasItem {
			return false
		}
	}

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
