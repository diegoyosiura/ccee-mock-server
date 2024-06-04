package utils

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
)

func DecodeBodyData(request *http.Request) (response []byte) {
	var err error

	switch request.Header.Get("Content-Encoding") {
	case "gzip":
		var reader *gzip.Reader
		reader, err = gzip.NewReader(request.Body)

		if err != nil {
			log.Println(fmt.Sprintf("Erro ao extrair informações: %s", err))
			_ = reader.Close()
			return nil
		}
		response, err = io.ReadAll(reader)
		err = reader.Close()
		if err != nil {
			log.Println(fmt.Sprintf("Erro ao fechar escritor: %s", err))
			return nil
		}
	default:
		response, err = io.ReadAll(request.Body)
		if err != nil {
			log.Println(fmt.Sprintf("Erro ao ler requisição: %s", err))
			return nil
		}
	}

	return RemoveNamespacesCCEEBytes([]byte(CleanXMLCCEEString(string(response))))
}
