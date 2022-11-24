package mime_types

const (
	charsetUTF8 = "charset=UTF-8"
	// PROPFIND Method can be used on collection and property resources. 
	PROPFIND = "PROPFIND"
	// REPORT Method can be used to get information about resource
	REPORT = "REPORT"
)

// MIME - Multipurpose Internet Mail Extensions
// Wiki - https://ru.wikipedia.org/wiki/MIME
// RFC  - https://www.rfc-editor.org/rfc/rfc2045
const (
	MIMEApplicationJSON = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML = "application/xml"
	MIMEApplicationXMLCharsetUTF8 = MIMEApplicationXML + "; " + charsetUTF8
	MIMEApplicationForm = "application/x-www-from-urlencoder"
	MIMEApplicationProtobuf = "application/protobuf"
	MIMEApplicationMsgPack = "application/msgpack"
	
	MIMETextXML			= "text/xml"
	MIMETextXMLCharsetUTF8 = MIMETextXML + "; " + charsetUTF8
	MIMETextHTML = "text/html"
	MIMETextHTMLCharsetUTF8 = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain 				= "text/plain"
	MIMETextPlainCharsetUTF8 = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm = "multipart/from-data"
	MIMEOctetStream = "application/octet-stream"
)
