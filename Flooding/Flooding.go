package Flooding

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	. "../BEB"
)

const (
	Prop    = "proposal"
	Deliver = "deliver"
	Decided = "decided"
	Crash   = "crash"
)

var (
	Decision = -1
	Clock    = 0
	Round    = 1
	Proccess = []string{}
)

type Proposal struct {
	From   string `json:"from"`
	Number int    `json:"number"`
}

type Flooding_Module struct {
	Correct_events []string
	ReceivedFrom   [][]string
	Proposals      [][]Proposal
	BEB            BestEffortBroadcast_Module
	FailRound      int
}

type JSON map[string]interface{}

func (module Flooding_Module) Init(addresses []string, failRound int) {
	module.Correct_events = addresses

	n_proccess := len(module.Correct_events)

	module.ReceivedFrom = make([][]string, n_proccess)
	module.Proposals = make([][]Proposal, n_proccess)
	module.ReceivedFrom[0] = module.Correct_events

	module.BEB = BestEffortBroadcast_Module{
		Req: make(chan BestEffortBroadcast_Req_Message),
		Ind: make(chan BestEffortBroadcast_Ind_Message)}

	module.FailRound = failRound

	module.BEB.Init(addresses[0])
	time.Sleep(4 * time.Second)

	module.Start()
}

func (module Flooding_Module) Start() {
	my_addr := module.Correct_events[0]
	module.ReceivedFrom[1] = append(module.ReceivedFrom[1], my_addr)


	proc_id := strings.Split(my_addr, ":")[1]
	Clock = Clock + 1
	printClock(proc_id, Clock, nil, "Generating propose (1)")

	module.Receive()
	module.Send(Prop)
}

func (module Flooding_Module) Send(typeMessage string) {
	switch typeM := typeMessage; typeM {
	case Prop:
		prop := Proposal{From: module.Correct_events[0], Number: generateRandom()}
		module.Proposal(prop)
	case Decided:
		module.Decided()
	case Crash:
		module.Crash()
	default:
		fmt.Printf("Outra opcao")
	}
}

func (module Flooding_Module) Receive() {
	go func() {
		for {
			in := <-module.BEB.Ind

			if module.FailRound != -1 && module.FailRound == Round {
				return
			}

			message := strings.Split(in.Message, "§")
			in.From = message[1]
			in.Message = message[0]
			var data JSON
			json.Unmarshal([]byte(in.Message), &data)

			round := int(data["round"].(float64))

			proc_id := strings.Split(module.Correct_events[0], ":")[1]
			clock := strconv.Itoa(int(data["clock"].(float64)))
			sender_id := strings.Split(in.From, ":")[1]
			sender_clock := []string{`"` + sender_id + `":"` + clock + `"`}

			switch typeM := data["type"]; typeM {
			case Prop:
				if !InSet(in.From, module.ReceivedFrom[round]){
					module.ReceivedFrom[round] = append(module.ReceivedFrom[round], in.From)
				}
				
				value := reflect.ValueOf(data["data"])
				
				for i := 0; i < value.Len(); i++ {
					mapped := value.Index((i)).Interface().(map[string]interface{})
					from := mapped["from"].(string)
					number := int(mapped["number"].(float64))
					
					prop_received := Proposal{From: from, Number: number}
					
					module.Proposals[round-1] = append(module.Proposals[round-1], prop_received)
				}


				message := "Received proposals from " + sender_id + " (" + strconv.Itoa(Round) + ")"

				Clock = Clock + 1
				printClock(proc_id, Clock, sender_clock, message)

				module.CheckAndDecide()

			case Decided:
				message := "Received decided from " + sender_id + " (" + strconv.Itoa(Round) + ")"
				Clock = Clock + 1
				printClock(proc_id, Clock, sender_clock, message)

				if IsInCorrects(module.Correct_events, in.From) && Decision == -1 {
					Decision = int(data["data"].(float64))
					module.Decided()
				}

				
				if !InSet(in.From, Proccess) {
					Proccess = append(Proccess, in.From)
				}

				if IsSubSet(Proccess, module.Correct_events) && len(Proccess) == len(module.Correct_events) - 1 {

					Decision = -1
					Round    = 1
					Proccess = []string{}

					n_proccess := len(module.Correct_events)
					module.ReceivedFrom = make([][]string, n_proccess)
					module.Proposals = make([][]Proposal, n_proccess)
					module.ReceivedFrom[0] = module.Correct_events

					time.Sleep(4 * time.Second)

					module.Start()

					break;
				}

			case Crash:
				for i, v := range module.Correct_events {
					if v == in.From {
						module.Correct_events = append(module.Correct_events[:i], module.Correct_events[i+1:]...)
						break
					}
				}

			default:
				fmt.Printf("Outra opcao")
			}
		}

	}()
}

