package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
)

func ImportEmployees(w http.ResponseWriter, r *http.Request){
	err:= r.ParseMultipartForm(10<<20)
	if err!=nil{
		http.Error(w, "Failed", http.StatusBadRequest)
		return
	}
	file, _, err:=r.FormFile("file")
	if err!=nil{
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return

	}
	defer file.Close()

	reader:=csv.NewReader(file)
	records, err:= reader.ReadAll()
	if err!=nil{
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	fmt.Printf("Uploaded rows: %d\n", len(records))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully imported %d records"))

}
