package response

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/apperror"
	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestMethodNotAllowed_Status405(t *testing.T) {
	w := httptest.NewRecorder()
	MethodNotAllowed(w, "method not allowed")
	testkit.AssertStatus(t, w, http.StatusMethodNotAllowed)
}

func TestGone_Status410(t *testing.T) {
	w := httptest.NewRecorder()
	Gone(w, "resource deleted")
	testkit.AssertStatus(t, w, http.StatusGone)
}

func TestGatewayTimeout_Status504(t *testing.T) {
	w := httptest.NewRecorder()
	GatewayTimeout(w, "upstream timeout")
	testkit.AssertStatus(t, w, http.StatusGatewayTimeout)
}

func TestPaginated_EnvelopeFields(t *testing.T) {
	w := httptest.NewRecorder()
	Paginated(w, []string{"a", "b", "c"}, 50, 2, 3)

	testkit.RequireEqual(t, w.Code, http.StatusOK)
	var got PagedResult[string]
	testkit.AssertJSON(t, w.Body.Bytes(), &got)
	testkit.AssertEqual(t, got.Total, 50)
	testkit.AssertEqual(t, got.Page, 2)
	testkit.AssertEqual(t, got.Limit, 3)
	testkit.AssertLen(t, got.Data, 3)
	testkit.AssertTrue(t, got.HasMore)
}

func TestPaginated_HasMore_False_OnLastPage(t *testing.T) {
	w := httptest.NewRecorder()
	Paginated(w, []int{4, 5}, 5, 2, 3)

	var got PagedResult[int]
	testkit.AssertJSON(t, w.Body.Bytes(), &got)
	testkit.AssertFalse(t, got.HasMore)
}

func TestPaginated_EmptyData(t *testing.T) {
	w := httptest.NewRecorder()
	Paginated(w, []string{}, 0, 1, 10)

	var got PagedResult[string]
	testkit.AssertJSON(t, w.Body.Bytes(), &got)
	testkit.AssertEqual(t, got.Total, 0)
	testkit.AssertFalse(t, got.HasMore)
}

func TestAppError_WithAppError(t *testing.T) {
	w := httptest.NewRecorder()
	AppError(w, apperror.NotFound("user not found"))

	testkit.AssertStatus(t, w, http.StatusNotFound)
	testkit.AssertHeader(t, w, "Content-Type", "application/json")

	var body map[string]any
	testkit.AssertJSON(t, w.Body.Bytes(), &body)
	testkit.AssertEqual(t, body["code"], "NOT_FOUND")
	testkit.AssertEqual(t, body["message"], "user not found")
}

func TestAppError_WithWrappedAppError(t *testing.T) {
	inner := apperror.BadRequest("invalid input")
	wrapped := fmt.Errorf("handler: %w", inner)

	w := httptest.NewRecorder()
	AppError(w, wrapped)

	testkit.AssertStatus(t, w, http.StatusBadRequest)
}

func TestAppError_WithGenericError(t *testing.T) {
	w := httptest.NewRecorder()
	AppError(w, errors.New("something broke"))

	testkit.AssertStatus(t, w, http.StatusInternalServerError)

	var body map[string]string
	testkit.AssertJSON(t, w.Body.Bytes(), &body)
	testkit.AssertEqual(t, body["error"], "Internal server error")
}

func TestRedirect_Found(t *testing.T) {
	w := httptest.NewRecorder()
	r := testkit.NewRequest("GET", "/old", nil)
	Redirect(w, r, "/new", http.StatusFound)
	testkit.AssertEqual(t, w.Code, http.StatusFound)
	testkit.AssertEqual(t, w.Header().Get("Location"), "/new")
}

func TestRedirect_MovedPermanently(t *testing.T) {
	w := httptest.NewRecorder()
	r := testkit.NewRequest("GET", "/legacy", nil)
	Redirect(w, r, "https://example.com/new", http.StatusMovedPermanently)
	testkit.AssertEqual(t, w.Code, http.StatusMovedPermanently)
	testkit.AssertEqual(t, w.Header().Get("Location"), "https://example.com/new")
}

