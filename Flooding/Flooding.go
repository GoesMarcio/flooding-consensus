package Flooding

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"os"

	. "../BEB"
)

const (
	Prop    = "proposal"
	Deliver = "deliver"
	Decided = "decided"
	Crash   = "crash"
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
	CrashBool      bool
	FailRound      int
	Clock          int
}

type JSON map[string]interface{}

func (module Flooding_Module) Init(addresses []string, crash bool, failRound int) {
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
	
	module.CrashBool = crash
	module.FailRound = failRound
	module.Clock = 1

	module.BEB.Init(addresses[0])
	time.Sleep(4 * time.Second)

	proc_id := strings.Split(module.Correct_events[0], ":")[1]
	printClock(proc_id, module.Clock, nil)

	module.Clock = module.Clock + 1
	module.Send(Prop)
	module.Receive()
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
			message := strings.Split(in.Message, "§")
			in.From = message[1]
			in.Message = message[0]

			var data JSON
			json.Unmarshal([]byte(in.Message), &data)

			round := int(data["round"].(float64))

			switch typeM := data["type"]; typeM {
			case Prop:
				// fmt.Println("recebi uma proposta de: ", in.From)
				module.ReceivedFrom[round] = append(module.ReceivedFrom[round], in.From)

				value := reflect.ValueOf(data["data"])

				for i := 0; i < value.Len(); i++ {
					mapped := value.Index((i)).Interface().(map[string]interface{})
					from := mapped["from"].(string)
					number := int(mapped["number"].(float64))

					prop_received := Proposal{From: from, Number: number}

					module.Proposals[round-1] = append(module.Proposals[round-1], prop_received)
				}

				// fmt.Println("o cara q me mandou ta no clock", data["clock"])

				// printClock(proc_id, module.Clock, nil)

				// var procs = ""
				// for _, item := range module.ReceivedFrom[round] {
				// 	procs = procs + `"` + strings.Split(item, ":")[1] + `":` + strconv.Itoa(module.Round) + `,`
				// }
				// procs = procs[:len(procs)-1]

				proc_id := strings.Split(module.Correct_events[0], ":")[1]
				clock := strconv.Itoa(int(data["clock"].(float64)))
				sender_id := []string{strings.Split(in.From, ":")[1] + " : " + clock}
				// fmt.Println(proc_id + ` {` + procs + `}`)
				printClock(proc_id, module.Clock, sender_id)

				module.Clock = module.Clock + 1

				module.CheckAndDecide()

			case Decided:
				fmt.Println("recebi uma decisao de: ", in.From)
				// fmt.Println(module.Decision)
				// fmt.Println(time.Now().Format("01/02/06 15:04:05") + " Receive Decision")

				if IsInCorrects(module.Correct_events, in.From) && module.Decision == -1 {
					module.Decision = int(data["data"].(float64))
					module.Decided()
					// fmt.Printf("\nReceived from %s decision: %d\n", in.From, module.Decision)
				}
			
			case Crash:
				fmt.Println("recebi um crash de: ", in.From)

				for i, v := range module.Correct_events {
					if v == in.From {
						module.Correct_events = append(module.Correct_events[:i], module.Correct_events[i+1:]...)
						break
					}
				}

				// fmt.Println(module.Correct_events)

			default:
				fmt.Printf("Outra opcao")
			}
		}
	}()
}

func (module Flooding_Module) Proposal(v Proposal) {
	go func() {
		module.Proposals[0] = append(module.Proposals[0], v)
		
		data := make(JSON)
		data["data"] = module.Proposals[0]
		data["clock"] = module.Clock
		data["type"] = Prop
		data["round"] = module.Round
		
		messageJson, _ := json.Marshal(data)
		encoded := string(messageJson)
		proc_id := strings.Split(module.Correct_events[0], ":")[1]
		
		printClock(proc_id, module.Clock, nil)
		
		module.Clock = module.Clock + 1

		addresses := []string{}
		if module.CrashBool && module.FailRound == module.Round{
			addresses = []string{module.Correct_events[1]}
		}else{
			addresses = module.Correct_events
		}
		req := BestEffortBroadcast_Req_Message{
			Addresses: addresses,
			Message:   encoded + "§" + module.Correct_events[0]}
		module.BEB.Req <- req

		if module.CrashBool && module.FailRound == module.Round{
			module.Crash()
		}
	}()

}

func (module Flooding_Module) Decided() {
	go func() {

		data := make(JSON)
		data["data"] = module.Decision
		data["clock"] = module.Clock
		data["type"] = Decided
		data["round"] = module.Round

		messageJson, _ := json.Marshal(data)
		encoded := string(messageJson)

		proc_id := strings.Split(module.Correct_events[0], ":")[1]

		printClock(proc_id, module.Clock, nil)
		module.Clock = module.Clock + 1


		req := BestEffortBroadcast_Req_Message{
			Addresses: module.Correct_events[1:],
			Message:   encoded + "§" + module.Correct_events[0]}
		module.BEB.Req <- req

	}()

}

func (module Flooding_Module) CheckAndDecide() {
	// fmt.Println(module.Correct_events)
	// fmt.Println(module.ReceivedFrom[module.Round])
	if IsSubSet(module.Correct_events, module.ReceivedFrom[module.Round]) && module.Decision == -1 {
		if IsEqualSet(module.ReceivedFrom[module.Round], module.ReceivedFrom[module.Round-1]) {
			module.Decision = Min(module.Proposals[module.Round-1])
			module.Send(Decided)
			// fmt.Printf("\nProcess %s decided: %d\n", module.Correct_events[0], module.Decision)

		} else {
			module.Round = module.Round + 1
			data := make(JSON)
			data["data"] = module.Proposals
			data["type"] = Prop
			data["round"] = module.Round
			
			fmt.Println("cheguei aqui cpx", module.Round)
		}

	}
}

func (module Flooding_Module) Crash() {
	go func() {
		fmt.Println("Crashei")

		data := make(JSON)
		data["data"] = ""
		data["type"] = Crash
		data["round"] = module.Round

		messageJson, _ := json.Marshal(data)
		encoded := string(messageJson)
		// proc_id := strings.Split(module.Correct_events[0], ":")[1]

		// fmt.Println(time.Now().Format("01/02/06 15:04:05") + " Propose")
		// fmt.Println(proc_id + ` {"` + proc_id + `":1}`)

		req := BestEffortBroadcast_Req_Message{
			Addresses: module.Correct_events[1:],
			Message:   encoded + "§" + module.Correct_events[0]}
		module.BEB.Req <- req

		fmt.Println("enviando meu crash pra: ", module.Correct_events[1:])
		
		time.Sleep(1 * time.Second)

		os.Exit(0)
	}()
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

func generateRandom() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	return r1.Intn(100)
}


func printClock(port string, clock int, clocks_arr []string){
	
	clocks := `{"` + port + `":`+strconv.Itoa(clock)+`,`
	if clocks_arr != nil{
		for _, sender := range clocks_arr{
			clocks += sender + ","
		}
	}
	clocks = clocks[:len(clocks)-1]
	clocks += `}`
	fmt.Println("[" + time.Now().Format("01/02/06 15:04:05.000") + "] [" + port + "] "+clocks)
}