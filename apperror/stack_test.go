package apperror

import (
	"errors"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestWithStack_AppError(t *testing.T) {
	orig := NotFound("item not found")
	err := WithStack(orig)

	testkit.AssertTrue(t, HasStack(err))
	testkit.AssertTrue(t, len(err.Stack) > 0)
	testkit.AssertEqual(t, err.Code, CodeNotFound)
	testkit.AssertEqual(t, err.Message, "item not found")

	// Should contain this test function in the stack
	found := false
	for _, f := range err.Stack {
		if strings.Contains(f.Function, "TestWithStack_AppError") {
			found = true
			break
		}
	}
	testkit.AssertTrue(t, found)
}

func TestWithStack_GenericError(t *testing.T) {
	base := errors.New("something broke")
	err := WithStack(base)

	testkit.AssertTrue(t, HasStack(err))
	testkit.AssertEqual(t, err.Code, CodeInternal)
	testkit.AssertTrue(t, errors.Is(err, base))
}

func TestWithStack_Nil(t *testing.T) {
	testkit.AssertNil(t, WithStack(nil))
}

func TestWithStack_DoesNotMutateOriginal(t *testing.T) {
	orig := BadRequest("bad input")
	withStack := WithStack(orig)

	testkit.AssertFalse(t, HasStack(orig))
	testkit.AssertTrue(t, HasStack(withStack))
}

func TestHasStack_NoStack(t *testing.T) {
	testkit.AssertFalse(t, HasStack(NotFound("x")))
}

func TestHasStack_NonAppError(t *testing.T) {
	testkit.AssertFalse(t, HasStack(errors.New("plain")))
}

func TestFrame_String(t *testing.T) {
	f := Frame{Function: "main.handler", File: "/app/main.go", Line: 42}
	s := f.String()
	testkit.AssertContains(t, s, "main.handler")
	testkit.AssertContains(t, s, "/app/main.go:42")
}

func TestStackTrace_String(t *testing.T) {
	st := StackTrace{
		{Function: "pkg.A", File: "a.go", Line: 1},
		{Function: "pkg.B", File: "b.go", Line: 2},
	}
	s := st.String()
	testkit.AssertContains(t, s, "pkg.A")
	testkit.AssertContains(t, s, "pkg.B")
}

func TestStackTrace_Empty(t *testing.T) {
	var st StackTrace
	testkit.AssertEqual(t, st.String(), "")
}
