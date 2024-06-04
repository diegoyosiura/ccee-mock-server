package pld

import (
	"ampereconsultoria.com.br/ccee/mock/utils"
	"context"
	"fmt"
	mongoDb "github.com/ampere-consultoria/sagace-v2-mongod"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"math"
	"math/rand/v2"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func PLDHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/xml; charset=utf-8")

	for h, _ := range request.Header {
		if strings.ToUpper(h) == "SOAPACTION" && strings.ToUpper(request.Header.Get(h)) == "LISTARPLD" {
			body := utils.DecodeBodyData(request)
			_, _ = response.Write(parsePLD(request.Context(), response, body))
			return
		}
	}

	response.WriteHeader(500)
	_, _ = response.Write(utils.CreateServerError("Server.2001", "Acesso Negado", "02", utils.CCEEFault{
		SchemaFault: "tns:securityFault",
		ErrorCode:   "2001",
		Message:     "URI de entrada inv&#225;lida",
		Uri:         "/ws/prec/PLDBSv1",
	}))
}

func parsePLD(ctx context.Context, response http.ResponseWriter, request []byte) []byte {

	rePerfilAgente := regexp.MustCompile(`<codigoperfilagente>(.*)</codigoperfilagente>`)
	reUsuario := regexp.MustCompile(`<username>(.*)</username>`)
	rePaginacaoPagina := regexp.MustCompile(`<paginacao>.*?<numero>([0-9]+)</numero>.*?</paginacao>`)
	rePaginacaoQuantidade := regexp.MustCompile(`<paginacao>.*?<quantidadeitens>([0-9]+)</quantidadeitens>.*?</paginacao>`)
	reTipoPLDSemanal := regexp.MustCompile(`<valores>.*?<valor>.*?<tipo>semanal</tipo>.*?</valor></valores>`)

	perfilRq := rePerfilAgente.FindSubmatch(request)
	usuarioRq := reUsuario.FindSubmatch(request)
	paginacaoPaginaRq := rePaginacaoPagina.FindSubmatch(request)
	paginacaoQuantidadeRq := rePaginacaoQuantidade.FindSubmatch(request)

	if len(perfilRq) != 2 || len(usuarioRq) != 2 || len(paginacaoPaginaRq) != 2 || len(paginacaoQuantidadeRq) != 2 {
		response.WriteHeader(500)
		return utils.CreateServerError("Server.2001", "Acesso Negado", "02", utils.CCEEFault{
			SchemaFault: "tns:securityFault",
			ErrorCode:   "2001",
			Message:     "Rejected: Usuario nao esta autorizado a usar o codigoPerfilAgente.",
			Uri:         "/ws/prec/PLDBSv1",
		})
	}

	paginacaoPagina, _ := strconv.Atoi(string(paginacaoPaginaRq[1]))
	paginacaoQuantidade, _ := strconv.Atoi(string(paginacaoQuantidadeRq[1]))
	perfilAgente, _ := strconv.Atoi(string(perfilRq[1]))
	//usuario, _ := string(usuarioRq[1])

	reInicio := regexp.MustCompile(`<inicio>([0-9]{4}-[0-9]{2}-[0-9]{2}).*?([0-9]{2}:[0-9]{2}:[0-9]{2})</inicio>`)
	reFim := regexp.MustCompile(`<fim>([0-9]{4}-[0-9]{2}-[0-9]{2}).*?([0-9]{2}:[0-9]{2}:[0-9]{2})</fim>`)
	reSemanal := regexp.MustCompile(`<tipo>semanal</tipo>`)
	reHorario := regexp.MustCompile(`<tipo>horario</tipo>`)

	if reSemanal.Match(request) || reHorario.Match(request) {
		gpInicio := reInicio.FindSubmatch(request)
		gpFim := reFim.FindSubmatch(request)

		if len(gpInicio) == 3 && len(gpFim) == 3 {
			inicio, errInicio := time.Parse(utils.DateTimeLayout, fmt.Sprintf("%s %s", gpInicio[1], gpInicio[2]))
			fim, errFim := time.Parse(utils.DateTimeLayout, fmt.Sprintf("%s %s", gpFim[1], gpFim[2]))

			if errInicio == nil && errFim == nil {
				if inicio.Before(fim) {
					if inicio.Year() != fim.Year() {
						response.WriteHeader(500)
						return utils.CreateServerError("Server.3006", "Par&#226;metros Invalidos", "4", utils.CCEEFault{
							SchemaFault: "flt:invalidParametersFault",
							ErrorCode:   "3006",
							Message:     "N&#227;o &#233; permitido solicitar um per&#237;odo que ultrapasse o limite de um mesmo ano. Verifique os par&#226;metros informados.",
							Uri:         "/ws/prec/PLDBSv1",
						})
					}
					var outputch <-chan []byte
					if reTipoPLDSemanal.Match(request) {
						outputch = pldSemanal(ctx, response, perfilAgente, paginacaoPagina, paginacaoQuantidade, inicio, fim)
					}
					outputch = pldHorario(ctx, response, perfilAgente, paginacaoPagina, paginacaoQuantidade, inicio, fim)

					for {
						select {
						case <-ctx.Done():
							return utils.CreateServerError("Server.3006", "Par&#226;metros Invalidos", "4", utils.CCEEFault{
								SchemaFault: "flt:invalidParametersFault",
								ErrorCode:   "3006",
								Message:     "A data final deve ser maior que a data inicial do per&#237;odo. Verifique os par&#226;metros informados.",
								Uri:         "/ws/prec/PLDBSv1",
							})
						case responseData := <-outputch:
							return responseData
						}
					}

				} else {
					response.WriteHeader(500)
					return utils.CreateServerError("Server.3006", "Par&#226;metros Invalidos", "4", utils.CCEEFault{
						SchemaFault: "flt:invalidParametersFault",
						ErrorCode:   "3006",
						Message:     "A data final deve ser maior que a data inicial do per&#237;odo. Verifique os par&#226;metros informados.",
						Uri:         "/ws/prec/PLDBSv1",
					})
				}
			}
		}
	}

	response.WriteHeader(500)
	return utils.CreateServerError("Server.3006", "Par&#226;metros Invalidos", "4", utils.CCEEFault{
		SchemaFault: "flt:invalidParametersFault",
		ErrorCode:   "3006",
		Message:     "Informe um tipo de PLD v&#225;lido. Verifique os par&#226;metros informados.",
		Uri:         "/ws/prec/PLDBSv1",
	})
}

