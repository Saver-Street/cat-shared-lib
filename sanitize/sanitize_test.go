package sanitize

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestDocFilename_Normal(t *testing.T) {
	testkit.AssertEqual(t, DocFilename("report.pdf"), "report.pdf")
}

func TestDocFilename_Empty(t *testing.T) {
	testkit.AssertEqual(t, DocFilename(""), "unnamed")
}

func TestDocFilename_PathTraversal(t *testing.T) {
	testkit.AssertEqual(t, DocFilename("../../../etc/passwd"), "passwd")
}

func TestDocFilename_ControlChars(t *testing.T) {
	testkit.AssertEqual(t, DocFilename("file\x00name\x01.txt"), "filename.txt")
}

func TestDocFilename_AllControlChars(t *testing.T) {
	testkit.AssertEqual(t, DocFilename("\x01\x02\x03"), "unnamed")
}

func TestDocFilename_DEL(t *testing.T) {
	testkit.AssertEqual(t, DocFilename("file\x7fname.txt"), "filename.txt")
}

func TestNilIfEmpty_Empty(t *testing.T) {
	testkit.AssertNil(t, NilIfEmpty(""))
}

func TestNilIfEmpty_NonEmpty(t *testing.T) {
	got := NilIfEmpty("hello")
	testkit.RequireNotNil(t, got)
	testkit.AssertEqual(t, *got, "hello")
}

func TestIsDuplicateKey_True(t *testing.T) {
	err := &pgconn.PgError{Code: "23505"}
	testkit.AssertTrue(t, IsDuplicateKey(err))
}

func TestIsDuplicateKey_False(t *testing.T) {
	err := errors.New("some other error")
	testkit.AssertFalse(t, IsDuplicateKey(err))
}

func TestIsDuplicateKey_Nil(t *testing.T) {
	testkit.AssertFalse(t, IsDuplicateKey(nil))
}

func TestDocFilename_Unicode(t *testing.T) {
	testkit.AssertEqual(t, DocFilename("résumé_日本語.pdf"), "résumé_日本語.pdf")
}

func TestDocFilename_LongFilename(t *testing.T) {
	long := strings.Repeat("a", 300) + ".pdf"
	testkit.AssertEqual(t, len(DocFilename(long)), len(long))
}

func TestDocFilename_OnlySpaces(t *testing.T) {
	testkit.AssertEqual(t, DocFilename("   "), "   ")
}

func TestDocFilename_DotFile(t *testing.T) {
	testkit.AssertEqual(t, DocFilename(".gitignore"), ".gitignore")
}

func TestDocFilename_Emoji(t *testing.T) {
	testkit.AssertEqual(t, DocFilename("📄document.pdf"), "📄document.pdf")
}

func TestIsDuplicateKey_WrappedError(t *testing.T) {
	inner := &pgconn.PgError{Code: "23505"}
	wrapped := fmt.Errorf("insert failed: %w", inner)
	testkit.AssertTrue(t, IsDuplicateKey(wrapped))
}

func TestNilIfEmpty_Whitespace(t *testing.T) {
	got := NilIfEmpty(" ")
	testkit.RequireNotNil(t, got)
	testkit.AssertEqual(t, *got, " ")
}

// --- Benchmarks ---

func BenchmarkDocFilename(b *testing.B) {
	for b.Loop() {
		DocFilename("../path/to/my\x00file\x7fname.pdf")
	}
}

func BenchmarkNilIfEmpty(b *testing.B) {
	for b.Loop() {
		NilIfEmpty("")
		NilIfEmpty("hello")
	}
}

func BenchmarkIsDuplicateKey(b *testing.B) {
	err := errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")
	for b.Loop() {
		IsDuplicateKey(err)
	}
}

func TestTrimAndNilIfEmpty_Empty(t *testing.T) {
	testkit.AssertNil(t, TrimAndNilIfEmpty(""))
}

func TestTrimAndNilIfEmpty_WhitespaceOnly(t *testing.T) {
	testkit.AssertNil(t, TrimAndNilIfEmpty("   "))
}

func TestTrimAndNilIfEmpty_NonEmpty(t *testing.T) {
	got := TrimAndNilIfEmpty("  hello  ")
	testkit.RequireNotNil(t, got)
	testkit.AssertEqual(t, *got, "hello")
}

