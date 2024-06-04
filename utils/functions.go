package utils

import (
	"regexp"
	"strings"
)

func CleanXMLCCEEString(xml string) string {
	reWitespaces := regexp.MustCompile(`[\t\n]`)
	reSpaces := regexp.MustCompile(`>\s+<`)
	xml = reWitespaces.ReplaceAllString(xml, " ")
	xml = reSpaces.ReplaceAllString(xml, "><")
	return xml
}
func RemoveNamespacesCCEEBytes(xml []byte) []byte {
	return []byte(RemoveNamespacesCCEEString(string(xml)))
}

func RemoveNamespacesCCEEString(xml string) string {
	xml = strings.ToLower(xml)
	reEnv := regexp.MustCompile("<soapenv:envelope.*?>.*?<soapenv:header")
	reNS := regexp.MustCompile("<[0-9a-z-_.]*?:")
	reNSE := regexp.MustCompile("</[0-9a-z-_.]*?:")
	reSpace := regexp.MustCompile("[\\s+\\n\\r]")
	reSpaceS := regexp.MustCompile("\\s+")
	reSpaceTag := regexp.MustCompile("> <")
	return reSpaceTag.ReplaceAllString(reSpaceS.ReplaceAllString(reSpace.ReplaceAllString(reNSE.ReplaceAllString(reNS.ReplaceAllString(reEnv.ReplaceAllString(xml,
		"<envelope><header"), "<"), "</"), " "), " "), "><")
}