func TestText(t *testing.T) {
	w := httptest.NewRecorder()
	Text(w, http.StatusOK, "hello world")
	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
	testkit.AssertEqual(t, w.Body.String(), "hello world")
}

func TestText_CustomStatus(t *testing.T) {
	w := httptest.NewRecorder()
	Text(w, http.StatusAccepted, "processing")
	testkit.AssertEqual(t, w.Code, http.StatusAccepted)
	testkit.AssertEqual(t, w.Body.String(), "processing")
}

func TestHTML(t *testing.T) {
	w := httptest.NewRecorder()
	HTML(w, http.StatusOK, "<h1>Hello</h1>")
	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
	testkit.AssertEqual(t, w.Body.String(), "<h1>Hello</h1>")
}

func TestHTML_CustomStatus(t *testing.T) {
	w := httptest.NewRecorder()
	HTML(w, http.StatusNotFound, "<h1>Not Found</h1>")
	testkit.AssertEqual(t, w.Code, http.StatusNotFound)
	testkit.AssertEqual(t, w.Body.String(), "<h1>Not Found</h1>")
}

func TestStream(t *testing.T) {
	w := httptest.NewRecorder()
	body := strings.NewReader("file-content-here")
	Stream(w, http.StatusOK, "application/octet-stream", body)

	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Header().Get("Content-Type"), "application/octet-stream")
	testkit.AssertEqual(t, w.Body.String(), "file-content-here")
}

func TestStream_CSV(t *testing.T) {
	w := httptest.NewRecorder()
	csv := "name,age\nAlice,30\n"
	Stream(w, http.StatusOK, "text/csv", strings.NewReader(csv))

	testkit.AssertEqual(t, w.Header().Get("Content-Type"), "text/csv")
	testkit.AssertEqual(t, w.Body.String(), csv)
}

func TestDownload(t *testing.T) {
	w := httptest.NewRecorder()
	body := strings.NewReader("csv,data\n1,2\n")
	Download(w, "text/csv", "report.csv", body)

	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Header().Get("Content-Type"), "text/csv")
	testkit.AssertEqual(t, w.Header().Get("Content-Disposition"), `attachment; filename="report.csv"`)
	testkit.AssertEqual(t, w.Body.String(), "csv,data\n1,2\n")
}

func TestXML(t *testing.T) {
	type Item struct {
		XMLName xml.Name `xml:"item"`
		Name    string   `xml:"name"`
		Value   int      `xml:"value"`
	}
	w := httptest.NewRecorder()
	XML(w, http.StatusOK, Item{Name: "test", Value: 42})

	testkit.AssertStatus(t, w, http.StatusOK)
	testkit.AssertHeader(t, w, "Content-Type", "application/xml; charset=utf-8")
	testkit.AssertContains(t, w.Body.String(), "<?xml")
	testkit.AssertContains(t, w.Body.String(), "<name>test</name>")
	testkit.AssertContains(t, w.Body.String(), "<value>42</value>")
}

func TestCreatedWithLocation(t *testing.T) {
	w := httptest.NewRecorder()
	CreatedWithLocation(w, "/api/users/42", map[string]string{"id": "42"})

	testkit.AssertStatus(t, w, http.StatusCreated)
	testkit.AssertHeader(t, w, "Location", "/api/users/42")
	testkit.AssertContains(t, w.Body.String(), `"id"`)
}

func TestNotModified(t *testing.T) {
	w := httptest.NewRecorder()
	NotModified(w)
	testkit.AssertStatus(t, w, http.StatusNotModified)
	testkit.AssertEqual(t, w.Body.Len(), 0)
}

func TestSeeOther(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/submit", nil)
	SeeOther(w, r, "/result")
	testkit.AssertStatus(t, w, http.StatusSeeOther)
	testkit.AssertHeader(t, w, "Location", "/result")
}
