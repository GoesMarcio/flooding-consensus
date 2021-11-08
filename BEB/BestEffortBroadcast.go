package BestEffortBroadcast

/*
  Construido como parte da disciplina de Sistemas Distribuidos
  Semestre 2018/2  -  PUCRS - Escola Politecnica
  Estudantes:  Andre Antonitsch e Rafael Copstein
  Professor: Fernando Dotti  (www.inf.pucrs.br/~fldotti)
  Algoritmo baseado no livro:
  Introduction to Reliable and Secure Distributed Programming
  Christian Cachin, Rachid Gerraoui, Luis Rodrigues

  Para uso vide ao final do arquivo, ou aplicacao chat.go que usa este
*/

import (
	PP2PLink "../PP2PLink"
)

type BestEffortBroadcast_Req_Message struct {
	Addresses []string
	Message   string
}

type BestEffortBroadcast_Ind_Message struct {
	From    string
	Message string
}

type BestEffortBroadcast_Module struct {
	Ind      chan BestEffortBroadcast_Ind_Message
	Req      chan BestEffortBroadcast_Req_Message
	Pp2plink PP2PLink.PP2PLink
}

func (module BestEffortBroadcast_Module) Init(address string) {

	// fmt.Println("Init BEB!")
	module.Pp2plink = PP2PLink.PP2PLink{
		Req: make(chan PP2PLink.PP2PLink_Req_Message),
		Ind: make(chan PP2PLink.PP2PLink_Ind_Message)}
	module.Pp2plink.Init(address)
	module.Start()
}

func (module BestEffortBroadcast_Module) Start() {

	go func() {
		for {
			select {
			case y := <-module.Req:
				module.Broadcast(y)
			case y := <-module.Pp2plink.Ind:
				module.Deliver(PP2PLink2BEB(y))
			}
		}
	}()
}

func (module BestEffortBroadcast_Module) Broadcast(message BestEffortBroadcast_Req_Message) {

	for i := 0; i < len(message.Addresses); i++ {
		msg := BEB2PP2PLink(message)
		msg.To = message.Addresses[i]
		module.Pp2plink.Req <- msg
		// fmt.Println("Sent to " + message.Addresses[i])
	}
}

func (module BestEffortBroadcast_Module) Deliver(message BestEffortBroadcast_Ind_Message) {

	// fmt.Println("Received '" + message.Message + "' from " + message.From)
	module.Ind <- message
	// fmt.Println("# End BEB Received")
}

func BEB2PP2PLink(message BestEffortBroadcast_Req_Message) PP2PLink.PP2PLink_Req_Message {

	return PP2PLink.PP2PLink_Req_Message{
		To:      message.Addresses[0],
		Message: message.Message}

}

func PP2PLink2BEB(message PP2PLink.PP2PLink_Ind_Message) BestEffortBroadcast_Ind_Message {

	return BestEffortBroadcast_Ind_Message{
		From:    message.From,
		Message: message.Message}
}

/*
func main() {

	if (len(os.Args) < 2) {
		fmt.Println("Please specify at least one address:port!")
		return
	}

	addresses := os.Args[1:]
	fmt.Println(addresses)

	mod := BestEffortBroadcast_Module{
		Req: make(chan BestEffortBroadcast_Req_Message),
		Ind: make(chan BestEffortBroadcast_Ind_Message) }
	mod.Init(addresses[0])

	msg := BestEffortBroadcast_Req_Message{
		Addresses: addresses,
		Message: "BATATA!" }

	yy := make(chan string)
	mod.Req <- msg
	<- yy
}
*/
