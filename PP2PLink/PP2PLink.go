/*
  Construido como parte da disciplina de Sistemas Distribuidos
  Semestre 2018/2  -  PUCRS - Escola Politecnica
  Estudantes:  Andre Antonitsch e Rafael Copstein
  Professor: Fernando Dotti  (www.inf.pucrs.br/~fldotti)
  Algoritmo baseado no livro:
  Introduction to Reliable and Secure Distributed Programming
  Christian Cachin, Rachid Gerraoui, Luis Rodrigues

  Semestre 2019/1
  Melhorado como parte da disciplina de Sistemas Distribuídos
  Reaproveita conexões TCP já abertas
  Vinicius Sesti e Gabriel Waengertner

  Semestre 2020/1
  Separa mensagens de qualquer tamanho atee 4 digitos.
  Sender envia tamanho no formato 4 digitos (preenche com 0s a esquerda)
  Receiver recebe 4 digitos, calcula tamanho do buffer a receber,
  e recebe com io.ReadFull para preencher o buffer.
*/

package PP2PLink

import (
	"fmt"
	"io"
	"net"
	"strconv"
)

type PP2PLink_Req_Message struct {
	To      string
	Message string
}

type PP2PLink_Ind_Message struct {
	From    string
	Message string
}

type PP2PLink struct {
	Ind   chan PP2PLink_Ind_Message
	Req   chan PP2PLink_Req_Message
	Run   bool
	Cache map[string]net.Conn // cache de conexoes - reaproveita conexao com destino ao inves de abrir outra
}

func (module *PP2PLink) Init(address string) {

	fmt.Println("Init PP2PLink!")
	if module.Run {
		return
	}

	module.Cache = make(map[string]net.Conn)
	module.Run = true
	module.Start(address)
}

func (module *PP2PLink) Start(address string) {

	go func() {

		listen, _ := net.Listen("tcp4", address)
		for {

			// aceita repetidamente tentativas novas de conexao
			conn, err := listen.Accept()

			// para cada conexao lanca rotina de tratamento
			go func() {
				// repetidamente recebe mensagens na conexao TCP (sem fechar)
				// e passa para cima
				for {
					if err != nil {
						fmt.Println(err)
						continue
					}
					bufTam := make([]byte, 4) // le tamanho da mensagem
					_, err := io.ReadFull(conn, bufTam)
					if err != nil {
						fmt.Println(err)
						continue
					}
					tam, err := strconv.Atoi(string(bufTam))
					bufMsg := make([]byte, tam)        // declara buffer do tamanho exato
					_, err = io.ReadFull(conn, bufMsg) // le do tamanho do buffer ou da erro
					if err != nil {
						fmt.Println(err)
						continue
					}
					msg := PP2PLink_Ind_Message{
						From:    conn.RemoteAddr().String(),
						Message: string(bufMsg)}

					module.Ind <- msg // repassa mensagem para modulo superior
				}
			}()
		}
	}()

	go func() {
		for {
			message := <-module.Req
			module.Send(message)
		}
	}()

}

func (module *PP2PLink) Send(message PP2PLink_Req_Message) {
	var conn net.Conn
	var ok bool
	var err error

	// ja existe uma conexao aberta para aquele destinatario?
	if conn, ok = module.Cache[message.To]; ok {
		//fmt.Printf("Reusing connection %v to %v\n", conn.LocalAddr(), message.To)
	} else { // se nao tiver, abre e guarda na cache
		conn, err = net.Dial("tcp", message.To)
		if err != nil {
			fmt.Println(err)
			return
		}
		module.Cache[message.To] = conn
	}
	//fmt.Println("  ", message.Message, "          tam: ", len(message.Message))
	strSize := strconv.Itoa(len(message.Message))
	for len(strSize) < 4 {
		strSize = "0" + strSize
	}
	if !(len(strSize) == 4) {
		fmt.Println("ERROR AT PPLINK MESSAGE SIZE CALCULATION - INVALID MESSAGES MAY BE IN TRANSIT")
	}
	fmt.Fprintf(conn, strSize)         // write size of message to connection - the size of strSize is allwais 4 char
	fmt.Fprintf(conn, message.Message) // write message to connection
}
