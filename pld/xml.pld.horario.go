package pld

const XMLPLDResponse = `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:hdr="http://xmlns.energia.org.br/MH/v1">
    <soapenv:Header>
        <mh:messageHeader xmlns:mh="http://xmlns.energia.org.br/MH/v1">
            <mh:codigoPerfilAgente>%d</mh:codigoPerfilAgente>
            <mh:transactionId>%s</mh:transactionId>
        </mh:messageHeader>
        <hdr:paginacao>
            <hdr:numero>%d</hdr:numero>
            <hdr:quantidadeItens>%d</hdr:quantidadeItens>
            <hdr:totalPaginas>%d</hdr:totalPaginas>
            <hdr:quantidadeTotalItens>%d</hdr:quantidadeTotalItens>
        </hdr:paginacao>
    </soapenv:Header>
    <soapenv:Body>
        <bm:listarPLDResponse xmlns:bm="http://xmlns.energia.org.br/BM/v1" xmlns:bo="http://xmlns.energia.org.br/BO/v1">
            <bm:plds>
%s
            </bm:plds>
        </bm:listarPLDResponse>
    </soapenv:Body>
</soapenv:Envelope>`

const XMLPLDResponseBody = `<bm:pld>
    <bo:vigencia>
        <bo:inicio>%sT%02d:00:00-02:00</bo:inicio>
        <bo:fim>%sT%02d:00:00-02:00</bo:fim>
    </bo:vigencia>
    <bo:valores>
%s
    </bo:valores>
</bm:pld>`

const XMLPLDResponseValor = `        <bo:valor>
            <bo:indicadorRedeEletrica>false</bo:indicadorRedeEletrica>
            <bo:submercado>
                <bo:codigo>%d</bo:codigo>
                <bo:nome>%s</bo:nome>
            </bo:submercado>
            <bo:tipo>%s</bo:tipo>
            <bo:valor>
                <bo:codigo>BRL</bo:codigo>
                <bo:valor>%.2f</bo:valor>
            </bo:valor>
        </bo:valor>`