func TestTrimAndNilIfEmpty_NoTrimNeeded(t *testing.T) {
	got := TrimAndNilIfEmpty("world")
	testkit.RequireNotNil(t, got)
	testkit.AssertEqual(t, *got, "world")
}

// --- TruncateFilename table-driven tests ---

func TestTruncateFilename(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"normal with extension", "report.pdf", 20, "report.pdf"},
		{"truncated preserves ext", "averylongfilename.txt", 10, "averyl.txt"},
		{"exact length", "file.txt", 8, "file.txt"},
		{"shorter than maxLen", "hi.go", 100, "hi.go"},
		{"maxLen zero", "file.txt", 0, ""},
		{"maxLen negative", "file.txt", -5, ""},
		{"empty name", "", 10, ""},
		{"empty name zero maxLen", "", 0, ""},
		{"extension longer than maxLen", "x.verylongext", 3, ".verylongext"},
		{"extension equal to maxLen", "x.tx", 3, ".tx"},
		{"no extension", "readme", 4, "read"},
		{"no extension within limit", "readme", 100, "readme"},
		{"unicode filename", "résumé.pdf", 7, "rés.pdf"},
		{"unicode ext preserved", "日本語ファイル.txt", 6, "日本.txt"},
		{"emoji filename", "📄📝.txt", 5, "📄.txt"},
		{"dot file", ".gitignore", 5, ".gitignore"},
		{"single char and ext", "a.go", 4, "a.go"},
		{"maxLen 1 with ext", "file.txt", 1, ".txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testkit.AssertEqual(t, TruncateFilename(tt.input, tt.maxLen), tt.want)
		})
	}
}

// --- MaxLength table-driven tests ---

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"normal truncation", "hello world", 5, "hello"},
		{"within limit", "hello", 10, "hello"},
		{"exact limit", "hello", 5, "hello"},
		{"maxLen zero", "hello", 0, ""},
		{"maxLen negative", "hello", -1, ""},
		{"empty string", "", 10, ""},
		{"empty string zero maxLen", "", 0, ""},
		{"unicode truncation", "héllo", 3, "hél"},
		{"unicode within limit", "héllo", 100, "héllo"},
		{"CJK characters", "日本語テスト", 3, "日本語"},
		{"emoji truncation", "👍🎉🚀💡", 2, "👍🎉"},
		{"single char", "x", 1, "x"},
		{"maxLen 1 long string", "abcdef", 1, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testkit.AssertEqual(t, MaxLength(tt.input, tt.maxLen), tt.want)
		})
	}
}

// --- SanitizeEmail table-driven tests ---

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal email", "alice@example.com", "alice@example.com"},
		{"leading/trailing whitespace", "  alice@example.com  ", "alice@example.com"},
		{"uppercase", "Alice@Example.COM", "alice@example.com"},
		{"whitespace and uppercase", "  User@Example.COM  ", "user@example.com"},
		{"empty string", "", ""},
		{"only whitespace", "   ", ""},
		{"tab whitespace", "\tuser@test.com\t", "user@test.com"},
		{"mixed whitespace", " \t user@test.com \n ", "user@test.com"},
		{"already lowercase no spaces", "test@test.io", "test@test.io"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testkit.AssertEqual(t, SanitizeEmail(tt.input), tt.want)
		})
	}
}

// --- IsDatabaseError table-driven tests ---

func TestIsDatabaseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
		want bool
	}{
		{"unique violation match", &pgconn.PgError{Code: "23505"}, "23505", true},
		{"foreign key violation", &pgconn.PgError{Code: "23503"}, "23503", true},
		{"not null violation", &pgconn.PgError{Code: "23502"}, "23502", true},
		{"check violation", &pgconn.PgError{Code: "23514"}, "23514", true},
		{"non-matching code", &pgconn.PgError{Code: "42000"}, "23505", false},
		{"nil error", nil, "23505", false},
		{"nil error different code", nil, "23503", false},
		{"plain error not pgconn", errors.New("some error"), "23505", false},
		{"wrapped pgconn error", fmt.Errorf("insert failed: %w", &pgconn.PgError{Code: "23505"}), "23505", true},
		{"wrong code for unique", &pgconn.PgError{Code: "23505"}, "23503", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testkit.AssertEqual(t, IsDatabaseError(tt.err, tt.code), tt.want)
		})
	}
}

func BenchmarkTruncateFilename(b *testing.B) {
	for b.Loop() {
		TruncateFilename("long_document_name.pdf", 15)
	}
}