func pldHorario(ctx context.Context, response http.ResponseWriter, codPerfilAgente int,
	pagina int,
	itens int,
	inicio time.Time, fim time.Time) <-chan []byte {

	output := make(chan []byte)

	go func() {
		defer close(output)

		collectionPLD := mongoDb.MongoConnection.GetCollection("ccee_mock", "pld_horario")
		totalItens := 0

		inicioRef, _ := time.Parse("2006-01-02 15:04:05", inicio.Format("2006-01-02 15:04:05"))
		for inicioRef.Unix() <= fim.Unix() {
			totalItens += 24
			inicioRef = inicioRef.Add(24 * time.Hour)
		}

		totalPaginas := totalItens / itens

		if totalItens%itens > 0 {
			totalPaginas++
		}

		if pagina > totalPaginas {
			response.WriteHeader(500)
			output <- utils.CreateServerError("Server.3006", "Par&#226;metros Invalidos", "4", utils.CCEEFault{
				SchemaFault: "flt:noDataFoundFault",
				ErrorCode:   "3001",
				Message:     "Nenhum PLD encontrado",
				Uri:         "/ws/prec/PLDBSv1",
			})
			return
		}

		itemCountStart := ((pagina - 1) * itens) + 1
		itemCounter := 1

		inicioRef, _ = time.Parse("2006-01-02 15:04:05", inicio.Format("2006-01-02 15:04:05"))
		var plds []string
		for inicioRef.Unix() <= fim.Unix() {
			var xmlPldValores []string

			filter := bson.D{
				{"ano", inicioRef.Year()},
				{"mes", int(inicioRef.Month())},
				{"dia", int(inicioRef.Day())},
				{"hora", 1},
			}

			var resultPLD []PLD
			cursor, err := collectionPLD.Find(ctx, filter)

			if err == nil {
				err = cursor.All(context.TODO(), &resultPLD)
			}
			if err != nil || len(resultPLD) == 0 {
				var insertPLD []interface{}
				for i := 1; i <= 4; i++ {
					pldValue := math.Round(rand.Float64() * 600)
					xmlPldValores = append(xmlPldValores, fmt.Sprintf(XMLPLDResponseValor,
						i, map[int]string{1: "SUDESTE", 2: "SUL", 3: "NORDESTE", 4: "NORTE"}[i],
						"HORARIO",
						pldValue,
					))
					for h := 0; h <= 23; h++ {
						insertPLD = append(insertPLD, PLD{
							Ano:        int64(inicioRef.Year()),
							Mes:        int64(inicioRef.Month()),
							Dia:        int64(inicioRef.Day()),
							Hora:       int64(h),
							Submercado: int64(i),
							Valor:      pldValue,
						})
					}
				}
				_, _ = collectionPLD.InsertMany(ctx, insertPLD)
			} else {
				for _, pldQr := range resultPLD {
					xmlPldValores = append(xmlPldValores, fmt.Sprintf(XMLPLDResponseValor,
						pldQr.Submercado, map[int]string{1: "SUDESTE", 2: "SUL", 3: "NORDESTE", 4: "NORTE"}[int(pldQr.Submercado)],
						"HORARIO",
						pldQr.Valor,
					))
				}
			}

			for i := 0; i <= 23; i++ {
				if itemCounter >= itemCountStart && itemCounter < itemCountStart+itens {
					if i != 23 {
						plds = append(plds, fmt.Sprintf(XMLPLDResponseBody,
							inicioRef.Format("2006-01-02"), i,
							inicioRef.Format("2006-01-02"), i+1, strings.Join(xmlPldValores, "\n")))
					} else {
						plds = append(plds, fmt.Sprintf(XMLPLDResponseBody,
							inicioRef.Format("2006-01-02"), i,
							inicioRef.Add(24*time.Hour).Format("2006-01-02"),
							0, strings.Join(xmlPldValores, "\n")))
					}
				}

				itemCounter++
			}
			inicioRef = inicioRef.Add(24 * time.Hour)
		}

		transactionID, _ := uuid.NewUUID()
		xml := fmt.Sprintf(XMLPLDResponse,
			codPerfilAgente,
			transactionID.String(), pagina, itens, totalPaginas, totalItens,
			strings.Join(plds, "\n"))

		response.WriteHeader(200)
		output <- []byte(xml)
	}()

	return output
}

