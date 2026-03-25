package sanitize

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

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

func TestCompact_Strings(t *testing.T) {
	got := Compact([]string{"a", "", "b", "", "c"})
	testkit.AssertLen(t, got, 3)
	testkit.AssertEqual(t, got[0], "a")
	testkit.AssertEqual(t, got[1], "b")
	testkit.AssertEqual(t, got[2], "c")
}

func TestCompact_Ints(t *testing.T) {
	got := Compact([]int{0, 1, 0, 2, 3})
	testkit.AssertLen(t, got, 3)
	testkit.AssertEqual(t, got[0], 1)
}

func TestCompact_NoZeros(t *testing.T) {
	got := Compact([]string{"a", "b"})
	testkit.AssertLen(t, got, 2)
}

func TestCompact_AllZeros(t *testing.T) {
	got := Compact([]string{"", ""})
	testkit.AssertLen(t, got, 0)
}

func TestContains_Found(t *testing.T) {
	testkit.AssertTrue(t, Contains([]string{"a", "b", "c"}, "b"))
	testkit.AssertTrue(t, Contains([]int{1, 2, 3}, 3))
}

func TestContains_NotFound(t *testing.T) {
	testkit.AssertFalse(t, Contains([]string{"a", "b"}, "z"))
	testkit.AssertFalse(t, Contains([]int{1, 2}, 5))
}

func TestContains_Empty(t *testing.T) {
	testkit.AssertFalse(t, Contains([]string{}, "a"))
}