func BenchmarkMaxLength(b *testing.B) {
	s := "a string that is fairly long for benchmarking purposes"
	for b.Loop() {
		MaxLength(s, 20)
	}
}

func BenchmarkSanitizeEmail(b *testing.B) {
	for b.Loop() {
		SanitizeEmail("  User@Example.COM  ")
	}
}

func BenchmarkIsDatabaseError(b *testing.B) {
	err := &pgconn.PgError{Code: "23505"}
	for b.Loop() {
		IsDatabaseError(err, "23505")
	}
}

func BenchmarkTrimAndNilIfEmpty(b *testing.B) {
	for b.Loop() {
		TrimAndNilIfEmpty("  hello  ")
	}
}

func TestNullString_Nil(t *testing.T) {
	testkit.AssertEqual(t, NullString(nil, "default"), "default")
}

func TestNullString_Value(t *testing.T) {
	testkit.AssertEqual(t, NullString(testkit.Ptr("hello"), "default"), "hello")
}

func TestNullString_Empty(t *testing.T) {
	testkit.AssertEqual(t, NullString(testkit.Ptr(""), "default"), "")
}

func TestNullInt64_Nil(t *testing.T) {
	testkit.AssertEqual(t, NullInt64(nil, 42), int64(42))
}

func TestNullInt64_Value(t *testing.T) {
	testkit.AssertEqual(t, NullInt64(testkit.Ptr(int64(99)), 0), int64(99))
}

func TestNullInt64_Zero(t *testing.T) {
	testkit.AssertEqual(t, NullInt64(testkit.Ptr(int64(0)), 42), int64(0))
}

func TestNullBool_Nil(t *testing.T) {
	testkit.AssertTrue(t, NullBool(nil, true))
}

func TestNullBool_False(t *testing.T) {
	testkit.AssertFalse(t, NullBool(testkit.Ptr(false), true))
}

func TestNullBool_True(t *testing.T) {
	testkit.AssertTrue(t, NullBool(testkit.Ptr(true), false))
}

func TestDeref_Nil(t *testing.T) {
	var p *string
	testkit.AssertEqual(t, Deref(p, "default"), "default")
}
func TestDeref_NonNil(t *testing.T) {
	s := "hello"
	testkit.AssertEqual(t, Deref(&s, "default"), "hello")
}
func TestDeref_NilInt(t *testing.T) {
	var p *int
	testkit.AssertEqual(t, Deref(p, 42), 42)
}
func TestDeref_NonNilInt(t *testing.T) {
	n := 7
	testkit.AssertEqual(t, Deref(&n, 42), 7)
}
func TestDeref_NilBool(t *testing.T) {
	var p *bool
	testkit.AssertTrue(t, Deref(p, true))
}
func TestDeref_NonNilBool(t *testing.T) {
	b := false
	testkit.AssertFalse(t, Deref(&b, true))
}

func TestStripHTML_Basic(t *testing.T) {
	testkit.AssertEqual(t, StripHTML("<p>Hello</p>"), "Hello")
}

func TestStripHTML_NoTags(t *testing.T) {
	testkit.AssertEqual(t, StripHTML("plain text"), "plain text")
}

func TestStripHTML_Empty(t *testing.T) {
	testkit.AssertEqual(t, StripHTML(""), "")
}

func TestStripHTML_NestedTags(t *testing.T) {
	testkit.AssertEqual(t, StripHTML("<div><b>bold</b> text</div>"), "bold text")
}

func TestStripHTML_Attributes(t *testing.T) {
	testkit.AssertEqual(t, StripHTML("<a href=\"https://example.com\">link</a>"), "link")
}

func TestStripHTML_SelfClosing(t *testing.T) {
	testkit.AssertEqual(t, StripHTML("line1<br/>line2"), "line1line2")
}

func TestStripHTML_Script(t *testing.T) {
	testkit.AssertEqual(t, StripHTML("<script>alert(1)</script>safe"), "alert(1)safe")
}

func TestStripHTML_OnlyTags(t *testing.T) {
	testkit.AssertEqual(t, StripHTML("<p></p>"), "")
}

func TestMask_CreditCard(t *testing.T) {
	testkit.AssertEqual(t, Mask("4111111111111111", 4), "************1111")
}

func TestMask_ShortString(t *testing.T) {
	testkit.AssertEqual(t, Mask("abc", 4), "abc")
}

func TestMask_ExactLength(t *testing.T) {
	testkit.AssertEqual(t, Mask("abcd", 4), "abcd")
}

