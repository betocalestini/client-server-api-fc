package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	UsdBrl `json:"usdbrl"`
}

type UsdBrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Dolar struct {
	Bid string
}

const url = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

func OpenConnectionWithDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal("Erro ao fazer a conexão com o Banco de dados: ", err)
	}
	return db
}

func CreateTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE if not exists cotacoes (id INTEGER PRIMARY KEY, code TEXT, codein TEXT, name TEXT, high TEXT, low TEXT, var_bid TEXT, pct_change TEXT, bid TEXT, ask TEXT, timestamp TEXT, create_date TEXT)`)
	if err != nil {
		log.Fatal("Erro ao criar tabelas no Banco de dados: ", err)
	}
	fmt.Println("Tabela criada com sucesso!")
}

func InsertData(cotacao Cotacao) error {
	//criando o contexto
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	//abrindo a conexão com o banco de dados
	db := OpenConnectionWithDB()
	defer db.Close()

	//preparando o statemant
	stmt, err := db.Prepare("insert into cotacoes (code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) values (?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	//executando a requisição passando o contexto, e atribuindo à struct criada
	_, err = stmt.ExecContext(ctx, &cotacao.Code, &cotacao.Codein, &cotacao.Name, &cotacao.High, &cotacao.Low, &cotacao.VarBid, &cotacao.PctChange, &cotacao.Bid, &cotacao.Ask, &cotacao.Timestamp, &cotacao.CreateDate)
	if err != nil {
		return err
	}
	return nil
}

func CotacaoGetHandler(w http.ResponseWriter, r *http.Request) {
	//criando o contexto
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	//criando a requisição
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Mensagem": "Ocorreu um erro interno"}`))
		fmt.Println("Erro ao preparar a requisição: ", err)
		return
	}
	req.Header.Add("Accept", "application/json")

	//executando a requisição
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte(`{"Mensagem": "Não foi possível obter os dados dentro do tempo limite"}`))
		fmt.Println("Erro ao executar a requisição: ", err)
		return
	}
	defer resp.Body.Close()

	//lendo o corpo da resposta
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Mensagem": "Ocorreu um erro ao processar a resposta"}`))
		fmt.Println("Erro ao ler o corpo da resposta: ", err)
		return
	}

	//convertendo o json da resposta
	var cotacao Cotacao
	err = json.Unmarshal(res, &cotacao)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Mensagem": "Ocorreu um erro ao processar a resposta"}`))
		fmt.Println("Erro ao converter a resposta: ", err)
		return
	}

	//inserindo os dados no banco de dados
	err = InsertData(cotacao)
	if err != nil {
		fmt.Println("Erro ao inserir os dados no Banco: ", err)
	}

	//enviando o Bid ao cliente
	var dolar Dolar
	dolar.Bid = cotacao.Bid
	retornoClient, err := json.Marshal(dolar)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Mensagem": "Ocorreu um erro ao processar a resposta"}`))
		fmt.Println("Erro ao converter a struct: ", err)
		return
	}
	w.Write([]byte(retornoClient))
}

func main() {
	mux := http.NewServeMux()

	db := OpenConnectionWithDB()
	defer db.Close()
	CreateTable(db)

	mux.HandleFunc("/cotacao", CotacaoGetHandler)

	fmt.Println("Servidor rodando ...")
	http.ListenAndServe(":8080", mux)
}
