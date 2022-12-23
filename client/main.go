package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Dolar struct {
	Bid string `json:"bid"`
}

type MensagemErro struct {
	Mensagem string `json:"mensagem"`
}

const url = "http://localhost:8080/cotacao"

func RequisitarCotacao() []byte {
	//criando o contexto
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)
	defer cancel()

	//preparando a requisiçãop
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatal("Erro ao preparar a requisição: ", err)
	}
	req.Header.Add("Accept", "application/json")

	//executando a requisição
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Erro ao executar a requisição: ", err)
	}
	defer resp.Body.Close()

	//lendo o corpo da resposta
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Erro ao ler o corpo da resposta: ", err)
	}
	return res
}

func GravarArquivoDeCotacao(res []byte) {
	var dolar Dolar
	var serverError MensagemErro
	var resultado string

	err := json.Unmarshal(res, &dolar)
	if err != nil {
		log.Fatal("Erro ao converter o JSON da resposta: ", err)
	}

	if dolar.Bid == "" {
		err := json.Unmarshal(res, &serverError)
		if err != nil {
			fmt.Println("Erro ao converter a mensagem: ", err)
		}
		resultado = fmt.Sprintf("Erro: %v \n", serverError.Mensagem)
		fmt.Println("Erro ao receber a resposta do servidor: ", resultado)
	} else {
		resultado = fmt.Sprintf("Dólar: %v \n", dolar.Bid)
	}

	file, err := os.OpenFile("cotacao.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Erro ao criar o arquivo: ", err)
	}
	defer file.Close()

	//Aqui poderia ser colocado a data hora da pesquisa, para manter no arquivo, porém como não foi pedido na atividade, não foi feito.
	_, err = file.WriteString(resultado)
	if err != nil {
		log.Fatal("Erro ao gravar no arquivo: ", err)
	}
}

func main() {
	res := RequisitarCotacao()
	GravarArquivoDeCotacao(res)
}