func TestMask_Empty(t *testing.T) {
	testkit.AssertEqual(t, Mask("", 4), "")
}

func TestMask_ZeroVisible(t *testing.T) {
	testkit.AssertEqual(t, Mask("secret", 0), "******")
}

func TestMask_NegativeVisible(t *testing.T) {
	testkit.AssertEqual(t, Mask("key", -1), "***")
}

func TestMask_SingleChar(t *testing.T) {
	testkit.AssertEqual(t, Mask("x", 0), "*")
}

func TestMask_APIKey(t *testing.T) {
	testkit.AssertEqual(t, Mask("sk-abc123def456", 6), "*********def456")
}

func TestTrimStrings(t *testing.T) {
	got := TrimStrings([]string{"  hello ", "world", "  ", ""})
	testkit.AssertLen(t, got, 2)
	testkit.AssertEqual(t, got[0], "hello")
	testkit.AssertEqual(t, got[1], "world")
}

func TestTrimStrings_Empty(t *testing.T) {
	got := TrimStrings([]string{})
	testkit.AssertLen(t, got, 0)
}

func TestTrimStrings_AllEmpty(t *testing.T) {
	got := TrimStrings([]string{" ", "", "\t"})
	testkit.AssertLen(t, got, 0)
}

func TestTrimStrings_Nil(t *testing.T) {
	got := TrimStrings(nil)
	testkit.AssertLen(t, got, 0)
}

func TestSlugify(t *testing.T) {
testkit.AssertEqual(t, Slugify("Hello World"), "hello-world")
}

func TestSlugify_SpecialChars(t *testing.T) {
testkit.AssertEqual(t, Slugify("My Blog Post!!! #1"), "my-blog-post-1")
}

func TestSlugify_ConsecutiveHyphens(t *testing.T) {
testkit.AssertEqual(t, Slugify("a   b   c"), "a-b-c")
}

func TestSlugify_LeadingTrailing(t *testing.T) {
testkit.AssertEqual(t, Slugify("  --hello-- "), "hello")
}

func TestSlugify_Empty(t *testing.T) {
testkit.AssertEqual(t, Slugify(""), "")
}

func TestEscapeHTML(t *testing.T) {
tests := []struct {
name   string
input  string
expect string
}{
{"script tag", `<script>alert("xss")</script>`, `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`},
{"ampersand", `a & b`, `a &amp; b`},
{"single quote", `it's`, `it&#39;s`},
{"plain text", "hello", "hello"},
{"empty", "", ""},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
testkit.AssertEqual(t, EscapeHTML(tt.input), tt.expect)
})
}
}

func TestTruncate_Short(t *testing.T) {
testkit.AssertEqual(t, Truncate("hello", 10), "hello")
}

func TestTruncate_Exact(t *testing.T) {
testkit.AssertEqual(t, Truncate("hello", 5), "hello")
}

func TestTruncate_Long(t *testing.T) {
testkit.AssertEqual(t, Truncate("hello world", 8), "hello w…")
}

func TestTruncate_MaxOne(t *testing.T) {
testkit.AssertEqual(t, Truncate("hello", 1), "…")
}

func TestTruncate_Unicode(t *testing.T) {
testkit.AssertEqual(t, Truncate("héllo wörld", 6), "héllo…")
}

func TestTruncate_Empty(t *testing.T) {
testkit.AssertEqual(t, Truncate("", 5), "")
}

func TestRemoveNonPrintable(t *testing.T) {
tests := []struct {
name   string
input  string
expect string
}{
{"plain text", "hello world", "hello world"},
{"with null byte", "hello\x00world", "helloworld"},
{"with control chars", "line1\x01\x02\x03line2", "line1line2"},
{"preserves newline", "line1\nline2", "line1\nline2"},
{"preserves tab", "col1\tcol2", "col1\tcol2"},
{"empty", "", ""},
{"unicode", "héllo wörld", "héllo wörld"},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
testkit.AssertEqual(t, RemoveNonPrintable(tt.input), tt.expect)
})
}
}

func TestNormalizeWhitespace(t *testing.T) {
tests := []struct {
name   string
input  string
expect string
}{
{"multiple spaces", "hello   world", "hello world"},
{"tabs and newlines", "hello\t\n\tworld", "hello world"},
{"leading trailing", "  hello  ", "hello"},
{"already clean", "hello world", "hello world"},
{"empty", "", ""},
{"only spaces", "   ", ""},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
testkit.AssertEqual(t, NormalizeWhitespace(tt.input), tt.expect)
})
}
}