func pldSemanal(ctx context.Context, response http.ResponseWriter, codPerfilAgente int,
	pagina int,
	itens int,
	inicio time.Time, fim time.Time) <-chan []byte {

	output := make(chan []byte)

	go func() {
		defer close(output)
		totalItens := 0

		inicioRef, _ := time.Parse("2006-01-02 15:04:05", inicio.Format("2006-01-02 15:04:05"))
		for inicioRef.Unix() <= fim.Unix() {
			totalItens += 24
			inicioRef = inicioRef.Add(24 * time.Hour)
		}

		totalPaginas := totalItens / itens

		if totalItens%itens > 0 {
			totalPaginas++
		}

		if pagina > totalPaginas {
			response.WriteHeader(500)
			output <- utils.CreateServerError("Server.3006", "Par&#226;metros Invalidos", "4", utils.CCEEFault{
				SchemaFault: "flt:noDataFoundFault",
				ErrorCode:   "3001",
				Message:     "Nenhum PLD encontrado",
				Uri:         "/ws/prec/PLDBSv1",
			})

			return
		}

		itemCountStart := ((pagina - 1) * itens) + 1
		itemCounter := 1

		inicioRef, _ = time.Parse("2006-01-02 15:04:05", inicio.Format("2006-01-02 15:04:05"))
		var plds []string
		for inicioRef.Unix() <= fim.Unix() {
			var xmlPldValores []string

			for i := 1; i <= 4; i++ {
				xmlPldValores = append(xmlPldValores, fmt.Sprintf(XMLPLDResponseValor,
					i, map[int]string{1: "SUDESTE", 2: "SUL", 3: "NORDESTE", 4: "NORTE"}[i],
					"MEDIA_SEMANAL",
					math.Round(rand.Float64()*600),
				))
			}
			for i := 0; i <= 23; i++ {
				if itemCounter >= itemCountStart && itemCounter < itemCountStart+itens {
					if i != 23 {
						plds = append(plds, fmt.Sprintf(XMLPLDResponseBody,
							inicioRef.Format("2006-01-02"), i,
							inicioRef.Format("2006-01-02"), i+1, strings.Join(xmlPldValores, "\n")))
					} else {
						plds = append(plds, fmt.Sprintf(XMLPLDResponseBody,
							inicioRef.Format("2006-01-02"), i,
							inicioRef.Add(24*time.Hour).Format("2006-01-02"),
							0, strings.Join(xmlPldValores, "\n")))
					}
				}

				itemCounter++
			}
			inicioRef = inicioRef.Add(24 * time.Hour)
		}

		transactionID, _ := uuid.NewUUID()
		xml := fmt.Sprintf(XMLPLDResponse,
			codPerfilAgente,
			transactionID.String(), pagina, itens, totalPaginas, totalItens,
			strings.Join(plds, "\n"))

		response.WriteHeader(200)
		output <- []byte(xml)
	}()

	return output
}
