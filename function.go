package p

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

type LastUpdate struct {
	Atualizado string
	Rede       string
}

type RespJson struct {
	LastUpdate string `json:"lastupdate"`
	NextUpdate string `json:"nextupdate"`
	Now        string `json:"now"`
}

func CheckUpdate(w http.ResponseWriter, r *http.Request) {

	// When I ran the function within the same Firestore project in GCP, it was not necessary to authenticate the function to connect to the Firestore database
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "PROJECT_ID")
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	// Set qual Collection e Docuumento queremos ler do banco NoSQL no Firestore
	FirebaseCollection := "COLLECTION-ID"
	FirebaseDoc := "DOCUMENT-ID"
	doc, err := client.Collection(FirebaseCollection).Doc(FirebaseDoc).Get(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	// Define layout de data (padrão GoLang) - Estas data e hora passada na variável layout é o padrão do GoLang para escolher o formato da data
	layout := "02/01 15:04"
	timezone, _ := time.LoadLocation("America/Sao_Paulo")
	time.Local = timezone
	now := time.Now().In(timezone)

	j, err := json.Marshal(doc.Data())
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	// Converte String para Data e set o timezone correto
	resp := LastUpdate{}
	json.Unmarshal([]byte(j), &resp)
	Lastdate, err := time.ParseInLocation(layout, resp.Atualizado, timezone)
	if err != nil {
		fmt.Println(err)
	}

	LastUpdate := Lastdate.AddDate(time.Now().Year(), 0, 0)
	// Soma 45 minutos desde a data e hora da ultima atualização como Target para monitoramento da próxima atualização
	NextDateUpdate := Lastdate.AddDate(time.Now().Year(), 0, 0).Add(time.Minute * 45)

	js := RespJson{
		LastUpdate: LastUpdate.Format(layout),
		NextUpdate: NextDateUpdate.Format(layout),
		Now:        now.Format(layout),
	}

	RespJsonData, err := json.Marshal(js)
	if err != nil {
		fmt.Println(err)
	}

	// Compara se data/hora de próxima atualizar é maior que a data/hora atual e retorna HTTP Status Code para na API
	if NextDateUpdate.After(now) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(RespJsonData))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(RespJsonData))
	}
}