func TestCamelToSnake(t *testing.T) {
tests := []struct{ in, want string }{
{"", ""},
{"foo", "foo"},
{"fooBar", "foo_bar"},
{"FooBar", "foo_bar"},
{"fooBarBaz", "foo_bar_baz"},
{"HTTPClient", "http_client"},
{"getHTTPResponse", "get_http_response"},
{"XMLParser", "xml_parser"},
{"simpleXML", "simple_xml"},
{"ID", "id"},
{"userID", "user_id"},
}
for _, tt := range tests {
t.Run(tt.in, func(t *testing.T) {
testkit.AssertEqual(t, CamelToSnake(tt.in), tt.want)
})
}
}

func TestSnakeToCamel(t *testing.T) {
tests := []struct{ in, want string }{
{"", ""},
{"foo", "foo"},
{"foo_bar", "fooBar"},
{"foo_bar_baz", "fooBarBaz"},
{"http_client", "httpClient"},
{"USER_ID", "userId"},
{"__double__under__", "doubleUnder"},
{"single", "single"},
}
for _, tt := range tests {
t.Run(tt.in, func(t *testing.T) {
testkit.AssertEqual(t, SnakeToCamel(tt.in), tt.want)
})
}
}

func TestMapKeys(t *testing.T) {
m := map[string]int{
"Hello": 1,
"WORLD": 2,
}
got := MapKeys(m, strings.ToLower)
testkit.AssertEqual(t, got["hello"], 1)
testkit.AssertEqual(t, got["world"], 2)
testkit.AssertLen(t, got, 2)
}

func TestMapKeys_Empty(t *testing.T) {
got := MapKeys(map[string]int{}, strings.ToUpper)
testkit.AssertLen(t, got, 0)
}

func TestFilter(t *testing.T) {
nums := []int{1, 2, 3, 4, 5, 6}
even := Filter(nums, func(n int) bool { return n%2 == 0 })
testkit.AssertLen(t, even, 3)
testkit.AssertEqual(t, even[0], 2)
testkit.AssertEqual(t, even[1], 4)
testkit.AssertEqual(t, even[2], 6)
}

func TestFilter_Empty(t *testing.T) {
result := Filter([]string{}, func(s string) bool { return true })
testkit.AssertLen(t, result, 0)
}

func TestFilter_NoneMatch(t *testing.T) {
result := Filter([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
testkit.AssertLen(t, result, 0)
}

func TestFilter_Strings(t *testing.T) {
words := []string{"hello", "", "world", "", "go"}
nonEmpty := Filter(words, func(s string) bool { return s != "" })
testkit.AssertLen(t, nonEmpty, 3)
}

func TestUnique_Strings(t *testing.T) {
got := Unique([]string{"a", "b", "a", "c", "b"})
testkit.AssertLen(t, got, 3)
testkit.AssertEqual(t, got[0], "a")
testkit.AssertEqual(t, got[1], "b")
testkit.AssertEqual(t, got[2], "c")
}

func TestUnique_Ints(t *testing.T) {
got := Unique([]int{1, 2, 3, 2, 1})
testkit.AssertLen(t, got, 3)
}

func TestUnique_Empty(t *testing.T) {
got := Unique([]string{})
testkit.AssertLen(t, got, 0)
}

func TestUnique_NoDuplicates(t *testing.T) {
got := Unique([]int{1, 2, 3})
testkit.AssertLen(t, got, 3)
}

func TestMap_StringToUpper(t *testing.T) {
got := Map([]string{"a", "b", "c"}, strings.ToUpper)
testkit.AssertLen(t, got, 3)
testkit.AssertEqual(t, got[0], "A")
testkit.AssertEqual(t, got[2], "C")
}

func TestMap_IntToString(t *testing.T) {
got := Map([]int{1, 2, 3}, func(n int) string {
return fmt.Sprintf("%d", n)
})
testkit.AssertLen(t, got, 3)
testkit.AssertEqual(t, got[0], "1")
}

func TestMap_Nil(t *testing.T) {
got := Map[string, string](nil, strings.ToUpper)
testkit.AssertNil(t, got)
}

func TestMap_Empty(t *testing.T) {
got := Map([]string{}, strings.ToUpper)
testkit.AssertLen(t, got, 0)
}
