package utils

import (
	"fmt"
	"github.com/google/uuid"
)

type CCEEFault struct {
	SchemaFault string
	ErrorCode   string
	Message     string
	Uri         string
}

func CreateServerError(faultCode string, faultString string, faultActor string, fault CCEEFault) []byte {
	transactionID, _ := uuid.NewUUID()
	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">
    <env:Body>
        <env:Fault>
            <faultcode>%s</faultcode>
            <faultstring>%s</faultstring>
            <faultactor>%s</faultactor>
            <detail>
                <%s xmlns:flt="http://xmlns.energia.org.br/FM">
                    <flt:errorCode>%s</flt:errorCode>
                    <flt:message>%s</flt:message>
                    <flt:uri>%s</flt:uri>
                    <flt:transactionId>%s</flt:transactionId>
                </%s>
            </detail>
        </env:Fault>
    </env:Body>
</env:Envelope>`, faultCode, faultString, faultActor,
		fault.SchemaFault,
		fault.ErrorCode, fault.Message, fault.Uri,
		transactionID.String(),
		fault.SchemaFault))
}
