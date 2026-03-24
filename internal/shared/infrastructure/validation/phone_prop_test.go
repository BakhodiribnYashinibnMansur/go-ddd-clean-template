package validation

import (
	"strings"
	"testing"
	"testing/quick"
	"unicode"
)

func TestSanitizePhone_Property(t *testing.T) {
	// Property 1: Output must only contain digits and optionally a leading '+'
	propertyFormat := func(input string) bool {
		output := SanitizePhone(input)
		if output == "" {
			return true
		}

		runes := []rune(output)
		startIndex := 0
		if runes[0] == '+' {
			startIndex = 1
			// If it's just "+", is that valid sanitized output?
			// Logic: input "+" -> Prefix yes -> sb="+" -> loop empty -> return "+"
			// Yes.
		}

		for i := startIndex; i < len(runes); i++ {
			if !unicode.IsDigit(runes[i]) {
				t.Logf("Found non-digit in output: %c, output: %s", runes[i], output)
				return false
			}
		}
		return true
	}

	// Property 2: Idempotency. Sanitize(Sanitize(x)) == Sanitize(x)
	propertyIdempotency := func(input string) bool {
		once := SanitizePhone(input)
		twice := SanitizePhone(once)
		return once == twice
	}

	// Property 3: Subsequence. The output digits must be a subsequence of the input digits.
	// (Sanitize shouldn't introduce new digits or reorder them)
	propertySubsequence := func(input string) bool {
		output := SanitizePhone(input)

		// If output has +, skip it for digit comparison
		outDigits := output
		if strings.HasPrefix(output, "+") {
			outDigits = output[1:]
		}

		// Check if outDigits is a subsequence of input (using runes, not bytes)
		inputRunes := []rune(input)
		inputIdx := 0
		for _, outChar := range outDigits {
			found := false
			for i := inputIdx; i < len(inputRunes); i++ {
				if inputRunes[i] == outChar {
					inputIdx = i + 1
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	config := &quick.Config{
		MaxCount: 1000,
	}

	if err := quick.Check(propertyFormat, config); err != nil {
		t.Errorf("Format property failed: %v", err)
	}

	if err := quick.Check(propertyIdempotency, config); err != nil {
		t.Errorf("Idempotency property failed: %v", err)
	}

	if err := quick.Check(propertySubsequence, config); err != nil {
		t.Errorf("Subsequence property failed: %v", err)
	}
}
