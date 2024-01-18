package ochp

import (
	"fmt"
	"net/http"
)

/* All errors returned by OCHP prodecures implement OchpErr, which is a superset
 * of the error interface. Not all fields may be available depending on how early
 * the error happened. If the underlying HTTP request succeeded, HttpResponse and
 * optionally HttpResponseBody (if configured) will return nonzero values. If
 * the underlying SOAP response was decoded successfully, ResultCode and
 * ResultDescription will return nonzero values.
 */
type OchpErr interface {
	// Error describes the error
	Error() string
	// Unwrap may return an underlying error that caused this OCHP error. These are
	// generally errors coming from underlying libraries, such as http or xml.
	Unwrap() error
	// ResultCode may return the resultCode field from the OCHP Result object
	ResultCode() string
	// ResultDescription may return the resultDescription field from the OCHP Result object
	ResultDescription() string
	// HttpResponse may return the response of the underlying HTTP request
	HttpResponse() *http.Response
	// HttpResponseBody may return a copy of the response body of the underlying
	// HTTP request. As copying this introduces computational and memory overhead,
	// it is only returned if the client has been configured to do so.
	HttpResponseBody() []byte
}

/* Err is returned when the underlying HTTP request was not able to succeed for
 * any reason, including when the request never happened due to errors earlier
 * in the code. It thus returns zero values for all HTTP or OCHP related
 * methods. It generally contains a wrapped error returned by underlying
 * libraries.
 */
type Err struct {
	wrapped error
}

func (e Err) Error() string {
	if e.wrapped != nil {
		return e.wrapped.Error()
	}
	return "unknown error"
}

func (e Err) Unwrap() error {
	return e.wrapped
}

func (e Err) ResultCode() string {
	return ""
}

func (e Err) ResultDescription() string {
	return ""
}

func (e Err) HttpResponse() *http.Response {
	return nil
}

func (e Err) HttpResponseBody() []byte {
	return nil
}

/* ErrDecode is returned when the underlying HTTP request suceeded, but it
 * wasn't able to be deserialized properly. It thus returns zero values for all
 * OCHP related methods. It generally contains a wrapped error returned by
 * underlying libraries.
 */
type ErrDecode struct {
	Err
	httpResponse     *http.Response
	httpResponseBody []byte
}

func (e ErrDecode) Error() string {
	statusDetails := ""
	if e.httpResponse != nil {
		statusDetails = fmt.Sprint(" with HTTP status", e.httpResponse.StatusCode)
	}

	wrappedErrorDetails := ""
	if err2 := e.Unwrap(); err2 != nil {
		wrappedErrorDetails = ":\n" + err2.Error()
	}

	return fmt.Sprintf("server responded%s, but response was unable to be decoded successfully%s", statusDetails, wrappedErrorDetails)
}

func (e ErrDecode) HttpResponse() *http.Response {
	return e.httpResponse
}

func (e ErrDecode) HttpResponseBody() []byte {
	return e.httpResponseBody
}

// ErrEmptyResponse is returned when the underlying HTTP request suceeded, but
// the server returned an empty response body. This is a special case, as the
// OCHP server does this upon bad requests in certain cases.
type ErrEmptyResponse struct{ ErrDecode }

func (e ErrEmptyResponse) Error() string {
	return "Server unexpectedly responded with an empty response. This indicates an internal error. This might be caused by a bad request"
}

/* ochpErr contains the data and logic used by any errors that happen after
 * decoding of the XML response. Any errors relating to issues found in the
 * parsed OCHP response use this.
 */
type ochpErr struct {
	ErrDecode
	resultCode        string
	resultDescription string
}

func (e ochpErr) Error() string {
	if e.httpResponse == nil || e.resultCode == "" && e.resultDescription == "" {
		return e.ErrDecode.Error()
	}

	if e.httpResponse.StatusCode >= 200 && e.httpResponse.StatusCode < 300 {
		return fmt.Sprintf("server responded with http status %d and resultCode \"%s\": %s", e.httpResponse.StatusCode, e.resultCode, e.resultDescription)
	}

	return fmt.Sprintf("server responded with resultCode \"%s\": %s", e.resultCode, e.resultDescription)
}

func (e ochpErr) ResultCode() string {
	return e.resultCode
}

func (e ochpErr) ResultDescription() string {
	return e.resultDescription
}

// ErrPartly is returned when the OCHP Result object had resultCode "partly".
// For certain procedures, this error may be returned together with data
// describing which data the procedure wasn't able to process.
type ErrPartly struct{ ochpErr }

// ErrNotFound is returned when the OCHP Result object had resultCode "not-found"
type ErrNotFound struct{ ochpErr }

// ErrNotAuthorized is returned when the OCHP Result object had resultCode "not-authorized"
type ErrNotAuthorized struct{ ochpErr }

// ErrNotSupported is returned when the OCHP Result object had resultCode "not-supported"
type ErrNotSupported struct{ ochpErr }

// ErrInvalidId is returned when the OCHP Result object had resultCode "invalid-id"
type ErrInvalidId struct{ ochpErr }

// ErrServer is returned when the OCHP Result object had resultCode "server"
type ErrServer struct{ ochpErr }

// ErrFormat is returned when the OCHP Result object had resultCode "format"
type ErrFormat struct{ ochpErr }

// ErrRoaming is returned when the OCHP Result object had resultCode "roaming"
type ErrRoaming struct{ ochpErr }

// ErrUnknownResultCode is returned when the OCHP Result object had a resultCode
// that was not "ok", but also not one of the other known resultCodes.
type ErrUnknownResultCode struct{ ochpErr }

// ErrHttp is returned if the underlying HTTP request returned a non-2XX status.
// There may or may not be any ResultCode, ResultDescription or other data present.
type ErrHttp struct{ ochpErr }

// assert that base types implement expected interfaces
var _ error = Err{}
var _ error = ErrDecode{}
var _ error = ochpErr{}
var _ OchpErr = Err{}
var _ OchpErr = ErrDecode{}
var _ OchpErr = ochpErr{}
