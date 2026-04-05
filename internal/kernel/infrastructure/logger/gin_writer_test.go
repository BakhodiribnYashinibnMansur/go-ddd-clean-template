package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewColorfulWriter(t *testing.T) {
	var buf bytes.Buffer
	cw := NewColorfulWriter(&buf)
	if cw == nil {
		t.Fatal("expected non-nil ColorfulWriter")
	}
}

func TestColorfulWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	cw := NewColorfulWriter(&buf)

	input := "[GIN-debug] GET /api/v1/users --> handler.GetUsers (4 handlers)"
	n, err := cw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestColorfulWriter_GINDebugPrefix(t *testing.T) {
	var buf bytes.Buffer
	cw := NewColorfulWriter(&buf)

	_, _ = cw.Write([]byte("[GIN-debug] some message"))
	output := buf.String()

	if strings.Contains(output, "[GIN-debug]") {
		t.Error("expected [GIN-debug] to be replaced with colorized version")
	}
	if !strings.Contains(output, "GIN_DEBUG") {
		t.Error("expected output to contain 'GIN_DEBUG'")
	}
}

func TestColorfulWriter_GINPrefix(t *testing.T) {
	var buf bytes.Buffer
	cw := NewColorfulWriter(&buf)

	_, _ = cw.Write([]byte("[GIN] some message"))
	output := buf.String()

	if strings.Contains(output, "[GIN]") {
		t.Error("expected [GIN] to be replaced")
	}
	if !strings.Contains(output, "GIN_CORE") {
		t.Error("expected output to contain 'GIN_CORE'")
	}
}

func TestColorfulWriter_WarningPrefix(t *testing.T) {
	var buf bytes.Buffer
	cw := NewColorfulWriter(&buf)

	_, _ = cw.Write([]byte("[WARNING] something"))
	output := buf.String()

	if strings.Contains(output, "[WARNING]") {
		t.Error("expected [WARNING] to be replaced")
	}
	if !strings.Contains(output, "WARNING") {
		t.Error("expected output to contain 'WARNING'")
	}
}

func TestColorfulWriter_ErrorPrefix(t *testing.T) {
	var buf bytes.Buffer
	cw := NewColorfulWriter(&buf)

	_, _ = cw.Write([]byte("[ERROR] something"))
	output := buf.String()

	if strings.Contains(output, "[ERROR]") {
		t.Error("expected [ERROR] to be replaced")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("expected output to contain 'ERROR'")
	}
}

func TestColorizeHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			input := method + " /test"
			output := colorizeHTTPMethods(input)
			if output == input {
				t.Errorf("expected %q to be colorized, but got same string", method)
			}
		})
	}
}

func TestColorizePaths(t *testing.T) {
	input := "GET /api/v1/users/123"
	output := colorizePaths(input)
	if output == input {
		t.Error("expected paths to be colorized")
	}
	// The path should still be present in the output (with color codes)
	if !strings.Contains(output, "/api/v1/users/123") {
		t.Error("expected path to be present in output")
	}
}

func TestColorizeRouters(t *testing.T) {
	input := "--> mypackage.MyHandler"
	output := colorizeRouters(input)
	if output == input {
		t.Error("expected router handler to be colorized")
	}
}

func TestColorizeHandlers(t *testing.T) {
	input := "(4 handlers)"
	output := colorizeHandlers(input)
	if output == input {
		t.Error("expected handler info to be colorized")
	}
}

func TestColorizeGinOutput(t *testing.T) {
	input := "Line 1\nLine 2\nLine 3"
	output := ColorizeGinOutput(input)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if !strings.Contains(output, "Line 1") {
		t.Error("expected output to contain 'Line 1'")
	}
	if !strings.Contains(output, "Line 2") {
		t.Error("expected output to contain 'Line 2'")
	}
}

func TestColorizeGinOutput_EmptyInput(t *testing.T) {
	output := ColorizeGinOutput("")
	if output == "" {
		t.Error("expected non-empty output (box drawing chars)")
	}
}

func TestColorfulWriter_PlainText(t *testing.T) {
	var buf bytes.Buffer
	cw := NewColorfulWriter(&buf)

	input := "just plain text with no special markers"
	_, err := cw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "just plain text") {
		t.Error("expected plain text to pass through")
	}
}
