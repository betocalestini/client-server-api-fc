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

type Bid struct {
	Value string
}

const url = "http://localhost:8080/cotacao"

func RequisitarCotacao() []byte {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatal("Erro ao preparar a requisição: ", err)
	}

	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Erro ao executar a requisição: ", err)
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Erro ao ler o corpo da resposta: ", err)
	}
	return res
}

func GravarArquivoDeCotacao(res []byte) {
	var bid Bid

	err := json.Unmarshal(res, &bid)
	if err != nil {
		log.Fatal("Erro ao converter o JSON da resposta: ", err)
	}

	file, err := os.OpenFile("cotacao.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Erro ao criar o arquivo: ", err)

	}
	defer file.Close()

	//Aqui poderia ser colocado a data hora da pesquisa, para manter no arquivo, porém como não foi pedido na atividade, não foi feito.
	_, err = file.WriteString(fmt.Sprintf("Dólar: %v \n", bid.Value))
	if err != nil {
		log.Fatal("Erro ao gravar no arquivo: ", err)
	}

}

func main() {
	res := RequisitarCotacao()
	GravarArquivoDeCotacao(res)

}