func (module Flooding_Module) Proposal(v Proposal) {
	module.Proposals[0] = append(module.Proposals[0], v)

	Clock = Clock + 1

	data := make(JSON)
	data["data"] = module.Proposals[0]
	data["clock"] = Clock
	data["type"] = Prop
	data["round"] = Round

	messageJson, _ := json.Marshal(data)
	encoded := string(messageJson)
	proc_id := strings.Split(module.Correct_events[0], ":")[1]

	printClock(proc_id, Clock, nil, "Sending proposals (1)")

	addresses := module.Correct_events[1:]

	if module.FailRound == Round {
		addresses = []string{module.Correct_events[1]}
	}

	req := BestEffortBroadcast_Req_Message{
		Addresses: addresses,
		Message:   encoded + "§" + module.Correct_events[0]}
	module.BEB.Req <- req

	if module.FailRound == Round {
		module.Crash()
	}
}

func (module Flooding_Module) Decided() {

	Clock = Clock + 1

	data := make(JSON)
	data["data"] = Decision
	data["clock"] = Clock
	data["type"] = Decided
	data["round"] = Round

	messageJson, _ := json.Marshal(data)
	encoded := string(messageJson)

	proc_id := strings.Split(module.Correct_events[0], ":")[1]
	message := "Send decision (" + strconv.Itoa(Round) + ")"

	printClock(proc_id, Clock, nil, message)

	go func() {
		req := BestEffortBroadcast_Req_Message{
			Addresses: module.Correct_events[1:],
			Message:   encoded + "§" + module.Correct_events[0]}

		module.BEB.Req <- req
	}()

}

func (module Flooding_Module) CheckAndDecide() {
	if IsSubSet(module.Correct_events, module.ReceivedFrom[Round]) && Decision == -1 {
		if IsEqualSet(module.ReceivedFrom[Round], module.ReceivedFrom[Round-1]) {
			Decision = Min(module.Proposals[Round-1])
			proc_id := strings.Split(module.Correct_events[0], ":")[1]
			Clock = Clock + 1
			printClock(proc_id, Clock, nil, "Decided ("+strconv.Itoa(Round)+")")
			module.Send(Decided)

		} else {
			Clock = Clock + 1
			Round = Round + 1

			data := make(JSON)
			data["data"] = module.Proposals[Round-2]
			data["type"] = Prop
			data["round"] = Round
			data["clock"] = Clock

			messageJson, _ := json.Marshal(data)
			encoded := string(messageJson)
			proc_id := strings.Split(module.Correct_events[0], ":")[1]

			printClock(proc_id, Clock, nil, "Sending proposals ("+strconv.Itoa(Round)+")")

			req := BestEffortBroadcast_Req_Message{
				Addresses: module.Correct_events[1:],
				Message:   encoded + "§" + module.Correct_events[0]}
			module.BEB.Req <- req
		}

	}
}

func (module Flooding_Module) Crash() {

	Clock = Clock + 1

	data := make(JSON)
	data["data"] = ""
	data["type"] = Crash
	data["round"] = Round
	data["clock"] = Clock

	messageJson, _ := json.Marshal(data)
	encoded := string(messageJson)
	proc_id := strings.Split(module.Correct_events[0], ":")[1]
	message := "Crashed (" + strconv.Itoa(Round) + ")"

	printClock(proc_id, Clock, nil, message)

	req := BestEffortBroadcast_Req_Message{
		Addresses: module.Correct_events[1:],
		Message:   encoded + "§" + module.Correct_events[0]}
	module.BEB.Req <- req

	time.Sleep(1 * time.Second)
	os.Exit(0)
}

func IsInCorrects(corrects []string, process string) bool {
	for _, item := range corrects {
		if item == process {
			return true
		}
	}
	return false
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

func InSet(a string, b []string) bool {
	for _, itemb := range b {
		if a == itemb {
			return true
		}
	}

	return false
}

func generateRandom() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	return r1.Intn(100)
}

func printClock(port string, clock int, clocks_arr []string, event string) {

	clocks := `{"` + port + `":` + strconv.Itoa(clock) + `,`
	if clocks_arr != nil {
		for _, sender := range clocks_arr {
			clocks += sender + `,`
		}
	}
	clocks = clocks[:len(clocks)-1]
	clocks += `}`
	fmt.Println("[" + time.Now().Format("01/02/06 15:04:05.0000") + "] [" + port + "] " + clocks + " " + port + " " + event)
}
